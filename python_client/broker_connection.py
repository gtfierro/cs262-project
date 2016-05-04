import socket
import msgpack
import threading
import messages
import uuid as uuidlib


class Config:

    def __init__(self, broker_host, broker_port, coord_host, coord_port):
        self.broker_addr = (str(broker_host), int(broker_port))
        self.coord_addr = (str(coord_host), int(coord_port))


class BrokerConnection:

    def __init__(self, config, connect_callback, message_handler, uuid=None):
        self.uuid = uuid if uuid is not None else uuidlib.uuid4()
        self.uuid = str(self.uuid) # coerce to string
        self.config = config
        self.connect_callback = connect_callback
        self.message_handler = message_handler

        self.decoder = messages.MessageDecoder()

        # self.brokerSock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.brokerSock = socket.create_connection(self.config.broker_addr)

    def start(self):
        listen_thread = threading.Thread(target=self.listen)
        listen_thread.start()
        try:
            self.connect_callback()
        except IOError as e:
            print "IOError encountered: " + e.message
            # TODO Error handling

    def listen(self):
        while True:
            buf = self.brokerSock.recv(1024**2)
            if not buf:
                break  # TODO error handling
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
