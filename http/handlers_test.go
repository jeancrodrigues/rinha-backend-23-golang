package handler

import (
	"backend/db"
	"github.com/julienschmidt/httprouter"
	"github.com/pashagolub/pgxmock/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreatePessoa_whenValid_shouldCreate(t *testing.T) {
	pessoa := `{
		"apelido" : "josévalidok",
		"nome" : "José Roberto",
		"nascimento" : "2000-10-01",
		"stack" : ["C#", "Node", "Oracle"]
	}`

	recorder, _ := callCreatePessoa(pessoa, t)

	if recorder.Code != 201 {
		t.Error()
	}
}

func TestCreatePessoa_whenStackIsNull_shouldCreate(t *testing.T) {
	pessoa := `{
		"apelido" : "joséstacknullok",
		"nome" : "José Roberto",
		"nascimento" : "2000-10-01",
		"stack" : null
	}`

	recorder, _ := callCreatePessoa(pessoa, t)

	if recorder.Code != 201 {
		t.Error()
	}
}

func TestCreatePessoa_whenNascimentoInvalid_shouldFail(t *testing.T) {
	pessoa := `{
		"apelido" : "josénascimentoinvalid",
		"nome" : "José Roberto",
		"nascimento" : "2000-30-01",
		"stack" : null
	}`

	recorder, _ := callCreatePessoa(pessoa, t)

	if recorder.Code != 400 {
		t.Error()
	}
}

func TestCreatePessoa_whenApelidoIsNull_shouldFail(t *testing.T) {
	pessoa := `{
		"apelido" : null,
		"nome" : "José Roberto",
		"nascimento" : "2000-10-01",
		"stack" : ["C#", "Node", "Oracle"]
	}`

	recorder, _ := callCreatePessoa(pessoa, t)

	if recorder.Code != 400 {
		t.Error()
	}

}

func TestCreatePessoa_whenStackInvalid_shouldFail(t *testing.T) {
	pessoa := `{
		"apelido" : "josestackinvalidfail",
		"nome" : "José Roberto",
		"nascimento" : "2000-10-01",
		"stack" : [1, "Node", "Oracle"]
	}`

	recorder, _ := callCreatePessoa(pessoa, t)

	if recorder.Code != 400 {
		t.Error()
	}
}

func TestCreatePessoa_whenStackEmpty_shouldFail(t *testing.T) {
	pessoa := `{
		"apelido" : "josestackemptyfail",
		"nome" : "José Roberto",
		"nascimento" : "2000-10-01",
		"stack" : []
	}`

	recorder, _ := callCreatePessoa(pessoa, t)

	if recorder.Code != 400 {
		t.Error()
	}
}

func TestCreatePessoa_whenNomeInvalid_shouldFail(t *testing.T) {
	pessoa := `{
		"apelido" : "josenomeinvalidfail",
		"nome" : null,
		"nascimento" : "2000-10-01",
		"stack" : [1, "Node", "Oracle"]
	}`

	recorder, _ := callCreatePessoa(pessoa, t)

	if recorder.Code != 400 {
		t.Error()
	}
}

func TestCreatePessoa_whenNomeNotSet_shouldFail(t *testing.T) {
	pessoa := `{
		"apelido" : "josesemnomefail",
		"nascimento" : "2000-10-01",
		"stack" : ["Go", "Node", "Oracle"]
	}`

	recorder, _ := callCreatePessoa(pessoa, t)

	if recorder.Code != 400 {
		t.Error()
	}
}

func callCreatePessoa(pessoa string, t *testing.T) (*httptest.ResponseRecorder, pgxmock.PgxPoolIface) {
	req, _ := http.NewRequest("POST", "/pessoas", strings.NewReader(pessoa))
	router := httprouter.New()

	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	GetConnection = func() db.PgxIface {
		return mock
	}

	mock.ExpectExec("^insert into Pessoa values (.*)").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	router.Handle("POST", "/pessoas", CreatePessoa)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder, mock
}
