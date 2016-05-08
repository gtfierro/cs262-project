import msgpack
import sys
import socket
import uuid as uuidlib

class Client:
    def __init__(self, host, port, uuid=None):
        self.uuid = uuid if uuid is not None else uuidlib.uuid4()
        self.uuid = str(self.uuid) # coerce to string

        self.host = str(host)
        self.port = int(port)

        self.metadata = {}
        self._dirty_metadata = {}

        self.s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.s.connect((self.host, self.port))

    def subscribe(self, query):
        self.s.send(msgpack.packb(query))

    def add_metadata(self, d):
        strd = {str(k): str(v) for k,v in d.items()}
        self.metadata.update(strd)
        self._dirty_metadata = strd

    def publish(self, value):
        message = [self.uuid, self._dirty_metadata, value]
        print map(hex, map(ord, msgpack.packb(message)))
        self.s.send(msgpack.packb(message))
        self._dirty_metadata = {}

if __name__ == '__main__':
    c = Client("localhost", "4444")
    #c.add_metadata({"Room": "410", "Building": "Soda", "Device": "Temperature Sensor"})

    #import time
    #i = 0
    #while True:
    #    i += 1
    #    c.publish(i)
    #    time.sleep(5)
    c.subscribe(sys.argv[1])
