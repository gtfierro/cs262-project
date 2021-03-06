Motivation
==========

## Query-Based Syndication

This section discusses the motivation for syndication based on *queries*.

The problem to solve is how to conduct discovery and transport in an IoT
setting, characterized by lots of compute/memory-constrained embedded devices
distributed over a large area, that are richly described so that they may be
discovered by applications or other devices.

### Why Do We Need Descriptions?

Why are descriptions an essential part of this? The simple answer is that names
alone are not sufficient; this is why we have services like DNS.  In the IoT,
devices may have names, but at the scale of hundreds or thousands or millions
of devices, it is simply intractable for some application to find all
devices/services/etc it needs to operate by enumerating the list of all
possible names. There is the need for some mechanism that can return to a
petitioner the list of devices/endpoints that match some query. This query may
be as simple as asking for some device that fulfills some interface (the
approach taken by distributed object protocols such as CORBA and DCOM), or may
be more advanced such as asking for the list of devices with one or more of a
set of capabilities (e.g. temperature OR humidity sensors) within some set of
locations.

The vast majority of current discovery mechanisms do not allow applications to
discover services from their *context*. Instead, we get methods that are
insufficient for following reasons:

* the discovery mechanism only confirms the existence of some device but not a
  list of its services and characteristics, requiring an external lookup (UPnP,
  DNS-SD)
* restricted to a particular network type (Zigbee)
* only says what a device provides, but not where it is or how its placed in an
  environment (UPnP, DNS-Sd, Bonjour/Zeroconf, DCOM, CORBA)

A more fundamental problem is that, like DNS, these discovery mechanisms establish
one-time bindings. As devices more, or as they are installed or decomissioned, there
should exist some mechanism to inform interested clients that their binding may not
be valid. DNS-like timeuts are insufficient here because:

* the time between metadata/context changes may vary wildly device-to-device or
  even by a single device
* if a device changes before its timeout has expired and clients are not made
  aware, that device could be sending incorrect/irrelevant data to a client,
  and/or not sending data to clients that *do* need it

### What Do Descriptions Look Like?

No "universal" description scheme has emerged in the context of IoT, but here
are some representative examples of what is being used today:
* The Semantic Web: queries using Web Ontology Language (OWL)
    * Semantic Sensor Web: annotates sensor data with spatial/temporal/thematic metadata
    * Resource Descsription Framework: a method for conceptual description of web resources
    * Note: there are a TON of Semantic Web ontologies
* Project Haystack: semantic data model for taxonomies of building structure and equipment
* SensorML: models and XML encoding for describing sensor parameters and observations

Other more building-related ones:
* Green Building XML (GBXML)
* Industry Foundation Classes (IFC)
* OPC-UA (Open platform communications unified architecture)
* Open Building Information Exchange (OBIX)

We can also easily imagine non-standards-based description schemes using
relational tables or "bag of tags" or key-value pairs approaches. 

### What about Topics?

Topic-based pubsub methods are the most prevalent flavor of publish-subscribe models
that exist today: Kafka, MQTT, AMQP.... However, is the topic structure sufficient
for the above description schemes? There are two options:

1. saving producer metadata in topic paths
2. implementing a "discovery" overlay for topic-based pubsub

The main intuition here is that the space of descriptions that could apply to a
single stream is large enough that you need a structured approach to
representing and querying it. For our use case, this is some database that can
support "rich queries" (most likely relational).  In the simplest cases, where
topics and names are known a priori, the topic approach is certainly
sufficient, but in the IoT, we predict a need for more expressive descriptions.

#### Saving Metadata in Topics

Topic matching methods
- strict name matching (Kafka)
- hierarchical/wildcard matching (prefix, suffix) (MQTT)
- full regex (Kafka)

It is certainly *possible* to store an arbitrary structure inside what is
essentially a flat char array, but this requires strict rules on how exactly
the arbitrary descriptive structure should be packed, but matching on this is
likely to be slow because descriptions may be arbitrarily large and complex.
Strict name matching is pretty much out here, but this is possible with wildcard
or regex matching.

The other option is to introduce an external service that stores the publisher
descriptions and serves queries that resolve to the specific producer topics
that are then subscribed to. There are several issues here: firstly, how does a
producer change or update its discription/metadata? Would the producer connect
to an external service, which would require the producer to wait for the change
to be "confirmed" before it could start sending data that occured "after" the
change? 

Or would the producer send the description inside its message, and then the
broker handles this reevaluation? This is essentially the solution we're
proposing.
