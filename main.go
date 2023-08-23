package main

import (
	"backend/db"
	"backend/http"
	"flag"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"time"
)
import _ "net/http/pprof"

var (
	maxQueueSize = flag.Int("max_queue_size", 5000, "The size of job queue")
	batchSize    = flag.Int("batch_size", 1000, "The batch size")
	batchSleep   = flag.Int("batch_sleep", 2000, "Sleep time before batch")
)

type BatchHandler func(batchJobs []handler.Job)

func main() {

	flag.Parse()

	batchChannelPessoaSearch := make(chan handler.Job, *maxQueueSize)
	batchChannelPessoa := make(chan handler.Job, *maxQueueSize)

	go processBatch(batchChannelPessoaSearch, handler.SavePessoaSearchBatch)
	go processBatch(batchChannelPessoa, handler.SavePessoaBatch)

	log.Println("Starting server on 9999")

	time.Sleep(3 * time.Second) // wait for db is up

	conn := db.GetConnection()
	defer conn.Close()

	router := httprouter.New()

	router.POST("/pessoas", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		handler.CreatePessoa(batchChannelPessoa, batchChannelPessoaSearch, writer, request)
	})
	router.GET("/pessoas", handler.GetPessoas)
	router.GET("/pessoas/:id", handler.GetPessoa)
	router.GET("/contagem-pessoas", handler.GetPessoaCount)

	log.Fatal(http.ListenAndServe(":9999", router))
}

func processBatch(batchChannel chan handler.Job, batchHandler BatchHandler) {
	for {
		var batchJobs []handler.Job

		batchJobs = append(batchJobs, <-batchChannel)

		time.Sleep(time.Duration(*batchSleep) * time.Millisecond)

	Batch:

		for i := 0; i < *batchSize; i++ {

			select {
			case job := <-batchChannel:
				batchJobs = append(batchJobs, job)
			default:
				break Batch
			}
		}

		batchHandler(batchJobs)
	}
}
