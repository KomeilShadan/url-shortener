FROM golang:alpine

# Create a non-root user
RUN adduser --disabled-password --gecos '' komeil

# Set the Current Working Directory inside the container
WORKDIR /app

# Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go clean --modcache
# Copy the source code from the current directory to the Working Directory inside the container
COPY . .

ENV GOCACHE=/root/.cache/go-build
RUN --mount=type=cache,target="/root/.cache/go-build" go build -o main ./cmd

RUN chown -R komeil:komeil /app

USER komeil

EXPOSE 8083

# Command to run the executable
CMD ["./main"]

