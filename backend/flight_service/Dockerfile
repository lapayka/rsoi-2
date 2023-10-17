FROM golang

COPY ./ /backend
WORKDIR /backend
RUN go mod tidy && go mod vendor && go build -o ./flight_service/ flight_service/main.go

ENTRYPOINT ["/backend/flight_service/main"]