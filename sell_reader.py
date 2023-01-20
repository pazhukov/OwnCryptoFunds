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
queue = os.environ['SELL_QUEUE']

# Binance setup
client = Client(api_key, api_secret, testnet=True)

url = 'http://localhost:23000/new/order'
url_db = 'http://localhost:23001/order'

# RabbitMQ setup
credentials = pika.PlainCredentials(user, pwd)
connection = pika.BlockingConnection(pika.ConnectionParameters(host=host, credentials=credentials))
channel = connection.channel()
channel.queue_declare(queue=queue, durable=True)
print(' [*] Waiting for messages. To exit press CTRL+C')


def callback(ch, method, properties, body):
    data = json.loads(body.decode())
    print(" [x] Received ")
    print("Investor " + data["investor"])
    print("Fund " + data["fund"])
    print("Qty " + str(data["qty"]))

    #ch.basic_ack(delivery_tag=method.delivery_tag)
    #return

    crypto_symbol = ""
    if data["fund"] == 'BTC': 
        crypto_symbol = "BTCBUSD"
    elif data["fund"] == 'XRP':
        crypto_symbol = "XRPBUSD"

    # Create Binance Order
    if data["fund"] == 'BTC': 
        crypto_amount = round(data["qty"] * 0.0001, 5)
    elif data["fund"] == 'XRP':
        crypto_amount = round(data["qty"] * 1, 5)         
    order = client.order_market_sell(symbol=crypto_symbol, quantity=crypto_amount)

    # add to order queue
    payload = json.dumps({
        "order_id": str(order["orderId"]),
        "investor": data["investor"],
        "fund": data["fund"],
        "qty": float(order["executedQty"])
        })
    headers = {'Content-Type': 'application/json'}
    response = requests.request("POST", url, headers=headers, data=payload)
    if response.status_code == 200:
        print(" [x] Done add to queue")
        print(response.text)
        payload = json.dumps({
            'id': str(order["orderId"]),
            'type': 'S',
            'investor': data["investor"],
            'fund': data["fund"],
            'qty': float(order["executedQty"]),
            "amount":float(order["cummulativeQuoteQty"])
        })
        headers = {'Content-Type': 'application/json'}
        response = requests.request("POST", url_db, headers=headers, data=payload)
        if response.status_code == 200:
            print(" [x] Done add to db")
            print(response.text)
            ch.basic_ack(delivery_tag=method.delivery_tag)


channel.basic_qos(prefetch_count=1)
channel.basic_consume(queue=queue, on_message_callback=callback)

channel.start_consuming()
