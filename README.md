# OwnCryptoFunds

## Read it firstly

This is just an example of how a structure for the operation of collective investments (investment funds) can be organized in the world of cryptocurrency. At the same time, similar processes work for the world of conventional finance.

This information is not an individual investment recommendation, and the financial instruments or transactions mentioned in it may not correspond to your investment profile and investment goals (expectations). It is your task to determine whether a financial instrument or transaction matches your interests, investment objectives, investment horizon and acceptable risk level. I am not responsible for possible losses in case of making transactions or investing in the financial instruments mentioned in this information, and I do not recommend using this information as the sole source of information when making an investment decision.

With everything described above, the interaction between the client and the exchange will not be described - how to transfer money on the exchange, how to withdraw and transfer them to the client

## Zero step

1. Get Telegram BOT API KEY and replace {API_KEY_BOT} in bot.py 
2. Get cryptocompare.com API KEY and replace {API_KEY_CRYPTO} in bot.py
3. Get API KEY from Binance Testnet https://testnet.binance.vision/ in replace BINANCE_API_KEY and BINANCE_API_SECRET in .env
4. Install RabbitMQ and setup with new user and replace RABBITMQ_USER, RABBITMQ_PWD, RABBITMQ_HOST and RABBITMQ_PORT in .env

## How run all

1. Run bot.py for start Telegram bot
2. Run buy_reader.py, sell_reader.py and order_reader.py for reading queues
3. Run bd_api.go for work with SQLite DB
4. Run queue_api.gi for adding message to queues
5. Try bot

## How it works?

If you choose /buy_btc or /buy_xrp you can choose how amount you invest in fund. After choosing in queue BUY_QUEUE creating new message
```
{"investor":"6eb87d29-24a1-4529-a248-10468a5a429f","fund":"BTC","amount":50}
```
buy_reader.py read queue BUY_QUEUE and create order via Binace API. After creating order buy_reader.py add message to ORDER_QUEUE 
```
{"order_id":"8525569","investor":"6eb87d29-24a1-4529-a248-10468a5a429f","fund":"BTC","qty":0.0022}
```
order_reader.py read queue ORDER_QUEUE and accept order if it confirmed and filled/ That meaning you get some units of fund

Now with /portfolio you can see full you portfolio with how many money you invest in cryptofunds and it market value

Example
```
Your portfolio

* BTC - 39.4 units
Balance amount 66.31$
Market amount 89.78$
Buy /buy_btc or Sell /sell_btc

* XRP - 37.8 units
Balance amount 14.92$
Market amount 15.29$
Buy /buy_xrp or Sell /sell_xrp
```
To invest in the BTC fund, there is a command /buy_btc or sell_btc to exit the investment. An important feature - in the investor's portfolios, the amount of BTC is converted into the number of fund units. That is, the client's portfolio will indicate that he owns 39.4 units - this means that he owns approximately 0.00394 BTC (1 unit = 0.0001 BTC), this is indicated in the table **fundsrate**.

To invest in the XRP fund, there is a command /buy_btc or sell_btc to exit the investment. An important feature - in the investor's portfolios, the amount of BTC is converted into the number of fund units. That is, the client's portfolio will indicate that he owns 100 units - this means that he owns approximately 100 XRP (1 unit = 1 XRP), this is indicated in the table **fundsrate**.

For selling cryptofunds all messages writes in SELL_QUEUE and works that same as BUY QUEUE.

REST API DB using port 23001, REST API queues using port 23000. It helps communicate beetwen DB, queues and Binance.

## What can be improved?

- Error handler
- Show return on investment as a percentage

## More information

:globe_with_meridians: https://anteater.dev/crypto-funds/ 
