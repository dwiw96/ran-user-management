pg-exec:
	docker exec -it $(container) psql -p 5432 -h localhost -U user -d user
pg-stop:
	docker container stop $(container)

rd-exec:
	docker exec -it $(container) redis-cli -h localhost
rd-stop:
	docker container stop $(container)

migrate-create:
	migrate create -ext sql -dir internal/migrations -seq init
migrate-up:
	migrate -path internal/migrations -database "postgresql://user:user@localhost:5432/user?sslmode=disable" -verbose up $(v)
migrate-down:
	migrate -path internal/migrations -database "postgresql://user:user@localhost:5432/user?sslmode=disable" -verbose down $(v)
migrate-force:
	migrate -path internal/migrations -database "postgresql://user:user@localhost:5432/user?sslmode=disable" force $(v)
