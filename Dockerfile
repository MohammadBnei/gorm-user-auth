FROM golang:alpine

WORKDIR /app

COPY ["go.mod", "go.sum", "./"]

RUN go mod download

COPY . .

RUN go build -o user-api .

FROM alpine

WORKDIR /app

COPY --from=0 /app/user-api .

ENTRYPOINT [ "./user-api" ]