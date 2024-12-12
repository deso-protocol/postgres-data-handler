dev:
	go run .

dev-env:
	docker compose -f local.docker-compose.yml build && docker compose -f local.docker-compose.yml up

test-env:
	docker compose -f test.docker-compose.yml down --volumes && docker compose -f test.docker-compose.yml build && docker compose -f test.docker-compose.yml up

dev-env-down:
	docker compose -f local.docker-compose.yml down --volumes