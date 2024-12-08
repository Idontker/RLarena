# Build Stage
FROM golang:1.23.4 AS build-stage

RUN mkdir -p /app-data
WORKDIR /app
# Prepare your Go app
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /backend



# Final Stage
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /
# Copy static files and templates
COPY static /static
COPY templates /templates

# Copy the compiled binary
COPY --from=build-stage /backend /backend

# Ensure nonroot can access /app-data/*
COPY --from=build-stage /app-data /app-data

# Expose the application port
EXPOSE 8081

# TODO: 
# Run as a non-root user
# USER nonroot:nonroot

ENTRYPOINT ["/backend"]
