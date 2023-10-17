FROM golang

COPY ./ /backend
WORKDIR /backend
RUN go mod tidy && go mod vendor && go build -o ./ticket_service/ ticket_service/main.go

ENTRYPOINT ["/backend/ticket_service/main"]