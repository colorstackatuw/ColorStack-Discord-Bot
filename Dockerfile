# Use an official Python runtime as a parent image
FROM golang:1.22.4

# Set the working directory in the container to /app
WORKDIR /app

# Install git, required to clone the repository
RUN apt-get update && apt-get install -y git

# Copy the go.mod and go.sum files for container 
COPY go.mod go.sum ./  

# Run dependencies
RUN go mod download

# Copy the current directory contents into the container at /app
COPY . . 

# Build the Go application
RUN go build -o main . 

# Run Main the container launches
CMD ["./main"]
