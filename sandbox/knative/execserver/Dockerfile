FROM golang:1.13
WORKDIR /work
COPY go.* ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o server
