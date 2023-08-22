package db

import (
	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v2"
	"testing"
)

func TestFindPessoas(t *testing.T) {
}

func TestGetPessoaById(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	id := uuid.MustParse("22db9ec4-3ef7-11ee-be56-0242ac120002")

	rows := mock.NewRows([]string{"id", "nome", "apelido", "nascimento", "stack"})

	rows.AddRows([]any{
		id, "jose", "jose", "", []string{""},
	})

	mock.ExpectQuery("^select (.+) from pessoa where id = (.+)$").
		WithArgs(id).
		WillReturnRows(rows)

	// now we execute our method
	if _, err := GetPessoaById(mock, id); err != nil {
		t.Errorf("error was not expected while updating: %s", err)
	}

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSavePessoa(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	id := uuid.MustParse("22db9ec4-3ef7-11ee-be56-0242ac120002")
	pessoa := Pessoa{
		Id:         id,
		Apelido:    "jose",
		Nome:       "jose vanildo",
		Nascimento: "12-12-2012",
		Stack:      []string{"C#"},
	}

	mock.ExpectExec("^insert into pessoa values \\( \\$1, \\$2, \\$3, \\$4, \\$5 \\);$").
		WithArgs(id, pessoa.Apelido, pessoa.Nome, pessoa.Nascimento, pessoa.Stack).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	SavePessoa(mock, id, pessoa)

	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
