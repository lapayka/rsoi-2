FROM golang

COPY ./ /backend
WORKDIR /backend
RUN go mod tidy && go mod vendor && go build -o ./privilege-service/ privilege-service/main.go

ENTRYPOINT ["/backend/privilege-service/main"]