#!/usr/bin/env python3
import json
import uuid
import paho.mqtt.client as mqtt

CLIENT_ID = "3c4fb1ac-454c-4643-beb6-c94e159168e0"
BROKER = '127.0.0.1'
BROKER_PORT = 1883
TOPIC = f"yggdrasil/{CLIENT_ID}/data/in"

MESSAGE = {
  "type": "data",
  "message_id": str(uuid.uuid4()),
  # client_uuid doesn't seemt to be used
  # "client_uuid": CLIENT_ID,
  "version": 1,
  "sent": "2021-01-12T14:58:13+00:00", # str(datetime.datetime.now().isoformat()),
  "directive": 'convert2rhel',
  "content": 'https://raw.githubusercontent.com/r0x0d/convert2rhel-worker/main/python/command',
  "metadata": {
    "return_url": 'http://raw.example.com/return'
  }
}


# The callback for when the client receives a CONNACK response from the server.
def on_connect(client, userdata, flags, rc):
    print(f"Connected with result code  {str(rc)}")

    # Subscribing in on_connect() means that if we lose the connection and
    # reconnect then subscriptions will be renewed.
    client.subscribe("$SYS/#")

# The callback for when a PUBLISH message is received from the server.
def on_message(client, userdata, msg):
    print(f"{msg.topic} - {str(msg.payload)}")

client = mqtt.Client()
client.on_connect = on_connect
client.on_message = on_message

client.connect(BROKER, BROKER_PORT, 60)
print(client.publish(TOPIC, json.dumps(MESSAGE), 1, False))
