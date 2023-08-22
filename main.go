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
	maxWorkers   = flag.Int("max_workers", 10, "The number of workers to start")
	batchSize    = flag.Int("batch_size", 1000, "The batch size")
	batchSleep   = flag.Int("batch_sleep", 2000, "Sleep time before batch")
)

func main() {

	flag.Parse()

	batchChannel := make(chan handler.Job, *maxQueueSize)

	go func() {
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

			handler.MaterializePessoaToSearchTableBatch(batchJobs)

		}
	}()

	//jobs := make(chan handler.Job, *maxQueueSize)
	//for i := 1; i <= *maxWorkers; i++ {
	//	go func(i int) {
	//		for j := range jobs {
	//			batchChannel <- j
	//		}
	//	}(i)
	//}

	log.Println("Starting server on 9999")

	time.Sleep(10 * time.Second) // wait for db is up

	conn := db.GetConnection()
	defer conn.Close()

	router := httprouter.New()

	router.POST("/pessoas", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		handler.CreatePessoa(batchChannel, writer, request, params)
	})
	router.GET("/pessoas", handler.GetPessoas)
	router.GET("/pessoas/:id", handler.GetPessoa)
	router.GET("/contagem-pessoas", handler.GetPessoaCount)

	log.Fatal(http.ListenAndServe(":9999", router))
}
