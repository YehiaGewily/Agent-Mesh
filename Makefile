.PHONY: run stress

run:
	@chmod +x ./run.sh
	@./run.sh

stop:
	@docker compose down

stress:
	go run cmd/stress_test/main.go -count=500 -concurrency=20
