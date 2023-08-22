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
      stack      character varying[],
      stack_s    character varying
  );

  alter table public.pessoa owner to pessoa;

  create unique index pessoa_apelido_index on public.pessoa (apelido);

  create index pessoa_nome_index on public.pessoa (nome);

  create index pessoa_stack_index on public.pessoa (stack);

  create index pessoa_stacks_gin_index ON public.pessoa USING gin (stack_s gin_trgm_ops);

EOSQL