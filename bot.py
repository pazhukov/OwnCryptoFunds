import logging
import requests
from telegram.ext import Updater, CommandHandler, MessageHandler, Filters, CallbackQueryHandler
from telegram import InlineKeyboardButton, InlineKeyboardMarkup, ReplyKeyboardMarkup

DB_API = 'http://localhost:23001'
QUEUE_API = 'http://localhost:23000'
PRICE_API = 'https://min-api.cryptocompare.com/data/price?fsym=#CRYPTO#&tsyms=USD&api_key={API_KEY_CRYPTO}'

# Enable logging
logging.basicConfig(format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
                    level=logging.INFO)

logger = logging.getLogger(__name__)

def start(update, context):
    """Send a message when the command /start is issued."""
    tg_id = update.message.chat_id
    update.message.reply_text('-=Crypto Funds=-\n\nEasy way to buy or sell crypto')
    r = requests.post(DB_API + '/new/investor', json={"tg_id": str(tg_id)})
    if r.status_code == 500:
        update.message.reply_text('Some problem with you profile')    


def echo(update, context):
    update.message.reply_text("Use /help for help")


def help(update, context):
    update.message.reply_text("Help menu\n\n/buy - buy crypto funds\n/sell - sell crypto\n/portfolio - show portfolio")


def portfolio(update, context):
    tg_id = update.message.chat_id
    r = requests.get(DB_API + '/portfolio', json={"tg_id": str(tg_id)})
    if r.status_code == 500:
        update.message.reply_text('Some problem with you profile')  
        return 
    data = r.json()   

    out_message = "Your portfolio\n"

    for x in data['funds']:
        if x['count'] <= 0:
            continue
        fund_commands = ""
        if x['fund'] == "BTC":
            fund_commands = "Buy /buy_btc or Sell /sell_btc"
        elif x['fund'] == "XRP":   
            fund_commands = "Buy /buy_xrp or Sell /sell_xrp"    
        out_message = out_message + "\n" + "* " + x['fund'] + " - " + str(x['count']) + " units\n"
        out_message = out_message + "Balance amount " + str(round(x['amount'], 2)) + "$\n"
        price_http = requests.get(PRICE_API.replace("#CRYPTO#", x['fund']))
        if price_http.status_code == 200:
            data_price = price_http.json()
            usd = data_price['USD']
            market_value = x['crypto_amount'] * usd
            out_message = out_message + "Market amount " + str(round(market_value, 2)) + "$\n" + fund_commands + "\n"
    update.message.reply_text(out_message)


def buy(update, context):
    out_message = "Our crypto funds\n\n"
    out_message = out_message + "cfBTC - invest in Bitcoin\n/buy_btc\n1 unit of cfBTC = 0.0001 BTC\n\n"  
    out_message = out_message + "cfXRP - invest in Ripple\n/buy_xrp\n1 unit of cfXRP = 1 XRP"   
    update.message.reply_text(out_message)

def sell(update, context):
    out_message = "Use /portfolio for sell units"
    update.message.reply_text(out_message)    

def buy_btc(update, context):
    tg_id = update.message.chat_id
    out_message = "How much do you want to invest?"
    keyboard = []
    buttons = []
    buttons.append(InlineKeyboardButton("25 $", callback_data="buy_btc_25"))
    buttons.append(InlineKeyboardButton("50 $", callback_data="buy_btc_50"))
    buttons.append(InlineKeyboardButton("100 $", callback_data="buy_btc_100"))  
    keyboard.append(buttons)
    reply_markup = InlineKeyboardMarkup(keyboard)           
    context.bot.send_message(chat_id=tg_id, text=out_message, reply_markup=reply_markup)  


def buy_xrp(update, context):
    tg_id = update.message.chat_id
    out_message = "How much do you want to invest?"
    keyboard = []
    buttons = []
    buttons.append(InlineKeyboardButton("15 $", callback_data="buy_xrp_15"))
    buttons.append(InlineKeyboardButton("30 $", callback_data="buy_xrp_30"))
    buttons.append(InlineKeyboardButton("50 $", callback_data="buy_xrp_500"))  
    keyboard.append(buttons)
    reply_markup = InlineKeyboardMarkup(keyboard)           
    context.bot.send_message(chat_id=tg_id, text=out_message, reply_markup=reply_markup) 

def sell_btc(update, context):
    tg_id = update.message.chat_id
    clientCanSell = checkSell(tg_id, "BTC")
    if clientCanSell == False:
        update.message.reply_text("You doesn't have units BTC fund")
        return
    out_message = "How much do you want to sell?"
    keyboard = []
    buttons = []
    buttons.append(InlineKeyboardButton("25 %", callback_data="sell_btc_25"))
    buttons.append(InlineKeyboardButton("50 %", callback_data="sell_btc_50"))
    buttons.append(InlineKeyboardButton("100 %", callback_data="sell_btc_100"))  
    keyboard.append(buttons)
    reply_markup = InlineKeyboardMarkup(keyboard)           
    context.bot.send_message(chat_id=tg_id, text=out_message, reply_markup=reply_markup)     

