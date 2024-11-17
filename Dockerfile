# Step 1: Use the official Golang image to build the Go app
FROM golang:1.23-alpine AS builder

# Step 2: Set the current working directory inside the container
WORKDIR /app

# Step 3: Copy the Go application code to the working directory
COPY . .

# Step 4: Download Go modules if your project uses them
RUN go mod tidy

# Step 5: Build the Go app
RUN go build -o bin/main ./cmd/main

# Step 6: Use a minimal image to run the Go app
FROM alpine:latest

# Step 7: Set the working directory again
WORKDIR /app

# Step 8: Copy the built Go binary from the builder stage
COPY --from=builder /app/bin/main .
COPY --from=builder /app/static ./static

# Step 9: Expose any port your app is using (optional)
EXPOSE 8080

# Step 10: Command to run the Go app
CMD ["./main"]