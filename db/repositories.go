package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"log"
	"strings"
)

func CountPessoa(conn PgxIface) (int64, error) {
	sql := `select count(*) from pessoa`

	var count int64

	err := conn.QueryRow(context.Background(), sql).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetPessoaById(conn PgxIface, id uuid.UUID) (*Pessoa, error) {
	sql := "select id, apelido, nome, nascimento, stack from pessoa where id = $1"

	var pessoa Pessoa

	err := conn.QueryRow(context.Background(), sql, id).Scan(
		&pessoa.Id,
		&pessoa.Apelido,
		&pessoa.Nome,
		&pessoa.Nascimento,
		&pessoa.Stack)

	if err != nil {
		return nil, err
	}

	return &pessoa, nil
}

func FindPessoas(conn PgxIface, search string) ([]Pessoa, error) {

	sql := `select id, apelido, nome, nascimento, stack 
		from pessoa w
		where stack_s ilike '%' || $1 || '%s'
		limit 50`

	result, err := conn.Query(context.Background(), sql, search)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	pessoas := []Pessoa{}

	for {
		if !result.Next() {
			break
		}
		var pessoa Pessoa

		err := result.Scan(
			&pessoa.Id,
			&pessoa.Apelido,
			&pessoa.Nome,
			&pessoa.Nascimento,
			&pessoa.Stack)

		if err != nil {
			continue
		}

		pessoas = append(pessoas, pessoa)
	}

	return pessoas, nil
}

type ApelidoError struct {
	Msg string
}

func (e *ApelidoError) Error() string {
	return e.Msg
}

func SavePessoa(conn PgxIface, id uuid.UUID, pessoa Pessoa) error {

	sql := `insert into pessoa values ( $1, $2, $3, $4, $5::varchar[], $6);`

	search := strings.ToLower(strings.Join(pessoa.Stack, " ") + " " + pessoa.Nome + " " + pessoa.Apelido)

	exec, err := conn.
		Exec(context.Background(), sql, id, pessoa.Apelido, pessoa.Nome, pessoa.Nascimento, pessoa.Stack, search)

	log.Println(fmt.Sprintf("executed insert with result %+v", exec))

	if err != nil {
		log.Println(fmt.Sprintf("error executing insert %v", err))

		var pgerr *pgconn.PgError

		switch {
		case errors.As(err, &pgerr):
			{
				if pgerr.Code == "23505" && pgerr.ConstraintName == "pessoa_apelido_index" {
					return &ApelidoError{}
				}

			}
		default:
			return err
		}
	}

	return nil
}
