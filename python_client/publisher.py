from broker_connection import BrokerConnection
from messages import PublishMessage
import threading


class Publisher:

    # publish_loop should be a loop which will start whenever a connection
    # to the broker is established; it should exit whenever an error is encountered
    # and will be restarted once a connection is reestablished
    # receives this publisher as an argument
    def __init__(self, uuid, publish_loop, config):
        self.uuid = uuid
        self.metadata = {}
        self.dirty_md = False
        self.md_lock = threading.Lock()
        self.publish_loop = publish_loop

        self.broker_conn = BrokerConnection(config, self.connect_callback, None, uuid)

    def start(self):
        self.broker_conn.start()

    def connect_callback(self):
        publish_loop_thread = threading.Thread(target=self.publish_loop_wrap)
        publish_loop_thread.start()

    def publish_loop_wrap(self):
        self.publish_loop(self)

    def publish(self, value):
        metadata = {}
        self.md_lock.acquire()
        if self.dirty_md:
            metadata = self.metadata
            self.dirty_md = False
        self.md_lock.release()

        msg = PublishMessage(self.uuid, value, metadata)
        self.broker_conn.send(msg)

    def add_metadata(self, new_metadata):
        self.md_lock.acquire()
        for k, v in new_metadata.iteritems():
            self.metadata[k] = v
        self.dirty_md = True
        self.md_lock.release()
