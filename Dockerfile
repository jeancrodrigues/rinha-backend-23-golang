FROM golang:1.21

RUN apt update -y && apt install -y libhyperscan-dev

WORKDIR /usr/src/app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go install backend/db
RUN go install backend/http

RUN go build -o /usr/local/bin/app

EXPOSE 9999
CMD [ "app" ]