FROM alpine:latest AS daodao

RUN apk update
RUN apk upgrade
RUN apk add --update go gcc g++ vips-dev

WORKDIR /postgres-data-handler/src

COPY postgres-data-handler/go.mod postgres-data-handler/
COPY postgres-data-handler/go.sum postgres-data-handler/
COPY backend/go.mod backend/
COPY backend/go.mod backend/
COPY backend/go.sum backend/
COPY core/go.mod core/
COPY core/go.sum core/

WORKDIR /postgres-data-handler/src/postgres-data-handler

RUN go mod download

# include backend src
COPY backend/apis      ../backend/apis
COPY backend/config    ../backend/config
COPY backend/cmd       ../backend/cmd
COPY backend/miner     ../backend/miner
COPY backend/routes    ../backend/routes
COPY backend/countries ../backend/countries

# include daodao src
COPY postgres-data-handler/entries        entries
COPY postgres-data-handler/migrations    migrations
COPY postgres-data-handler/handler    handler
COPY postgres-data-handler/main.go       .

## include core src
COPY core/desohash ../core/desohash
COPY core/cmd       ../core/cmd
COPY core/lib       ../core/lib
COPY core/migrate   ../core/migrate

RUN go mod tidy

## build daodao backend
RUN GOOS=linux go build -mod=mod -a -installsuffix cgo -o bin/postgres-data-handler main.go
#
## create tiny image
#FROM alpine:latest
##
#RUN apk add --update vips-dev
##
#COPY --from=daodao /daodao/src/daodao-backend/bin/daodao-backend /daodao/bin/daodao-backend
#ENTRYPOINT ["/daodao/bin/daodao-backend"]
ENTRYPOINT ["/postgres-data-handler/src/postgres-data-handler/bin/postgres-data-handler"]
