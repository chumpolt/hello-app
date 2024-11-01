#Deploying hello1 (Node.js) and hello2 (Go) Applications on Kubernetes

This guide covers how to deploy two applications, hello1 (Node.js) and hello2 (Go), on Kubernetes. It includes setting up the necessary Deployments, Services, and Ingress configurations to route traffic to each app using an NGINX Ingress controller.

## Prerequisites

  •	Kubernetes cluster (Docker Desktop, Minikube, or any Kubernetes setup)

  •	kubectl configured to interact with your cluster

  •	NGINX Ingress Controller installed

  •	Modify your /etc/hosts file to include:

  
```
127.0.0.1 hello-app.local

```
 •	Create namespace "hello-app" in Kubernetes

```sh
kubectl create namespace hello-app
```

## Step 1: Prepare Docker Images for hello1 and hello2

### 1.1.	Build and push hello1 (Node.js) image:


  •	hello1 is a Node.js application using Express to handle requests.
 
  •	Dockerfile:
 
 
``` dockerfile
# Dockerfile for hello-1

FROM node:16

# Set the working directory in the container
WORKDIR /app

# Copy package.json and package-lock.json to install dependencies
COPY package*.json ./

# Install dependencies
RUN npm install express redis

# Copy the application files
COPY . .

# Set default environment variables for Redis connection
# These will be overridden by Kubernetes config or secrets

ENV host=redis-1 
ENV port=6379
ENV password=password

# Expose the application port
EXPOSE 8000

# Command to run the app
CMD ["node", "hello1/hello-1.js"]
```

•	Build and push:

```sh
docker build -t yourusername/hello1-app:latest .
docker push yourusername/hello1-app:latest
```

### 1.2.	Build and push hello2 (Go) image:

•	hello2 is a Go application that listens for requests.
 
•	Dockerfile:

``` dockerfile
# Use the official Golang image to allow running code directly with `go run`
FROM golang:1.18

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY hello2/go.mod hello2/go.sum ./
RUN go mod download

# Copy the entire application source code
COPY hello2/ .

# Set environment variables for Redis
ENV REDIS_HOST="localhost:6379"
ENV REDIS_PASSWORD="password"
ENV REDIS_DB="0"

# Expose the application port if necessary
EXPOSE 8000

# Run the application using `go run`
CMD ["go", "run", "main.go"]
``` 


