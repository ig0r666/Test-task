up: down
	docker compose up --build -d

down:
	docker compose down

clean:
	docker compose down -v

balancer-up: down
	docker compose up --build -d balancer serverpool

limiter-up: down
	docker compose up --build -d limiter postgres
