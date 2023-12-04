FROM golang:1.16

WORKDIR /BaumanS2SBot

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o BaumanS2SBot ./cmd/app/main.go

RUN chmod +x BaumanS2SBot

CMD [ "./BaumanS2SBot" ]