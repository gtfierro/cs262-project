import msgpack


class MessageDecoder:

    def __init__(self):
        self.unpacker = msgpack.Unpacker()

    def add_bytes(self, byte_buf):
        self.unpacker.feed(byte_buf)

    # Returns None if no messages are ready
    def get_message(self):
        msg_type = bytearray(self.unpacker.read_bytes(1))
        if len(msg_type) == 0:
            return None
        msg_type = int(msg_type[0])
        for val in self.unpacker:
            return MessageDecoder.message_from_map(msg_type, val)

    @staticmethod
    def message_from_map(msg_type, msg_map):
        if msg_type == PublishMessage.MSG_NUM:
            return PublishMessage.from_msgpack_map(msg_map)
        elif msg_type == SubscriptionDiffMessage.MSG_NUM:
            return SubscriptionDiffMessage.from_msgpack_map(msg_map)
        elif msg_type == BrokerAssignmentMessage.MSG_NUM:
            return BrokerAssignmentMessage.from_msgpack_map(msg_map)
        else:
            raise Exception('Unknown message type received!')


class PublishMessage:

    MSG_NUM = 0

    def __init__(self, uuid, value, metadata=None):
        self.uuid = uuid
        self.metadata = metadata if metadata is not None else {}
        self.value = value

    def pack(self):
        msgbytes = msgpack.packb({"UUID": self.uuid, "Metadata": self.metadata, "Value": self.value})
        return bytearray([self.MSG_NUM]) + msgbytes

    @staticmethod
    def from_msgpack_map(m):
        return PublishMessage(m['UUID'], m['Value'], m['Metadata'])


class QueryMessage:

    MSG_NUM = 1

    def __init__(self, uuid, query):
        self.uuid = uuid
        self.query = query

    def pack(self):
        return bytearray([self.MSG_NUM]) + msgpack.packb({"Query": self.query, "UUID": self.uuid})


class SubscriptionDiffMessage:

    MSG_NUM = 2

    def __init__(self, new_publishers, del_publishers):
        self.new_publishers = new_publishers
        self.del_publishers = del_publishers

    @staticmethod
    def from_msgpack_map(m):
        return SubscriptionDiffMessage(m['New'], m['Del'])


class BrokerRequestMessage:

    MSG_NUM = 3

    def __init__(self, local_broker_addr, is_publisher, uuid):
        self.local_broker_addr = local_broker_addr
        self.is_publisher = is_publisher
        self.uuid = uuid

    def pack(self):
        packer = msgpack.Packer(autoreset=False)
        packer.pack(self.MSG_NUM)
        packer.pack({"LocalBrokerAddr": self.local_broker_addr,
                     "IsPublisher": self.is_publisher, "UUID": self.uuid})
        return packer.bytes()


class BrokerAssignmentMessage:

    MSG_NUM = 8

    def __init__(self, broker_id, client_broker_addr):
        self.broker_id = broker_id
        self.client_broker_addr = client_broker_addr

    @staticmethod
    def from_msgpack_map(m):
        return BrokerAssignmentMessage(m['BrokerID'], m['ClientBrokerAddr'])
