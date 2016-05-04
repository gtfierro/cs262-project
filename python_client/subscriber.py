from broker_connection import BrokerConnection
from messages import QueryMessage, SubscriptionDiffMessage, PublishMessage


class Subscriber:

    def __init__(self, uuid, query, config):
        self.config = config
        self.uuid = uuid
        self.publish_handler = None
        self.diff_handler = None
        self.query = query

        self.broker_conn = BrokerConnection(config, self.subscribe, self.message_handle, uuid)

    def start(self):
        self.broker_conn.start()

    def subscribe(self):
        msg = QueryMessage(self.uuid, self.query)
        self.broker_conn.send(msg)

    def message_handle(self, message):
        if isinstance(message, PublishMessage):
            if self.publish_handler is not None:
                self.publish_handler(message)
        elif isinstance(message, SubscriptionDiffMessage):
            if self.diff_handler is not None:
                self.diff_handler(message)
        else:
            raise Exception('Received an unknown message! Type: ' + str(type(message)))

    def attach_publish_handler(self, publish_handler):
        self.publish_handler = publish_handler

    def attach_diff_handler(self, diff_handler):
        self.diff_handler = diff_handler
