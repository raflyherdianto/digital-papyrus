package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=digital_papyrus sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("DELETE FROM reviews")
	if err != nil {
		log.Fatal("delete reviews:", err)
	}
	_, err = db.Exec("DELETE FROM order_details")
	if err != nil {
		log.Fatal("delete order_details:", err)
	}
	_, err = db.Exec("DELETE FROM orders")
	if err != nil {
		log.Fatal("delete orders:", err)
	}

	fmt.Println("Cleared reviews, order_details, orders. Re-start backend to re-seed with correct UUIDs.")
}
