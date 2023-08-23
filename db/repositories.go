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

func CheckPessoaExistsByApelido(conn PgxIface, apelido string) (bool, error) {
	sql := "select count(*) > 0 from pessoa where apelido = $1"

	var exists bool

	err := conn.QueryRow(context.Background(), sql, apelido).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func FindPessoas(conn PgxIface, search string) ([]string, error) {

	sql := `select pessoaJson 
		from pessoa_search w
		where search ilike '%' || $1 || '%'
		limit 50`

	result, err := conn.Query(context.Background(), sql, search)

	if err != nil {
		//log.Println(err)
		return nil, err
	}

	pessoas := []string{}

	for {
		if !result.Next() {
			break
		}
		var pessoa []byte

		err := result.Scan(&pessoa)

		if err != nil {
			continue
		}

		pessoas = append(pessoas, string(pessoa))
	}

	return pessoas, nil
}

type ApelidoError struct {
	Msg string
}

func (e *ApelidoError) Error() string {
	return e.Msg
}

func SavePessoaSearch(conn PgxIface, pessoa Pessoa, pessoaJson []byte) error {

	sql := `insert into pessoa_search values ( $1, $2 );`

	search := strings.ToLower(strings.Join(pessoa.Stack, " ") + " " + pessoa.Nome + " " + pessoa.Apelido)

	_, err := conn.
		Exec(context.Background(), sql, search, pessoaJson)

	//log.Println(fmt.Sprintf("executed insert with result %+v", exec))

	if err != nil {
		//log.Println(fmt.Sprintf("error executing insert %v", err))
		return err
	}

	return nil
}

func SavePessoaSearchBatch(conn PgxIface, pessoaBatch []Pessoa, pessoaJson [][]byte) error {

	sql := "insert into pessoa_search values %s"

	params := []interface{}{}
	paramSql := []string{}

	for index, pessoa := range pessoaBatch {
		i := index * 2
		search := strings.ToLower(strings.Join(pessoa.Stack, " ") + " " + pessoa.Nome + " " + pessoa.Apelido)
		params = append(params, search, pessoaJson[index])
		paramSql = append(paramSql, fmt.Sprintf("($%d, $%d)", i+1, i+2))
	}

	_, err := conn.
		Exec(context.Background(), fmt.Sprintf(sql, strings.Join(paramSql, ",")), params...)

	//log.Println(fmt.Sprintf("executed insert with result %+v", exec))

	if err != nil {
		log.Println(fmt.Sprintf("error executing insert %v", err))
		return err
	}

	return nil
}

func SavePessoaBatch(conn PgxIface, pessoaBatch []Pessoa) error {

	sql := "insert into pessoa values %s"

	params := []interface{}{}
	paramSql := []string{}

	for index, pessoa := range pessoaBatch {
		i := index * 5
		params = append(params, pessoa.Id, pessoa.Apelido, pessoa.Nome, pessoa.Nascimento, pessoa.Stack)
		paramSql = append(paramSql, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d::varchar[])", i+1, i+2, i+3, i+4, i+5))
	}

	_, err := conn.
		Exec(context.Background(), fmt.Sprintf(sql, strings.Join(paramSql, ",")), params...)

	//log.Println(fmt.Sprintf("executed insert with result %+v", exec))

	if err != nil {
		log.Println(fmt.Sprintf("error executing insert %v", err))
		return err
	}

	return nil
}

func SavePessoa(conn PgxIface, id uuid.UUID, pessoa Pessoa) error {

	sql := `insert into pessoa values ( $1, $2, $3, $4, $5::varchar[]);`

	_, err := conn.
		Exec(context.Background(), sql, id, pessoa.Apelido, pessoa.Nome, pessoa.Nascimento, pessoa.Stack)

	//log.Println(fmt.Sprintf("executed insert with result %+v", exec))

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
