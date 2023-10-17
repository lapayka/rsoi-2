FROM golang

COPY ./ /backend
WORKDIR /backend
RUN go mod tidy && go mod vendor && go build -o ./api_gateway/ api_gateway/main.go

ENTRYPOINT ["/backend/api_gateway/main"]