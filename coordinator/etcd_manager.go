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

func (ec *EtcdConnection) GetCtx() context.Context {
	return context.Background()
	//ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	//return ctx
}

type EtcdManager interface {
	UpdateEntity(entity EtcdSerializable) error
	DeleteEntity(entity EtcdSerializable) error
	GetHighestKeyAtRev(prefix string, rev int64) (string, error)
	WriteToLog(idOrGeneral string, isSend bool, msg common.Sendable) error
	WatchLog(startKey string) string
	//CancelWatch()
	IterateOverAllEntities(entityType string, upToRev int64, processor func(EtcdSerializable)) error
	RegisterLogHandler(idOrGeneral string, handler LogHandler)
	UnregisterLogHandler(idOrGeneral string)
}

type EtcdManagerImpl struct {
	conn              *EtcdConnection
	leaderService     *LeaderService
	logHandlers       map[string]LogHandler
	cancelWatchChan   chan bool
	handlerLock       sync.RWMutex
	handlerCond       *sync.Cond
	highestKeyWritten string
	highKeyLock       sync.Mutex
	maxKeysPerRequest int64
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

func NewEtcdManager(etcdConn *EtcdConnection, leaderService *LeaderService, timeout time.Duration, maxKeysPerRequest int64) EtcdManager {
	em := new(EtcdManagerImpl)
	em.maxKeysPerRequest = maxKeysPerRequest
	em.conn = etcdConn
	em.leaderService = leaderService

	em.logHandlers = make(map[string]LogHandler)
	em.handlerCond = sync.NewCond(&em.handlerLock)
	em.cancelWatchChan = make(chan bool, 1)

	go em.handleLeaderLog()

	return em
}

// Doesn't return
func (em *EtcdManagerImpl) handleLeaderLog() {
	for {
		// Whenever a new leader is chosen, that new leader adds
		// an entry to the log indicating that messages after that are
		// from this replica
		<-em.leaderService.WaitForLeadership()
		em.WriteToLog(GeneralSuffix, false, &common.LeaderChangeMessage{})
		<-em.leaderService.WaitForNonleadership()
	}
}

func (em *EtcdManagerImpl) UpdateEntity(entity EtcdSerializable) error {
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
	_, err = em.conn.kv.Put(em.conn.GetCtx(), prefix+"/"+string(id), string(bytePacked))
	if err != nil {
		log.WithFields(log.Fields{
			"UUID": id, "error": err, "type": prefix,
		}).Error("Error while updating entity in etcd")
	}
	return err
}

func (em *EtcdManagerImpl) DeleteEntity(entity EtcdSerializable) error {
	if !em.leaderService.IsLeader() {
		return nil // No updates if you're not the leader
	}
	id, prefix := entity.GetIDType()
	_, err := em.conn.kv.Delete(em.conn.GetCtx(), prefix+"/"+string(id))
	if err != nil {
		log.WithFields(log.Fields{
			"UUID": id, "error": err, "type": prefix,
		}).Error("Error while deleting entity from etcd")
	}
	return err
}

func (em *EtcdManagerImpl) GetHighestKeyAtRev(prefix string, rev int64) (string, error) {
	opts := append(append(etcdc.WithLastKey(), etcdc.WithRev(rev)), etcdc.WithFromKey())
	resp, err := em.conn.kv.Get(em.conn.GetCtx(), prefix, opts...)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "revision": rev, "prefix": prefix,
		}).Error("Error while attempting to find highest key at revision ")
		return "", err
	}
	if len(resp.Kvs) == 0 {
		return "", nil
	}
	return string(resp.Kvs[0].Key), nil
}

func (em *EtcdManagerImpl) WatchFromKey(startKey string) etcdc.WatchChan {
	return em.conn.watcher.Watch(em.conn.GetCtx(), startKey, etcdc.WithFromKey())
}

func (em *EtcdManagerImpl) WriteToLog(idOrGeneral string, isSend bool, msg common.Sendable) error {
	var suffix string
	bytePacked, err := msg.Marshal()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "message": msg,
		}).Error("Error while marshalling message to write to log")
	}
	if isSend {
		suffix = idOrGeneral + SentSuffix
	} else {
		suffix = idOrGeneral + RcvdSuffix
	}
	newKey, err := em.newSequentialKV(LogPrefix, suffix, string(bytePacked))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "suffix": suffix, "message": msg,
		}).Error("Error while storing into log")
	}
	em.highKeyLock.Lock()
	if newKey > em.highestKeyWritten {
		em.highestKeyWritten = newKey
	}
	em.highKeyLock.Unlock()
	return err
}

