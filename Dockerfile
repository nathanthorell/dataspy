FROM golang:latest AS build
WORKDIR /app
COPY src/ .

ENV GOARCH=amd64
RUN go build -o /app/dataspy .

# Final image build
FROM alpine:latest
WORKDIR /app

# Copy the pre-built binary file from the previous stage
COPY --from=build /app/dataspy .

CMD ["./dataspy"]
