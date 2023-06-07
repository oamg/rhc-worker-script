#!/usr/bin/env python3
import json
import socket
import sys
import time
import uuid
from http import server
import paho.mqtt.client as mqtt
import multiprocessing

class CustomHandler(server.SimpleHTTPRequestHandler):
  def __init__(self, request, client_address, server, *, directory = None):
     super().__init__(request, client_address, server, directory=directory)

  def handle_one_request(self):
      super().handle_one_request()
      sys.exit(0)

def start_server(host, port):
  httpd = server.HTTPServer((host, port), CustomHandler)
  httpd.serve_forever()

def get_ip_address():
  host_ip = ""
  s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
  s.connect(("8.8.8.8", 80))
  host_ip = s.getsockname()[0]
  s.close()
  return host_ip

# This is changed everytime you refresh the box and register the machine again.
CLIENT_ID = "a723a681-6e49-4660-a07a-56359a515675"
BROKER = '127.0.0.1'
BROKER_PORT = 1883
TOPIC = f"yggdrasil/{CLIENT_ID}/data/in"

MESSAGE = {
  "type": "data",
  "message_id": str(uuid.uuid4()),
  # client_uuid doesn't seemt to be us  ed
  # "client_uuid": CLIENT_ID,
  "version": 1,
  "sent": "2021-01-12T14:58:13+00:00", # str(datetime.datetime.now().isoformat()),
  "directive": 'rhc-bash-worker',
  "content": f'http://{get_ip_address()}:8000/python/command',
  "metadata": {
    "return_url": 'http://raw.example.com/return'
  }
}

process = multiprocessing.Process(target=start_server, args=('0.0.0.0', 8000))
process.start()

print("Sleeping for 1 second to wait for the server")
time.sleep(1)

client = mqtt.Client()
client.connect(BROKER, BROKER_PORT, 60)
client.publish(TOPIC, json.dumps(MESSAGE), 1, False)
print("Published message to MQTT, serving content.")

process.join()
