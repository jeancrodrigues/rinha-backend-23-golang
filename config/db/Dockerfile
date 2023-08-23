FROM postgres as dumper

COPY db-init-entrypoint.sh /docker-entrypoint-initdb.d/

RUN ["sed", "-i", "s/exec \"$@\"/echo \"skipping...\"/", "/usr/local/bin/docker-entrypoint.sh"]

ENV PG_USER=postgres
ENV PGDATA=/data
ENV POSTGRES_USER=pessoa
ENV POSTGRES_PASSWORD=pessoa
ENV POSTGRES_DB=pessoa

RUN ["/usr/local/bin/docker-entrypoint.sh", "postgres"]

FROM postgres

COPY --from=dumper /data $PGDATA