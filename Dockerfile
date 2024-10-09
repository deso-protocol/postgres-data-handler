FROM alpine:latest AS postgres-data-handler

RUN apk update
RUN apk upgrade
RUN apk add --update bash go cmake g++ gcc git make vips-dev

COPY --from=golang:1.23-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /postgres-data-handler/src

COPY postgres-data-handler/go.mod postgres-data-handler/
COPY postgres-data-handler/go.sum postgres-data-handler/
COPY core/go.mod                  core/
COPY core/go.sum                  core/
COPY backend/go.mod               backend/
COPY backend/go.sum               backend/
COPY state-consumer/go.mod        state-consumer/
COPY state-consumer/go.sum        state-consumer/

WORKDIR /postgres-data-handler/src/postgres-data-handler

RUN go mod download

# include postgres data handler src
COPY postgres-data-handler/entries    entries
COPY postgres-data-handler/migrations migrations
COPY postgres-data-handler/handler    handler
COPY postgres-data-handler/main.go    .

# include core src
COPY core/desohash    ../core/desohash
COPY core/consensus   ../core/consensus
COPY core/collections ../core/collections
COPY core/bls         ../core/bls
COPY core/cmd         ../core/cmd
COPY core/lib         ../core/lib
COPY core/migrate     ../core/migrate

# include backend src
COPY backend/apis      ../backend/apis
COPY backend/config    ../backend/config
COPY backend/cmd       ../backend/cmd
COPY backend/miner     ../backend/miner
COPY backend/routes    ../backend/routes
COPY backend/countries ../backend/countries

# include state-consumer src
COPY state-consumer/consumer ../state-consumer/consumer

RUN go mod tidy

## build postgres data handler backend
RUN GOOS=linux go build -mod=mod -a -installsuffix cgo -o bin/postgres-data-handler main.go

ENTRYPOINT ["/postgres-data-handler/src/postgres-data-handler/bin/postgres-data-handler"]
