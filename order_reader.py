import pika
import time
import json
import requests
import os
from binance.client import Client
from dotenv import load_dotenv

load_dotenv()

#load variables
api_key = os.environ['BINANCE_API_KEY']
api_secret = os.environ['BINANCE_API_SECRET']
user = os.environ['RABBITMQ_USER']
pwd = os.environ['RABBITMQ_PWD']
host = os.environ['RABBITMQ_HOST']
queue = os.environ['ORDER_QUEUE']

# Binance setup
client = Client(api_key, api_secret, testnet=True)

url_db = 'http://localhost:23001/accept/order'

# RabbitMQ setup
credentials = pika.PlainCredentials(user, pwd)
connection = pika.BlockingConnection(pika.ConnectionParameters(host=host, credentials=credentials))
channel = connection.channel()
channel.queue_declare(queue=queue, durable=True)
print(' [*] Waiting for messages. To exit press CTRL+C')


def callback(ch, method, properties, body):
    data = json.loads(body.decode())
    print("===[x] Received ")
    print("Order ID " + data["order_id"])
    print("Investor " + data["investor"])
    print("Fund " + data["fund"])
    print("Order Qty " + str(data["qty"]))

    #ch.basic_ack(delivery_tag=method.delivery_tag)
    #return

    crypto_symbol = ""
    if data["fund"] == 'BTC': 
        crypto_symbol = "BTCBUSD"
    elif data["fund"] == 'XRP':
        crypto_symbol = "XRPBUSD"

    # check status order
    order = client.get_order(symbol=crypto_symbol, orderId=data["order_id"])
    if order['status'] == 'FILLED':
     payload = json.dumps({
        "id": str(order["orderId"])
        })
    headers = {'Content-Type': 'application/json'}
    response = requests.request("POST", url_db, headers=headers, data=payload)
    if response.status_code == 200:       
        print(response.text)
        print("done")
        ch.basic_ack(delivery_tag=method.delivery_tag)


channel.basic_qos(prefetch_count=1)
channel.basic_consume(queue=queue, on_message_callback=callback)

channel.start_consuming()
