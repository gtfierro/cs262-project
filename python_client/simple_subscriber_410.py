import time
from subscriber import Subscriber
from broker_connection import Config


def diff_handler(msg):
    print "New: " + str(msg.new_publishers)
    print "Del: " + str(msg.del_publishers)


def publish_handler(msg):
    print "Received value: " + str(msg.value)

uuid = 'subscriberuuid1'
config = Config('127.0.0.1', 4444, '127.0.0.1', 5505)
simple_subscriber = Subscriber(uuid, "Room = '410'", config)
simple_subscriber.attach_diff_handler(diff_handler)
simple_subscriber.attach_publish_handler(publish_handler)
simple_subscriber.start()

time.sleep(60)
