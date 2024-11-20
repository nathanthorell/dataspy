FROM golang:latest AS build
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -o dataspy .

# Final image build
FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates

# Copy the pre-built binary file from the previous stage
COPY --from=build /app/dataspy .
RUN chmod +x ./dataspy
CMD ["./dataspy"]
