//go:generate msgp
package main

import (
	"container/list"
	"fmt"
	"github.com/gtfierro/cs262-project/common"
)

const BrokerEntity = "broker"
const ClientEntity = "client"
const PublisherEntity = "publisher"

const BrokerLog = "log/broker/"
const GeneralLog = "log/general/"

func GetBrokerLogPrefix(brokerID common.UUID) string {
	return fmt.Sprintf("%v%v/", BrokerLog, brokerID)
}

type SerializableBroker struct {
	common.BrokerInfo
	Alive bool
}

func (b *Broker) ToSerializable() *SerializableBroker {
	sb := new(SerializableBroker)
	sb.BrokerInfo = b.BrokerInfo
	sb.Alive = b.IsAlive()
	return sb
}

func (sb *SerializableBroker) GetIDType() (common.UUID, string) {
	return sb.BrokerID, BrokerEntity
}

type Publisher struct {
	PublisherID     common.UUID
	CurrentBrokerID common.UUID
	HomeBrokerID    common.UUID
	Metadata        map[string]interface{}
}

func (pub *Publisher) GetIDType() (common.UUID, string) {
	return pub.PublisherID, PublisherEntity
}

type Client struct {
	ClientID           common.UUID
	CurrentBrokerID    common.UUID
	HomeBrokerID       common.UUID
	QueryString        string
	subscriberListElem *list.Element   `msg:"-"`
	query              *ForwardedQuery `msg:"-"`
}

func (c *Client) GetIDType() (common.UUID, string) {
	return c.ClientID, ClientEntity
}
