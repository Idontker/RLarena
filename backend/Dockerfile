FROM golang:1.23.4 AS build-stage

WORKDIR /app

# COPY go.mod go.sum ./
COPY go.mod ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /backend


# Run the tests in the container
# FROM build-stage AS run-test-stage
# RUN go test -v ./...

FROM gcr.io/distroless/base-debian11 AS build-image
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.23.0
RUN goose --version

FROM build-image AS build-release-stage

WORKDIR /

COPY static /static
COPY templates /templates
COPY --from=build-stage /backend /backend

EXPOSE 8081

USER nonroot:nonroot

ENTRYPOINT ["/backend"]
