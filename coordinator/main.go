package main

import (
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/gtfierro/cs262-project/common"
	"github.com/pkg/profile"
	"golang.org/x/net/context"
	"os"
	"runtime/trace"
	"time"
)

// config flags
var configfile = flag.String("c", "config.ini", "Path to configuration file")

func init() {
	// set up logging
	log.SetOutput(os.Stderr)
}

func main() {
	flag.Parse()
	// configure logging instance
	config, configLogmsg := common.LoadConfig(*configfile)
	common.SetupLogging(config)
	log.Info(configLogmsg)

	if config.Debug.Enable {
		var p interface {
			Stop()
		}
		switch config.Debug.ProfileType {
		case "cpu", "profile":
			p = profile.Start(profile.CPUProfile, profile.ProfilePath("."))
		case "block":
			p = profile.Start(profile.BlockProfile, profile.ProfilePath("."))
		case "mem":
			p = profile.Start(profile.MemProfile, profile.ProfilePath("."))
		case "trace":
			f, err := os.Create("trace.out")
			if err != nil {
				log.Fatal(err)
			}
			trace.Start(f)
		}
		time.AfterFunc(time.Duration(config.Debug.ProfileLength)*time.Second, func() {
			if p != nil {
				p.Stop()
			}
			if config.Debug.ProfileType == "trace" {
				trace.Stop()
			}

			os.Exit(0)
		})
	}

	newval := "NEWVAL"
	key := "randomKey7"
	conn := NewEtcdConnection([]string{"127.0.0.1:2379"})
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cmp := clientv3.Compare(clientv3.Version(key), "=", 0)

	leaseResp, err := conn.client.Grant(ctx, 5)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	txn := conn.kv.Txn(ctx)
	putKeyOp := clientv3.OpPut(key, newval, clientv3.WithLease(leaseResp.ID))
	getKeyOp := clientv3.OpGet(key)
	txnResp, err := txn.If(cmp).Then(putKeyOp).Else(getKeyOp).Commit()
	if err != nil {
		println("err")
		return
	}
	fmt.Printf("txn was successful: %v\n", txnResp.Succeeded)
	if !txnResp.Succeeded {
		// if it exists, keep the lease alive
		var leaseID int64
		for _, resp := range txnResp.Responses {
			if r, ok := resp.Response.(*etcdserverpb.ResponseUnion_ResponseRange); ok {
				leaseID = r.ResponseRange.Kvs[0].Lease
			}
		}
		conn.client.KeepAliveOnce(ctx, clientv3.LeaseID(leaseID))
	}

	time.Sleep(6 * time.Second)                               // let lease expire
	lkar, err := conn.client.KeepAliveOnce(ctx, leaseResp.ID) // attempt to keep alive
	if err == rpctypes.ErrLeaseNotFound {
		fmt.Printf("lease not found!\n")
	} else if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	fmt.Printf("TTL: %v\n", lkar.TTL)

	//server := NewServer(config)
	//go server.handleLeadership()
	//go server.handleBrokerMessages()
	//go server.monitorLog()
	//go server.monitorGeneralConnections()
	//server.listenAndDispatch()
}
