# Start from the official Golang base image
# FROM golang:1.23.4

# Set the Current Working Directory inside the container
# WORKDIR /app

# Copy go mod and sum files
# COPY go.mod go.sum ./
# COPY go.mod ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
# RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
# COPY . .

# Build the Go app
# RUN go build -o mainfile . || echo "Build failed"
# RUN CGO_ENABLED=0 GOOS=linux go build -o backendMain


# Expose port 8081 to the outside world
# EXPOSE 8081
# RUN chmod +x mainfile
# Command to run the executable
# CMD ["/app/backendMain"]
# CMD ["ls", ".", "-la" , "--recursive"]



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


FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY static /static
COPY templates /templates
COPY --from=build-stage /backend /backend

EXPOSE 8081

USER nonroot:nonroot

ENTRYPOINT ["/backend"]
