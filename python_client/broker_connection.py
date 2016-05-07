import socket
import re
import time
import threading
from messages import MessageDecoder, BrokerAssignmentMessage, BrokerRequestMessage
import uuid as uuidlib


class Config:

    def __init__(self, broker_host, broker_port, coord_host, coord_port):
        self.broker_addr = (str(broker_host), int(broker_port))
        self.coord_addr = (str(coord_host), int(coord_port))


class BrokerConnection:

    def __init__(self, config, is_publisher, connect_callback, message_handler, uuid=None):
        self.uuid = uuid if uuid is not None else uuidlib.uuid4()
        self.uuid = str(self.uuid) # coerce to string
        self.is_publisher = is_publisher
        self.config = config
        self.connect_callback = connect_callback
        self.message_handler = message_handler
        self.brokerSock = None
        self.coordSock = None

        self.decoder = MessageDecoder()

    def start(self):
        self.connect_to_broker(self.config.broker_addr)
        listen_thread = threading.Thread(target=self.listen)
        listen_thread.start()
        try:
            self.connect_callback()
        except IOError as e:
            print "IOError encountered: " + e.message
            self.start()

    def listen(self):
        while True:
            buf = self.brokerSock.recv(1024**2)
            if not buf:
                print "Received an empty buffer from broker"
                self.start()
                return
            self.decoder.add_bytes(buf)
            while True:
                message = self.decoder.get_message()
                if message is None:
                    break  # inner loop
                elif self.message_handler is not None:
                    self.message_handler(message)

    # may throw IOError
    def send(self, message):
        buf = message.pack()
        self.brokerSock.sendall(buf)

    def connect_to_broker(self, broker_addr):
        try:
            self.brokerSock = socket.create_connection(broker_addr)
            print "Connected to broker at " + str(broker_addr)
            return
        except IOError, socket.timeout:
            print "Error connecting to broker: " + str(broker_addr)
            self.do_failover()

    def do_failover(self):
        while True:
            print "Attempting to connect to coordinator"
            try:
                self.coordSock = socket.create_connection(self.config.coord_addr)
                break
            except IOError, socket.timeout:
                print "Error connecting to coordinator"
                time.sleep(1)
        listen_thread = threading.Thread(target=self.listen_coord)
        listen_thread.start()
        local_addr_str = self.config.broker_addr[0] + ':' + str(self.config.broker_addr[1])
        broker_request_message = BrokerRequestMessage(local_addr_str, self.is_publisher, self.uuid)
        while True:
            try:
                print "Sending request message to coordinator"
                self.coordSock.sendall(broker_request_message.pack())
            except IOError:
                time.sleep(1)
                break
            time.sleep(1)

    def listen_coord(self):
        print "Listening for response from coordinator"
        buf = self.coordSock.recv(4096)
        if not buf:
            print "Received an empty buffer from coordinator"
            self.do_failover()
            return
        self.decoder.add_bytes(buf)
        msg = self.decoder.get_message()
        if not isinstance(msg, BrokerAssignmentMessage):
            print "Received non-BrokerAssignmentMessage from coordinator!"
            self.do_failover()
            return
        print "Got an assignment from the coordinator"
        broker_addr = msg.client_broker_addr
        match = re.match('(.+?):(\d+)', broker_addr)
        if match is None:
            raise Exception("Error parsing new broker address: " + broker_addr)
        self.coordSock.close()
        self.connect_to_broker((match.group(1), int(match.group(2))))
        return
