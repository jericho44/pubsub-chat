## Jalankan lokal

`bash
go mod tidy
go run ./cmd/server
`

`bash http://localhost:8080 `

## Jalankan Docker

`bash
docker compose up --build
`

atau

`bash
docker build -t pubsub-chat:latest .
docker run --rm -p 8080:8080 pubsub-chat:latest
`
