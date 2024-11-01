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

•	Build and push:

```sh
docker build -t yourusername/hello2-app:latest .
docker push yourusername/hello2-app:latest
```

## Step 2: Deploy hello1 and hello2 Applications on Kubernetes

### 2.1 Create a Persistent Volume and Persistent Volume Claim for redis-1
To ensure Redis data persists across pod restarts, create a Persistent Volume (PV) and Persistent Volume Claim (PVC) for storage.

• Save to file "redis-1-pv.yml

```yaml
# redis-1-pv.ymlapiVersion: v1
kind: PersistentVolume
metadata:
  name: redis-1-pv
spec:
  capacity:
    storage: 1Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: "/mnt/data/redis-1"  
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redis-1-pvc
  namespace: hello-app
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
```


• Apply file

```yaml
kubectl apply -f redis-pv.yml
```
### 2.2  Create a Secret for Redis Password 
If Redis is configured with authentication, create a Kubernetes Secret to store the password.

```yaml
# redis-secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: redis-secret
  namespace: hello-app
type: Opaque
data:
  password: <base64_encoded_password>  # Replace with the base64-encoded Redis password
 
```
• Apply the secret:

```yaml
kubectl apply -f redis-secret.yaml
```
### 2.3. Deploy Redis as a StatefulSet

A StatefulSet ensures stable network identity and persistent storage, ideal for a database like Redis

```yaml
# redis-1-deployment.yml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-1
  namespace: hello-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis-1
  template:
    metadata:
      labels:
        app: redis-1
    spec:
      containers:
      - name: redis
        image: redis:6.2
        env:
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-secret
              key: password
        args: ["--requirepass", "$(REDIS_PASSWORD)"] 
        ports:
        - containerPort: 6379
          name: redis
        volumeMounts:
        - mountPath: /data
          name: redis-storage
      volumes:
      - name: redis-storage
        persistentVolumeClaim:
          claimName: redis-1-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: redis-1
  namespace: hello-app
spec:
  ports:
  - port: 6379
    targetPort: 6379
  selector:
    app: redis-1
  type: ClusterIP
```

• Apply the StatefulSet and Service:

```sh
kubectl apply -f redis-1-deployment.yaml
```
### 2.4. Deploy hello1 Deployment and Service
Save the following YAML to a file named hello1-deployment.yml
```yaml
# hello-1-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello1
  namespace: hello-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: hello1
  template:
    metadata:
      labels:
        app: hello1
    spec:
      imagePullSecrets:
        - name: regcred
      containers:
      - name: hello1
        image: docker.io/chumpol01/hello1-app:latest  
        ports:
        - containerPort: 8000
        env:
        - name: host
          value: "redis-1"  # Redis service name
        - name: port
          value: "6379"
        - name: password
          valueFrom:
            secretKeyRef:
              name: redis-secret
              key: password
```
Apply the deployment:

```sh
kubectl apply -f hello-1-deployment.yaml
```

Save the following YAML to a file named hello-1-service.yml
```yaml
# hello-1-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: hello1-service
  namespace: hello-app
spec:
  type: ClusterIP  # Change to NodePort or LoadBalancer if needed
  selector:
    app: hello1
  ports:
    - protocol: TCP
      port: 8000
      targetPort: 8000

```
Apply the services:

```sh
kubectl apply -f hello-1-service.yml
```
