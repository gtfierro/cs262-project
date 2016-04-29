package main

import (
	log "github.com/Sirupsen/logrus"
	etcdc "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/contrib/recipes"
	"github.com/gtfierro/cs262-project/common"
	"golang.org/x/net/context"
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
	maxKeysPerRequest int
}

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

func (em *EtcdManager) WatchFromKeyAtRevision(prefix string, startKey string) chan etcdc.WatchResponse {
	return em.conn.watcher.Watch(em.context, prefix+startKey, etcdc.WithFromKey())
}

func (em *EtcdManager) WriteToLog(prefix string, msg common.Sendable) error {
	bytePacked, err := msg.Marshal()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "message": *msg,
		}).Error("Error while marshalling message to write to log")
	}
	resp, err := recipe.NewSequentialKV(em.conn.kv, prefix, string(bytePacked))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err, "prefix": prefix, "message": *msg, "key": resp.Key(),
		}).Error("Error while storing into log")
	}
	return err
}

// Iterates over all of the entries that are currently in the system as of the time of
// the first call to etcd. Then creates and returns a watchChannel which will contain all
// entries added after that point, which should be processed externally. This channel should
// be closed when processing is complete . Also returns the revision number at which everything
// was processed; the watchChannel watches for revisions after this.
func (em *EtcdManager) IterateOverAllEntities(entityType string,
	processor func(EtcdSerializable)) (chan *etcdc.WatchResponse, int64, error) {
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

	watchChan := em.conn.watcher.Watch(em.context, entityType+"/", etcdc.WithPrefix(), etcdc.WithRev(startingRevision))
	return watchChan, startingRevision, nil
}