def sell_xrp(update, context):
    tg_id = update.message.chat_id
    clientCanSell = checkSell(tg_id, "XRP")
    if clientCanSell == False:
        update.message.reply_text("You doesn't have units XRP fund")
        return
    out_message = "How much do you want to sell?"
    keyboard = []
    buttons = []
    buttons.append(InlineKeyboardButton("25 %", callback_data="sell_xrp_25"))
    buttons.append(InlineKeyboardButton("50 %", callback_data="sell_xrp_50"))
    buttons.append(InlineKeyboardButton("100 %", callback_data="sell_xrp_100"))  
    keyboard.append(buttons)
    reply_markup = InlineKeyboardMarkup(keyboard)           
    context.bot.send_message(chat_id=tg_id, text=out_message, reply_markup=reply_markup)  


def checkSell(tg_id, fund):
    clientCanSell = False
    r = requests.get(DB_API + '/portfolio', json={"tg_id": str(tg_id)})
    if r.status_code == 500:
        return clientCanSell
    data = r.json()   
    for x in data['funds']:
        if x['fund'] == fund and x['count'] > 0:
            clientCanSell = True

    return  clientCanSell 

def getQtyForSell(tg_id, fund, percent):
    qty = 0
    r = requests.get(DB_API + '/portfolio', json={"tg_id": str(tg_id)})
    if r.status_code == 500:
        return qty
    data = r.json()   
    for x in data['funds']:
        if x['fund'] == fund and percent == 100: 
            qty = x['count']
        elif x['fund'] == fund and percent != 100:        
            qty = round((percent/100) * x['count'])

    return  qty       


def buttons_actions(update, context):  
    tg_id = update.callback_query.message.chat_id
    query = update.callback_query
    query.answer()

    out_message = ""
    fund = ""
    type = ""
    amount = 0
    
    r = requests.get(DB_API + '/investor', json={"tg_id": str(tg_id)})
    if r.status_code == 500:
        query.edit_message_text(text='Some problem with you profile')
        return
    data = r.json()
    guid =  data['id']         

    query_data = query.data
    if(query_data.find("buy_btc_") > -1):
        amount = int(query_data.replace("buy_btc_", ""))
        fund = "BTC"
        type = "buy"    
    elif (query_data.find("buy_xrp_") > -1):
        amount = int(query_data.replace("buy_xrp_", ""))    
        fund = "XRP"
        type = "buy" 
    elif (query_data.find("sell_xrp_") > -1):
        amount = int(query_data.replace("sell_xrp_", ""))    
        fund = "XRP"
        type = "sell" 
    elif (query_data.find("sell_btc_") > -1):
        amount = int(query_data.replace("sell_btc_", ""))    
        fund = "BTC"
        type = "sell"                 

    data = {}
    if type == "buy":
        data = {"investor": guid, "fund":fund, "amount":amount}
        r = requests.post(QUEUE_API + '/new/invest', json=data)
    elif type == "sell": 
        qty = getQtyForSell(tg_id, fund, amount);
        data = {"investor": guid, "fund":fund, "qty":qty}
        r = requests.post(QUEUE_API + '/new/sell', json=data)
    else:
        out_message = "Some problem with operation" 
        query.edit_message_text(text=out_message)
        return    

    if r.status_code == 200:
        out_message = "Your order is acceptred. Please wait when operation is finish"
    else:
        out_message = "Some problem with operation" 

    query.edit_message_text(text=out_message)


def error(update, context):
    """Log Errors caused by Updates."""
    logger.warning('Update "%s" caused error "%s"', update, context.error)

def main():

    updater = Updater("{API_KEY_BOT}", use_context=True)
	
    dp = updater.dispatcher

    dp.add_handler(CommandHandler("start", start))
    dp.add_handler(CommandHandler("help", help))
    dp.add_handler(CommandHandler("portfolio", portfolio))
    dp.add_handler(CommandHandler("buy", buy))
    dp.add_handler(CommandHandler("sell", sell))
    dp.add_handler(CommandHandler("buy_btc", buy_btc))
    dp.add_handler(CommandHandler("buy_xrp", buy_xrp))
    dp.add_handler(CommandHandler("sell_btc", sell_btc))
    dp.add_handler(CommandHandler("sell_xrp", sell_xrp))


    dp.add_handler(CallbackQueryHandler(buttons_actions))

    dp.add_handler(MessageHandler(Filters.text, echo))

    dp.add_error_handler(error)

    updater.start_polling()

    updater.idle()

if __name__ == '__main__':
    main()