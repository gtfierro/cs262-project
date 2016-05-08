import time
from broker_connection import Config
from publisher import Publisher


def publish_loop(publisher):
    publish_val = 0
    while True:
        try:
            publisher.publish(publish_val)
            publish_val += 1
        except IOError:
            return
        time.sleep(1)

uuid = 'publisheruuid1'
config = Config('127.0.0.1', 4444, '127.0.0.1', 5505)
simple_publisher = Publisher(uuid, publish_loop, config)
simple_publisher.add_metadata({'Room': '410'})
simple_publisher.start()

time.sleep(60)
