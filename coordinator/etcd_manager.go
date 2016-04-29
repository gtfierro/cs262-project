package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	etcdc "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/gtfierro/cs262-project/common"
	"golang.org/x/net/context"
	"strings"
	"sync"
	"time"
)

type EtcdConnection struct {
	client  *etcdc.Client
	kv      etcdc.KV
	watcher etcdc.Watcher
}

type EtcdManager struct {
	conn              *EtcdConnection
	leaderService     *LeaderService
	context           context.Context
	logHandlers       map[string]LogHandler
	cancelWatchChan   chan bool
	handlerLock       sync.RWMutex
	handlerCond       *sync.Cond
	maxKeysPerRequest int
}

type LogHandler func(common.Sendable, bool)

type EtcdSerializable interface {
	UnmarshalMsg([]byte) ([]byte, error)
	MarshalMsg([]byte) ([]byte, error)
	GetIDType() (common.UUID, string)
}

func NewEtcdConnection(endpoints []string) *EtcdConnection {
	var err error
	ec := new(EtcdConnection)
	etcdCfg := etcdc.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	}
	ec.client, err = etcdc.New(etcdCfg)
	if err != nil {
		log.WithField("endpoints", endpoints).Fatal("Unable to contact etcd server!")
	}
	ec.kv = etcdc.NewKV(ec.client)
	ec.watcher = etcdc.Watcher(ec.client)
	//defer etcdm.client.Close() TODO
	return ec
}

func NewEtcdManager(etcdConn *EtcdConnection, leaderService *LeaderService, timeout time.Duration, maxKeysPerRequest int) *EtcdManager {
	em := new(EtcdManager)
	em.maxKeysPerRequest = maxKeysPerRequest
	em.conn = etcdConn
	em.leaderService = leaderService
	em.context, _ = context.WithTimeout(context.Background(), timeout) // Do we need the cancel func?

	em.logHandlers = make(map[string]LogHandler)
	em.handlerCond = sync.NewCond(&em.handlerLock)
	em.cancelWatchChan = make(chan bool, 1)
	return em
}

func (em *EtcdManager) UpdateEntity(entity EtcdSerializable) error {
	if !em.leaderService.IsLeader() {
		return nil // No updates if you're not the leader
	}
	id, prefix := entity.GetIDType()
	bytePacked, err := entity.MarshalMsg([]byte{})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "UUID": id, "type": prefix,
		}).Error("Error while serializing entity etcd")
		return err
	}
	_, err = em.conn.kv.Put(em.context, prefix+"/"+string(id), string(bytePacked))
	if err != nil {
		log.WithFields(log.Fields{
			"UUID": id, "error": err, "type": prefix,
		}).Error("Error while updating entity in etcd")
	}
	return err
}

func (em *EtcdManager) DeleteEntity(entity EtcdSerializable) error {
	if !em.leaderService.IsLeader() {
		return nil // No updates if you're not the leader
	}
	id, prefix := entity.GetIDType()
	_, err := em.conn.kv.Delete(em.context, prefix+"/"+string(id))
	if err != nil {
		log.WithFields(log.Fields{
			"UUID": id, "error": err, "type": prefix,
		}).Error("Error while deleting entity from etcd")
	}
	return err
}

func (em *EtcdManager) GetHighestKeyAtRev(prefix string, rev int64) (string, error) {
	resp, err := em.conn.kv.Get(em.context, prefix, etcdc.WithRev(rev), etcdc.WithFromKey(), etcdc.WithLastKey())
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "revision": rev, "prefix": prefix,
		}).Error("Error while attempting to find highest key at revision ")
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", nil
	}
	return resp.Kvs[0].Key, nil
}

func (em *EtcdManager) WatchFromKey(prefix string, startKey string) chan etcdc.WatchResponse {
	return em.conn.watcher.Watch(em.context, prefix+"/"+startKey, etcdc.WithFromKey())
}

func (em *EtcdManager) WriteToLog(idOrGeneral string, isSend bool, msg common.Sendable) error {
	var suffix string
	bytePacked, err := msg.Marshal()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "message": *msg,
		}).Error("Error while marshalling message to write to log")
	}
	if isSend {
		suffix = idOrGeneral + SentSuffix
	} else {
		suffix = idOrGeneral + RcvdSuffix
	}
	err = em.newSequentialKV(LogPrefix, suffix, string(bytePacked))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "suffix": suffix, "message": *msg,
		}).Error("Error while storing into log")
	}
	return err
}

// TODO need something like this to know where to resume
//func (rcc *ReplicaCommConn) nudgeHighestRev(rev int64) {
//	rcc.revLock.Lock()
//	defer rcc.revLock.Unlock()
//	if rev > rcc.highestRev {
//		rcc.highestRev = rev
//	}
//}

