#!/usr/bin/env python3
import json
import multiprocessing
import os
import socket
import sys
import time
import uuid
import flask
import paho.mqtt.client as mqtt

def start_server():
  app = flask.Flask(__name__)

  @app.get("/command")
  def command():
    with open("python/command") as handler:
      return handler.read()

  @app.post("/upload")
  def upload():
    print(flask.request)
    file = flask.request.get('file')
    file.save(file.filename)
    print(flask.request.files.get('file'))
    return 'hi'

  app.run(host="0.0.0.0", port=8000)

def get_ip_address():
  host_ip = ""
  s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
  s.connect(("8.8.8.8", 80))
  host_ip = s.getsockname()[0]
  s.close()
  return host_ip

# This is changed everytime you refresh the box and register the machine again.
CLIENT_ID = "d0bf7de7-c5c3-43c5-8094-d9f327129cdc"
BROKER = '127.0.0.1'
BROKER_PORT = 1883
TOPIC = f"yggdrasil/{CLIENT_ID}/data/in"
IP_ADDRESS = get_ip_address()

MESSAGE = {
  "type": "data",
  "message_id": str(uuid.uuid4()),
  # client_uuid doesn't seemt to be us  ed
  # "client_uuid": CLIENT_ID,
  "version": 1,
  "sent": "2021-01-12T14:58:13+00:00", # str(datetime.datetime.now().isoformat()),
  "directive": 'rhc-bash-worker',
  "content": f'http://{IP_ADDRESS}:8000/command',
  "metadata": {
    "report_file": "/var/log/convert2rhel/convert2rhel-report.json",
    "return_url": f'http://{IP_ADDRESS}:8000/upload',
  }
}


def main():
  if not os.path.exists("python/command"):
    print("You must create a python/command file in order to continue.")
    sys.exit(1)

  process = multiprocessing.Process(target=start_server, args=())
  process.start()

  print("Sleeping for 1 second to wait for the server")
  time.sleep(1)

  client = mqtt.Client()
  client.connect(BROKER, BROKER_PORT, 60)
  client.publish(TOPIC, json.dumps(MESSAGE), 1, False)
  print("Published message to MQTT, serving content.")

  process.join()



if __name__ == "__main__":
   main()
