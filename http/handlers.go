package handler

import (
	"backend/db"
	"context"
	"fmt"
	"github.com/flier/gohs/hyperscan"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/julienschmidt/httprouter"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	GetConnection = db.GetConnection
	regexpr, _    = hyperscan.
			NewBlockDatabase(hyperscan.NewPattern("^\\d{4}\\-(0[1-9]|1[012])\\-(0[1-9]|[12][0-9]|3[01])$", hyperscan.SingleMatch))

	ctx             = context.Background()
	json            = jsoniter.ConfigCompatibleWithStandardLibrary
	redisConnection = redis.NewClient(&redis.Options{Addr: getRedisString()})
)

func getRedisString() string {
	redisString := os.Getenv("REDIS_URL")

	if redisString == "" {
		return "localhost:6379"
	}
	return redisString
}

type Job struct {
	name       string
	Pessoa     *db.Pessoa
	PessoaJson []byte
}

func GetPessoa(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	param := ps.ByName("id")
	id, err := uuid.Parse(param)

	if err != nil {
		http.NotFound(w, r)
		//log.Println(fmt.Sprintf("get Pessoa with invalid uuid %s %s", param, err))
		return
	}

	//log.Println(fmt.Sprintf("get Pessoa by id %s", id))

	result, err := redisConnection.Get(ctx, "id::"+id.String()).Result()

	if err == nil {
		//log.Printf("found in cache %+v\n", result)
		_, err := w.Write([]byte(fmt.Sprintf("%s", result)))
		if err != nil {
			//log.Println(err)
			return
		}
		return
	} else {
		//log.Println(err)
	}

	pessoa, err := db.GetPessoaById(GetConnection(), id)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	pessoaJson, _ := json.Marshal(pessoa)
	_, err = w.Write(pessoaJson)

	if err != nil {
		//log.Println(err)
		return
	}
}

func GetPessoas(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	searchTerm := r.URL.Query().Get("t")

	if searchTerm == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pessoas, err := db.FindPessoas(GetConnection(), searchTerm)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//log.Println(fmt.Sprintf("get pessoas by term %s", searchTerm))

	_, err = w.Write([]byte(fmt.Sprintf("[%s]", strings.Join(pessoas, ","))))

	if err != nil {
		//log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func CreatePessoa(batchChannelPessoa chan Job, batchChannelPessoaSearch chan Job,
	w http.ResponseWriter, r *http.Request) {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		//log.Printf("error reading body %s\n", err)
		return
	}

	handleCreatePessoa(batchChannelPessoa, batchChannelPessoaSearch, bytes, w)
}

func getNewUuid(resultChan chan uuid.UUID) {
	resultChan <- uuid.New()
}

func handleCreatePessoa(batchChannelPessoa chan Job,
	batchChannelPessoaSearch chan Job, bytes []byte, w http.ResponseWriter) {

	uuidChannel := make(chan uuid.UUID)
	go getNewUuid(uuidChannel)

	parsePessoaChannel := make(chan ParsePessoaResult)
	go parsePessoa(bytes, parsePessoaChannel)

	parseResult := <-parsePessoaChannel
	if parseResult.Error != nil {
		w.WriteHeader(http.StatusBadRequest)
		//log.Printf("error parsing input %s\n", parseResult.Error)
		return
	}

	pessoa := parseResult.Pessoa
	pessoa.Id = <-uuidChannel

	w.Header().Set("Location", fmt.Sprintf("/pessoas/%s", pessoa.Id))
	w.WriteHeader(http.StatusCreated)

	go persistPessoa(batchChannelPessoa, batchChannelPessoaSearch, pessoa)
}

func persistPessoa(batchChannelPessoa chan Job,
	batchChannelPessoaSearch chan Job, pessoa *db.Pessoa) {
	pessoaJson, _ := json.Marshal(pessoa)

	redisConnection.Set(ctx, "id::"+pessoa.Id.String(), pessoaJson, 0)
	redisConnection.Set(ctx, "apelido::"+pessoa.Apelido, true, 0)

	go func() {
		//fmt.Printf("added: %s %s\n", pessoa.Id)
		batchChannelPessoa <- Job{name: pessoa.Id.String(), Pessoa: pessoa}
		batchChannelPessoaSearch <- Job{pessoa.Id.String(), pessoa, pessoaJson}
	}()

	//log.Println(fmt.Sprintf("created Pessoa with id %s : body %+v", pessoa.Id, pessoa))
}

func GetPessoaCount(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	count, err := db.CountPessoa(GetConnection())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte(strconv.Itoa(int(count))))

	if err != nil {
		//log.Println(err)
		return
	}
}

type ParsePessoaResult struct {
	Pessoa *db.Pessoa
	Error  error
}

func parsePessoaNascimento(nascimento string, result chan bool) {
	result <- regexpr.MatchString(nascimento)
}

func checkPessoaExistsCache(apelido string, result chan bool) {
	_, err := redisConnection.Get(ctx, "apelido::"+apelido).Result()

	result <- err == nil
}

func parsePessoa(bytes []byte, result chan ParsePessoaResult) {
	var pessoa db.Pessoa
	err := json.Unmarshal(bytes, &pessoa)

	if err != nil {
		result <- ParsePessoaResult{
			Pessoa: nil,
			Error:  err,
		}
	}

	parseNascimentoChannel := make(chan bool)
	go parsePessoaNascimento(pessoa.Nascimento, parseNascimentoChannel)

	checkApelidoExistsChannel := make(chan bool)
	go checkPessoaExistsCache(pessoa.Apelido, checkApelidoExistsChannel)

	existsCache := <-checkApelidoExistsChannel
	if existsCache {
		result <- resultParsePessoaError("apelido must be unique")
	}

	if pessoa.Nome == "" || pessoa.Apelido == "" || (pessoa.Stack != nil && len(pessoa.Stack) == 0) {
		result <- resultParsePessoaError("field cannot be null")
	}

	nascimentoValido := <-parseNascimentoChannel
	if !nascimentoValido {
		result <- resultParsePessoaError("field nascimento invalid")
	}

	//log.Printf(fmt.Sprintf("parsed Pessoa %+v", pessoa))

	result <- resultParsePessoaSuccess(&pessoa)
}

func resultParsePessoaError(err string) ParsePessoaResult {
	return ParsePessoaResult{
		Pessoa: nil,
		Error:  fmt.Errorf(err),
	}
}

func resultParsePessoaSuccess(pessoa *db.Pessoa) ParsePessoaResult {
	return ParsePessoaResult{
		Pessoa: pessoa,
		Error:  nil,
	}
}

func SavePessoaSearchBatch(batch []Job) {
	//log.Println(fmt.Printf("batch started %s\n", batch[0].name))

	var pessoaBatch []db.Pessoa
	var pessoaJson [][]byte

	for _, job := range batch {
		pessoaBatch = append(pessoaBatch, *job.Pessoa)
		pessoaJson = append(pessoaJson, job.PessoaJson)
	}

	db.SavePessoaSearchBatch(GetConnection(), pessoaBatch, pessoaJson)

	log.Println(fmt.Printf("batch completed %s count %d!\n", batch[0].name, len(batch)))
}

func SavePessoaBatch(batch []Job) {
	//log.Println(fmt.Printf("batch started %s\n", batch[0].name))

	var pessoaBatch []db.Pessoa

	for _, job := range batch {
		pessoaBatch = append(pessoaBatch, *job.Pessoa)
	}

	db.SavePessoaBatch(GetConnection(), pessoaBatch)

	//log.Println(fmt.Printf("batch completed %s count %d!\n", batch[0].name, len(batch)))
}
