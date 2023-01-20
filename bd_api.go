package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"os"

	uuid "github.com/satori/go.uuid"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type InfoMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Investor struct {
	TG string `json:"tg_id"`
	ID string `json:"id"`
}

type Order struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	Investor  string  `json:"investor"`
	Fund      string  `json:"fund"`
	Qty       float64 `json:"qty"`
	Status    int     `json:"status"`
	AmountUSD float64 `json:"amount"`
}

type Portfolio struct {
	Funds []Fund `json:"funds"`
}

type Fund struct {
	Name        string  `json:"fund"`
	Count       float64 `json:"count"`
	AmountUSD   float64 `json:"amount"`
	AmoutCrypto float64 `json:"crypto_amount"`
}

var dbFile = "funds.db"

func main() {

	_, err := os.Stat(dbFile)
	if errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(dbFile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	}

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/new/investor", NewInvestor).Methods("POST")
	router.HandleFunc("/investor", GetInvestor).Methods("GET")
	router.HandleFunc("/order", NewOrder).Methods("POST")
	router.HandleFunc("/accept/order", AcceptOrder).Methods("POST")
	router.HandleFunc("/order", GetOrder).Methods("GET")
	router.HandleFunc("/portfolio", GetPortfolio).Methods("GET")

	log.Fatal(http.ListenAndServe(":23001", router))

}

func NewInvestor(w http.ResponseWriter, r *http.Request) {

	var input Investor
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		var info InfoMessage
		info.Code = 300
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.TG == "" {
		var info InfoMessage
		info.Code = 301
		info.Message = "Investor TG can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	db, err := sql.Open("sqlite3", dbFile)

	if err != nil {
		var info InfoMessage
		info.Code = 500
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)
		return
	}
	defer db.Close()

	investor := uuid.NewV4()

	rows, err := db.Query("select investor from investors where tg_id = ?", input.TG)

	if err != nil {
		var info InfoMessage
		info.Code = 500
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)
	}

	count := 0
	defer rows.Close()
	for rows.Next() {
		count = count + 1
	}

	if count == 0 {

		_, err = db.Exec("insert into investors (tg_id, investor) values (?, ?)", input.TG, investor)

		if err != nil {

			var info InfoMessage
			info.Code = 500
			info.Message = err.Error()
			json.NewEncoder(w).Encode(info)

		} else {

			var info InfoMessage
			info.Code = 200
			info.Message = "Investor added"
			json.NewEncoder(w).Encode(info)
		}
	} else {

		var info InfoMessage
		info.Code = 200
		info.Message = "Investor found"
		json.NewEncoder(w).Encode(info)

	}

	db.Close()
	return

}

func GetInvestor(w http.ResponseWriter, r *http.Request) {

	var input Investor
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		var info InfoMessage
		info.Code = 300
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.TG == "" {
		var info InfoMessage
		info.Code = 301
		info.Message = "Investor TG can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	db, err := sql.Open("sqlite3", dbFile)

	if err != nil {
		var info InfoMessage
		info.Code = 500
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)
		return
	}
	defer db.Close()

	rows, err := db.Query("select investor from investors where tg_id = ?", input.TG)

	if err != nil {
		var info InfoMessage
		info.Code = 500
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)
		return
	}

	investor := ""
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&investor)

		if err != nil {
			var info InfoMessage
			info.Code = 500
			info.Message = err.Error()
			json.NewEncoder(w).Encode(info)
		}
	}

	input.ID = investor

	json.NewEncoder(w).Encode(input)

	db.Close()
	return

}

func NewOrder(w http.ResponseWriter, r *http.Request) {

	var input Order
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		var info InfoMessage
		info.Code = 300
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.ID == "" {
		var info InfoMessage
		info.Code = 301
		info.Message = "Order ID can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.Type == "" {
		var info InfoMessage
		info.Code = 302
		info.Message = "Order Type can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.Investor == "" {
		var info InfoMessage
		info.Code = 303
		info.Message = "Order Investor can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.Fund == "" {
		var info InfoMessage
		info.Code = 304
		info.Message = "Order Fund can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.Qty == 0 {
		var info InfoMessage
		info.Code = 305
		info.Message = "Order Qty can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.AmountUSD == 0 {
		var info InfoMessage
		info.Code = 306
		info.Message = "Order Amount can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	db, err := sql.Open("sqlite3", dbFile)

	if err != nil {
		var info InfoMessage
		info.Code = 500
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)
		return
	}
	defer db.Close()

	_, err = db.Exec("insert into orders (id, type, investor, fund, qty, status, amount) values (?, ?, ?, ?, ?, ?, ?)", input.ID, input.Type, input.Investor, input.Fund, input.Qty, 0, math.Floor(input.AmountUSD*100)/100)

	if err != nil {

		var info InfoMessage
		info.Code = 500
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)

	} else {

		var info InfoMessage
		info.Code = 200
		info.Message = "Order added"
		json.NewEncoder(w).Encode(info)
	}

	db.Close()
	return

}

