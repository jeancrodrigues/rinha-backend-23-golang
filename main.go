package main

import (
	"backend/db"
	"backend/http"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"time"
)
import _ "net/http/pprof"

func main() {

	log.Println("Starting server on 9999")

	time.Sleep(10 * time.Second) // wait for db is up

	conn := db.GetConnection()
	defer conn.Close()

	router := httprouter.New()

	router.POST("/pessoas", handler.CreatePessoa)
	router.GET("/pessoas", handler.GetPessoas)
	router.GET("/pessoas/:id", handler.GetPessoa)
	router.GET("/contagem-pessoas", handler.GetPessoaCount)

	log.Fatal(http.ListenAndServe(":9999", router))
}
