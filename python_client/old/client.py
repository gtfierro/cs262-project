import msgpack
import sys
import socket
import uuid as uuidlib
from io import BytesIO, BufferedRWPair

uuid = "295b3cd0-0bf8-11e6-81fc-0cc47a0f7eea"

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
        self.s.send(chr(0x01)+msgpack.packb({"UUID": uuid, "Query": query}))
        self.unpacker = msgpack.Unpacker()
        while 1:
            data = self.s.recv(1024)
            if not data:
                continue
            self.unpacker.feed(data)
            for m in self.unpacker:
                print m

    def add_metadata(self, d):
        strd = {str(k): str(v) for k,v in d.items()}
        self.metadata.update(strd)
        self._dirty_metadata = strd

    def publish(self, value):
        message = [self.uuid, self._dirty_metadata, value]
        print map(hex, map(ord, msgpack.packb(message)))
        self.s.send(chr(0x00)+msgpack.packb(message))
        self._dirty_metadata = {}

if __name__ == '__main__':
    host = sys.argv[1] if len(sys.argv) > 1 else "localhost"
    c = Client(host, "4444")
    #c.add_metadata({"Room": "410", "Building": "Soda", "Device": "Temperature Sensor"})

    #import time
    #i = 0
    #while True:
    #    i += 1
    #    c.publish(i)
    #    time.sleep(5)
    c.subscribe("Room = '410'")
