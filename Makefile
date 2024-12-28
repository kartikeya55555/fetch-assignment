# Variables
APP_NAME := fetch-assignment
PORT := 8080
SERVER_URL := http://localhost:$(PORT)

############################################
# Primary Targets
############################################

build:
	docker-compose build

run: build
	docker-compose up -d
	@echo "All containers (LocalStack + API/Worker) are now running."

stop:
	docker-compose down
	@echo "All containers stopped."

logs:
	docker-compose logs -f

############################################
# Testing and Other Commands
############################################

test: run
	@echo "Waiting for server to be ready..."
	@sleep 3
	@echo "Running tests..."
	go test ./... -v

clean:
	docker-compose down --rmi all