func GetOrder(w http.ResponseWriter, r *http.Request) {

}

func AcceptOrder(w http.ResponseWriter, r *http.Request) {
	var input Order
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		var info InfoMessage
		info.Code = 300
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.ID == "" {
		var info InfoMessage
		info.Code = 301
		info.Message = "Order ID can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	db, err := sql.Open("sqlite3", dbFile)

	if err != nil {
		var info InfoMessage
		info.Code = 500
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)
		return
	}
	defer db.Close()

	_, err = db.Exec("update orders set status = 1 where id = ?", input.ID)

	if err != nil {

		var info InfoMessage
		info.Code = 500
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)
		return

	}

	var order_id = ""
	var order_type = ""
	var order_investor = ""
	var order_fund = ""
	var order_qty = 0.0
	var order_amount = 0.0
	row := db.QueryRow("select ord.id as id, ord.type as type, ord.investor as investor, ord.fund as fund, ord.qty as amount, round(ord.qty/ifnull(xrate.coin_count, 1), 5) as qty from orders as ord left join fundsrate as xrate on ord.fund = xrate.fund where ord.id = ?", input.ID)
	err_select := row.Scan(&order_id, &order_type, &order_investor, &order_fund, &order_amount, &order_qty)
	if err_select != nil {

		var info InfoMessage
		info.Code = 500
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)

	} else {
		if order_type == "S" {
			order_qty = -1 * order_qty
			order_amount = -1 * order_amount
		}
		_, err = db.Exec("insert into assets (investor, fund, amount, qty, order_id, status) values (?, ?, ?, ?, ?, ?)", order_investor, order_fund, order_qty, order_amount, order_id, 1)

		if err != nil {

			var info InfoMessage
			info.Code = 500
			info.Message = err.Error()
			json.NewEncoder(w).Encode(info)

		} else {

			var info InfoMessage
			info.Code = 200
			info.Message = "Order accepted, units created"
			json.NewEncoder(w).Encode(info)
		}

	}

	db.Close()
	return

}

func GetPortfolio(w http.ResponseWriter, r *http.Request) {

	var input Investor
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		var info InfoMessage
		info.Code = 300
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)
		return
	}

	if input.TG == "" {
		var info InfoMessage
		info.Code = 301
		info.Message = "Investor TG can't be empty"
		json.NewEncoder(w).Encode(info)
		return
	}

	db, err := sql.Open("sqlite3", dbFile)

	if err != nil {
		var info InfoMessage
		info.Code = 500
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)
		return
	}
	defer db.Close()

	rows, err := db.Query("select ast.fund, sum(ast.amount) as fund_amount, sum(ast.qty) as crypto_amount, sum(case when ast.amount < 0 then ord.amount * -1 else ord.amount end) as balance_amount from assets as ast  left join orders as ord on ast.order_id = ord.id where ast.investor in (select investor from investors where tg_id = ?) group by ast.fund", input.TG)

	if err != nil {
		var info InfoMessage
		info.Code = 500
		info.Message = err.Error()
		json.NewEncoder(w).Encode(info)
		return
	}

	var funds []Fund

	defer rows.Close()
	for rows.Next() {
		fund := Fund{}
		err := rows.Scan(&fund.Name, &fund.Count, &fund.AmoutCrypto, &fund.AmountUSD)
		if err != nil {
			var info InfoMessage
			info.Code = 500
			info.Message = err.Error()
			json.NewEncoder(w).Encode(info)
			return
		}

		funds = append(funds, fund)
	}

	portfolio := Portfolio{}
	portfolio.Funds = funds

	json.NewEncoder(w).Encode(portfolio)

	db.Close()
	return

}
