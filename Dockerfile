FROM alpine:latest AS handler

RUN apk update
RUN apk upgrade
RUN apk add --update go gcc g++ vips-dev

WORKDIR /postgres-data-handler/src

COPY postgres-data-handler/go.mod postgres-data-handler/
COPY postgres-data-handler/go.sum postgres-data-handler/
COPY core/go.mod core/
COPY core/go.sum core/
COPY state-consumer/go.mod state-consumer/
COPY state-consumer/go.sum state-consumer/

WORKDIR /postgres-data-handler/src/postgres-data-handler

RUN go mod download

# include postgres data handler src
COPY postgres-data-handler/entries        entries
COPY postgres-data-handler/migrations    migrations
COPY postgres-data-handler/handler    handler
COPY postgres-data-handler/main.go       .

# include core src
COPY core/desohash ../core/desohash
COPY core/cmd       ../core/cmd
COPY core/lib       ../core/lib
COPY core/migrate   ../core/migrate

#COPY state-consumer/consumer    ../state-consumer/consumer

RUN go mod tidy

# Install Delve debugger, specifying the installation path explicitly
ENV GOPATH=/root/go
RUN go install github.com/go-delve/delve/cmd/dlv@latest

## build postgres data handler backend
RUN GOOS=linux go build -mod=mod -a -installsuffix cgo -o bin/postgres-data-handler main.go

# Start a new stage from scratch for the final image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache vips-dev

# Copy the compiled binary and the Delve binary from the build stage
COPY --from=handler /postgres-data-handler/src/postgres-data-handler/bin/postgres-data-handler /app/
COPY --from=handler /root/go/bin/dlv /bin/

# Expose the port Delve will listen on
EXPOSE 2345

# Set the entry point to start the application under Delve's control
ENTRYPOINT ["/bin/dlv", "--listen=:2345", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "/app/postgres-data-handler"]
