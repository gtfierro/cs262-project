Single Node
===========

Here we implement a single-node server and discuss some of the design decisions (which
may be revisited)

## Messages and Transport

There are 3 types of messages: subscribe, publish and subscribediff.

To simplify cross-language support (for clients) and for simplicity in
parsing/packing, we are using [MessagePack/MsgPack](http://msgpack.org/) for
binary serialization. MsgPack is crazy fast in Go because it encodes types in
the message (thus avoiding runtime reflection) and the serialized messages are
smaller than JSON even though it can encode the same structures.

#### Subscribe
This is the message sent by a client when a connection to the
server is initiated and takes the form of a SQL-like predicate (described
below). Publishers do not use or see this message. 

Up for discussion: how powerful are these subscriptions? Can clients only subscribe to
published messages, or can clients also subscribe to some more detailed query, such
as "list of rooms containing thermostats", which involves a more expensive resolution
of a query?

Representation: This is a MsgPack string.

#### Publish

This is the message sent by producers and also the message received by clients.
Each `publish` message must carry Metadata (a description of the producer) and
Data (the actual message contents, likely a sensor reading). Metadata is a set
of key-value pairs.  Data (for now) is simply a float indicating a sensor
reading. 

Representation: This is a MsgPack array with length 3: 
1. unique identifier for the publisher (discussed below)
2. a MsgPack map containing all key-value pairs in metadata
3. a float containing the sensor reading. 

A publisher does not need to (and in fact *should not*) send all of its
metadata in each message. Metadata is stored at the broker on a per-producer
basis: when a new key-value pair arrives for a publisher, it performs an
"upsert" (insert if not exists, else update) for that producer. If the value is
a nil value (MsgPack supports nil values), then the key is deleted for that
producer.

For unique identifiers, we use UUIDv4 to remove the need to coordinate
publisher IDs.

#### SubscribeDiff 

This is the message received by clients when the set of publishers they are
subscribed to changes because of a change in that publisher's metadata.

Representation: This is a MsgPack map with 2 keys, "New" and "Del", which each
contain an array of UUIDs. The "New" list contains the list of publisher IDs
that now qualify for the client's subscription and the "Del" list contains the
set of publisher IDs who have changed their metadata so they no longer match
the subscription query.
