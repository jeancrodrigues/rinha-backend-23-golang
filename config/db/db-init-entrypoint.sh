#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL

  create extension pg_trgm;

  create table public.pessoa
  (
      id         uuid    not null constraint pessoa_pk primary key,
      apelido    varchar not null,
      nome       varchar not null,
      nascimento varchar not null,
      stack      character varying[]
  );

  create table public.pessoa_search
  (
      search    text,
      pessoaJson bytea
  );

  alter table public.pessoa owner to pessoa;
  alter table public.pessoa_search owner to pessoa;

  create unique index pessoa_apelido_index on public.pessoa (apelido);

  create index pessoa_stacks_gin_index ON public.pessoa_search USING gist (search gist_trgm_ops(siglen=256));


EOSQL
# create index pessoa_stacks_gin_index ON public.pessoa_search USING gin (search gin_trgm_ops);
# create index pessoa_stacks_gin_index ON public.pessoa USING gist (stack_s gist_trgm_ops(siglen=256));