// Does not return until cancellation - feeds messages into registered handlers
// starts watching AFTER max(startKey, highestKeyWritten)
func (em *EtcdManagerImpl) WatchLog(startKey string) (lastKey string) {
	// TODO also need to GC the log here: (also compaction!)
	//   every so often when reading, write to some node within the prefix
	//   what key you've read up to. also read the other replica's last read key.
	//  then you can delete everything up to the lower of the two
	var (
		handler func(common.Sendable, bool)
		found   bool
	)
	if startKey > em.highestKeyWritten {
		lastKey = startKey + "a"
	} else {
		lastKey = em.highestKeyWritten + "a"
	}
	watchChan := em.WatchFromKey(lastKey)
Loop:
	for {
		select {
		case <-em.cancelWatchChan:
			em.handlerLock.Lock()
			em.logHandlers = make(map[string]LogHandler)
			em.handlerLock.Unlock()
			return
		case watchResp := <-watchChan:
			if watchResp.Canceled {
				// Restart the watch
				watchChan = em.WatchFromKey(lastKey)
				continue Loop
			}
		EventLoop:
			for _, event := range watchResp.Events {
				if event.Type != mvccpb.PUT || !event.IsCreate() {
					log.WithFields(log.Fields{
						"eventType": event.Type, "key": event.Kv.Key,
					}).Warn("Non put+create event found in the general receive log!")
					continue EventLoop
				}
				msg, err := common.MessageFromBytes(event.Kv.Value)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err, "key": event.Kv.Key,
					}).Error("Error while unmarshalling bytes from general rcvd log at the given key")
					continue EventLoop
				}
				if newKey := string(event.Kv.Key); newKey > lastKey {
					lastKey = newKey + "a"
				}
				if _, ok := msg.(*common.LeaderChangeMessage); ok {
					em.handlerLock.RLock()
					for _, handler := range em.logHandlers {
						handler(msg, true)
					}
					em.handlerLock.RUnlock()
					return
				}
				fields := strings.Split(string(event.Kv.Key), "/")
				idOrGeneral := fields[len(fields)-2]
				isSent := fields[len(fields)-1] == "sent"
				em.handlerLock.RLock()
				for handler, found = em.logHandlers[idOrGeneral]; !found; {
					// Wait for proper handler to be available in case e.g. it hasn't registered yet
					em.handlerCond.Wait()
				}
				em.handlerLock.RUnlock()
				handler(msg, isSent)
			}
		}
	}
	return
}

//func (em *EtcdManagerImpl) CancelWatch() {
//	em.cancelWatchChan <- true
//}

func (em *EtcdManagerImpl) RegisterLogHandler(idOrGeneral string, handler LogHandler) {
	em.handlerLock.Lock()
	em.logHandlers[idOrGeneral] = handler
	em.handlerCond.Broadcast()
	em.handlerLock.Unlock()
}

func (em *EtcdManagerImpl) UnregisterLogHandler(idOrGeneral string) {
	em.handlerLock.Lock()
	delete(em.logHandlers, idOrGeneral)
	em.handlerLock.Unlock()
}

// Iterates over all of the entries that are currently in the system up to upToRev
// TODO this doesn't respect the maximum entity-at-once limit and just fetches ALL of them at once
func (em *EtcdManagerImpl) IterateOverAllEntities(entityType string, upToRev int64, processor func(EtcdSerializable)) error {
	var (
		entity EtcdSerializable
		resp   *etcdc.GetResponse
		err    error
	)

	//for {
	resp, err = em.conn.kv.Get(em.conn.GetCtx(), entityType+"/", etcdc.WithRev(upToRev),
		etcdc.WithPrefix(), etcdc.WithSerializable())
	if err != nil {
		return err
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
	}
	//if !resp.More { // No more keys to return
	//	break
	//}
	//}

	return nil
}

// newSequentialKV allocates a new sequential key <prefix>/nnnnn/suffix with a given
// value. Note: a bookkeeping node __<prefix> is also allocated.
// modified from https://github.com/coreos/etcd/blob/master/contrib/recipes/key.go
// to include a suffix as well
// NOTE: suffix MUST contain a slash e.g. "general/sent"
func (em *EtcdManagerImpl) newSequentialKV(prefix, suffix, val string) (string, error) {
	resp, err := em.conn.kv.Get(em.conn.GetCtx(), prefix, etcdc.WithLastKey()...)
	if err != nil {
		return "", err
	}

	// add 1 to last key, if any
	newSeqNum := 0
	if len(resp.Kvs) != 0 {
		fields := strings.Split(string(resp.Kvs[0].Key), "/")
		_, serr := fmt.Sscanf(fields[len(fields)-3], "%d", &newSeqNum)
		if serr != nil {
			return "", serr
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

	txn := em.conn.kv.Txn(em.conn.GetCtx())
	txnresp, err := txn.If(cmp).Then(reqPrefix, reqNewKey).Commit()
	if err != nil {
		return "", err
	}
	if !txnresp.Succeeded {
		return em.newSequentialKV(prefix, suffix, val) // retry
	}
	return newKey, nil
}
