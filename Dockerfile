FROM golang:1.25.4 as builder

WORKDIR /app/
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o porkbun-dns-updater .
RUN chmod +rwx porkbun-dns-updater

FROM golang:1.25.4
COPY --from=builder /app/porkbun-dns-updater .
ENTRYPOINT ["./porkbun-dns-updater"]
