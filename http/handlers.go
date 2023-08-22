package handler

import (
	"backend/db"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/flier/gohs/hyperscan"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

var (
	GetConnection = db.GetConnection
	regexpr, _    = hyperscan.
			NewBlockDatabase(hyperscan.NewPattern("^\\d{4}\\-(0[1-9]|1[012])\\-(0[1-9]|[12][0-9]|3[01])$", hyperscan.SingleMatch))

	redisConnection = redis.NewClient(&redis.Options{Addr: getRedisString()})
	ctx             = context.Background()
)

func getRedisString() string {
	redisString := os.Getenv("REDIS_URL")

	if redisString == "" {
		return "localhost:6379"
	}
	return redisString
}

func GetPessoa(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	param := ps.ByName("id")
	id, err := uuid.Parse(param)

	if err != nil {
		http.NotFound(w, r)
		log.Println(fmt.Sprintf("get pessoa with invalid uuid %s %s", param, err))
		return
	}

	log.Println(fmt.Sprintf("get pessoa by id %s", id))

	result, err := redisConnection.Get(ctx, "id::"+id.String()).Result()

	if err == nil {
		log.Printf("found in cache %+v\n", result)
		_, err := w.Write([]byte(result))

		if err != nil {
			log.Println(err)
			return
		}
		return
	} else {
		log.Println(err)
	}

	pessoa, err := db.GetPessoaById(GetConnection(), id)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	pessoaJson, _ := json.Marshal(pessoa)
	_, err = w.Write(pessoaJson)

	if err != nil {
		log.Println(err)
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

	log.Println(fmt.Sprintf("get pessoas by term %s", searchTerm))

	pessoaJson, _ := json.Marshal(pessoas)
	_, err = w.Write(pessoaJson)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func CreatePessoa(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	bytes, err := io.ReadAll(r.Body)

	pessoa, err := parsePessoa(bytes)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error parsing input %s\n", err)
		return
	}

	id, _ := uuid.NewUUID()
	pessoa.Id = id

	pessoaJson, _ := json.Marshal(pessoa)

	err = redisConnection.Set(ctx, "id::"+id.String(), pessoaJson, 0).Err()
	if err != nil {
		log.Println(err)
	}

	err = redisConnection.Set(ctx, "apelido::"+pessoa.Apelido, id.String(), 0).Err()
	if err != nil {
		log.Println(err)
	}

	err = db.SavePessoa(GetConnection(), id, *pessoa)

	if err != nil {

		var apelidoError *db.ApelidoError

		switch {
		case errors.As(err, &apelidoError):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	log.Println(fmt.Sprintf("created pessoa with id %s : body %+v", id, pessoa))

	w.Header().Set("Location", fmt.Sprintf("/pessoas/%s", id))
	w.WriteHeader(http.StatusCreated)

}

func GetPessoaCount(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	count, err := db.CountPessoa(GetConnection())

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte(strconv.Itoa(int(count))))

	if err != nil {
		log.Println(err)
		return
	}
}

func parsePessoa(bytes []byte) (*db.Pessoa, error) {
	var pessoa db.Pessoa
	err := json.Unmarshal(bytes, &pessoa)

	if err != nil {
		return nil, err
	}

	result, err := redisConnection.Get(ctx, "apelido::"+pessoa.Apelido).Result()

	if err == nil {
		return nil, fmt.Errorf("field apelido duplicated %+v", result)
	}

	if pessoa.Nome == "" || pessoa.Apelido == "" || (pessoa.Stack != nil && len(pessoa.Stack) == 0) {
		return nil, fmt.Errorf("field cannot be null")
	}

	if !regexpr.MatchString(pessoa.Nascimento) {
		return nil, fmt.Errorf("field nascimento invalid")
	}

	log.Printf(fmt.Sprintf("parsed pessoa %+v", pessoa))

	return &pessoa, nil
}