// Does not return until cancellation - feeds messages into registered handlers
func (em *EtcdManager) WatchLog(startKey string) {
	// TODO should have logic here that if the watchChan closes, we reopen it
	// TODO also need to GC the log here:
	//   every so often when reading, write to some node within the prefix
	//   what key you've read up to. also read the other replica's last read key.
	//  then you can delete everything up to the lower of the two

	//watchStartKey, err := rcc.etcdManager.GetHighestKeyAtRev(idOrGeneral +"rcvd/", startingRev)
	//if err != nil {
	//	return
	//}
	var (
		handler func(common.Sendable, bool)
		found   bool
	)
	watchChan := em.WatchFromKey(LogPrefix, startKey)
	for {
		select {
		case watchResp := <-watchChan:
			for _, event := range watchResp.Events {
				if event.Type != mvccpb.PUT || !event.IsCreate() {
					log.WithFields(log.Fields{
						"eventType": event.Type, "key": event.Kv.Key,
					}).Warn("Non put+create event found in the general receive log!")
					continue
				}
				msg, err := common.MessageFromBytes(event.Kv.Value)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err, "key": event.Kv.Key,
					}).Error("Error while unmarshalling bytes from general rcvd log at the given key")
					continue
				}
				fields := strings.Split(string(event.Kv.Key), "/")
				idOrGeneral := fields[len(fields)-2]
				isSent := fields[len(fields)-1] == "sent"
				em.handlerLock.RLock()
				for handler, found = em.logHandlers[idOrGeneral]; ; !found {
					// Wait for proper handler to be available in case e.g. it hasn't registered yet
					em.handlerCond.Wait()
				}
				em.handlerLock.RUnlock()
				handler(msg, isSent)
			}
		case <-em.cancelWatchChan:
			em.handlerLock.Lock()
			em.logHandlers = make(map[string]LogHandler)
			em.handlerLock.Unlock()
			return
		}
	}
}

func (em *EtcdManager) CancelWatch() {
	em.cancelWatchChan <- true
}

func (em *EtcdManager) RegisterLogHandler(idOrGeneral string, handler LogHandler) {
	em.handlerLock.Lock()
	em.logHandlers[idOrGeneral] = handler
	em.handlerCond.Broadcast()
	em.handlerLock.Unlock()
}

func (em *EtcdManager) UnregisterLogHandler(idOrGeneral string) {
	em.handlerLock.Lock()
	delete(em.logHandlers, idOrGeneral)
	em.handlerLock.Unlock()
}

// Iterates over all of the entries that are currently in the system as of the time of
// the first call to etcd. Then creates and returns a watchChannel which will contain all
// entries added after that point, which should be processed externally. This channel should
// be closed when processing is complete . Also returns the revision number at which everything
// was processed; the watchChannel watches for revisions after this.
func (em *EtcdManager) IterateOverAllEntities(entityType string,
	processor func(EtcdSerializable)) (int64, error) {
	var (
		entity EtcdSerializable
		resp   etcdc.GetResponse
		err    error
	)

	lastKey := entityType + "/"
	var startingRevision int64 = -1
	for {
		if startingRevision < 0 {
			resp, err = em.conn.kv.Get(em.context, lastKey,
				etcdc.WithLimit(em.maxKeysPerRequest), etcdc.WithFromKey(), etcdc.WithSerializable())
			startingRevision = resp.Header.Revision
		} else {
			resp, err = em.conn.kv.Get(em.context, lastKey, etcdc.WithRev(startingRevision),
				etcdc.WithLimit(em.maxKeysPerRequest), etcdc.WithFromKey(), etcdc.WithSerializable())
		}
		if err != nil {
			return nil, -1, err
		}
		for _, kv := range resp.Kvs {
			switch entityType {
			case BrokerEntity:
				entity = new(SerializableBroker)
			case ClientEntity:
				entity = new(Client)
			case PublisherEntity:
				entity = new(Publisher)
			default:
				log.WithField("entityType", entityType).Fatal("Invalid entity type passed to the EtcdManager")
			}
			entity.UnmarshalMsg(kv.Value)
			// TODO do we want this in the hopes of getting better parallelism?
			go func(ent EtcdSerializable) {
				processor(ent)
			}(entity)
			if string(kv.Key) > lastKey {
				lastKey = string(kv.Key)
			}
		}
		if !resp.More { // No more keys to return
			break
		}
	}

	return startingRevision, nil
}

// newSequentialKV allocates a new sequential key <prefix>/nnnnn/suffix with a given
// value. Note: a bookkeeping node __<prefix> is also allocated.
// modified from https://github.com/coreos/etcd/blob/master/contrib/recipes/key.go
// to include a suffix as well
// NOTE: suffix MUST contain a slash e.g. "general/sent"
func (em *EtcdManager) newSequentialKV(prefix, suffix, val string) error {
	resp, err := em.conn.kv.Get(em.context, prefix, etcdc.WithLastKey()...)
	if err != nil {
		return nil, err
	}

	// add 1 to last key, if any
	newSeqNum := 0
	if len(resp.Kvs) != 0 {
		fields := strings.Split(string(resp.Kvs[0].Key), "/")
		_, serr := fmt.Sscanf(fields[len(fields)-3], "%d", &newSeqNum)
		if serr != nil {
			return nil, serr
		}
		newSeqNum++
	}
	newKey := fmt.Sprintf("%s/%016d/%s", prefix, newSeqNum, suffix)

	// base prefix key must be current (i.e., <=) with the server update;
	// the base key is important to avoid the following:
	// N1: LastKey() == 1, start txn.
	// N2: New Key 2, New Key 3, Delete Key 2
	// N1: txn succeeds allocating key 2 when it shouldn't
	baseKey := "__" + prefix

	// current revision might contain modification so +1
	cmp := etcdc.Compare(etcdc.ModRevision(baseKey), "<", resp.Header.Revision+1)
	reqPrefix := etcdc.OpPut(baseKey, "")
	reqNewKey := etcdc.OpPut(newKey, val)

	txn := em.conn.kv.Txn(em.context)
	txnresp, err := txn.If(cmp).Then(reqPrefix, reqNewKey).Commit()
	if err != nil {
		return err
	}
	if !txnresp.Succeeded {
		return em.newSequentialKV(prefix, suffix, val) // retry
	}
	return nil
}
