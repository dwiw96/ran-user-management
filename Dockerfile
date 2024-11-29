# Build excutable file
FROM golang:1.23-alpine  AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# WORKDIR /app/cmd/app
# RUN go build -o ran_user main.go
RUN go build -o /app/cmd/app/ran_user /app/cmd/app/main.go

# Build image
# FROM alpine
# WORKDIR /app
# RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz | tar xvz
# RUN sudo mv migrate /usr/local/bin/
# RUN which migrate
# RUN migrate -path internal/migrations -database "postgresql://user:user@localhost:5432/user?sslmode=disable" -verbose up $(v)
# COPY --from=builder /app/cmd/app/ran_user .
EXPOSE 8080
CMD ["./cmd/app/ran_user"]
