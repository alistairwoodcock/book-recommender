package main;

import (
	"log"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

func checkError(err error){
	if(err != nil){
		print("Error!");
		panic(err);	
	}
}

var databasePath = "./data/data2.db";
var db *sql.DB;

func main(){

	var err error;
	db, err = sql.Open("sqlite3", databasePath);
	if err != nil {
		log.Fatal(err);
	}

	server();
}