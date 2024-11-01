# Deploying hello1 (Node.js) and hello2 (Go) Applications on Kubernetes

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
### 2.3. Deploy Redis as a StatefulSet (redis-1)

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
        image: docker.io/youeusername/hello1-app:latest  # Replace with your username
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

### 2.5 Verify the Deployment and Services for hello1 

Check the status of your pods to confirm that they’re running:
```sh
kubectl get pods -n hello-app
kubectl get svc -n hello-app
```


### 2.6 Create a Persistent Volume and Persistent Volume Claim for redis-2

redis-2-pv.yaml

```yaml
# redis-2-pv.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: redis-2-pv
  namespace: hello-app
spec:
  capacity:
    storage: 1Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: "/mnt/data/redis-2"  # Change this path as necessary for your environment

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redis-2-pvc
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
kubectl apply -f redis-2-pv.yml
```

### 2.7. Deploy Redis as a StatefulSet (redis-2)

```yaml
# redis-2-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-2
  namespace: hello-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis-2
  template:
    metadata:
      labels:
        app: redis-2
    spec:
      containers:
      - name: redis
        image: redis:6.2
        ports:
        - containerPort: 6379
        env:
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-secret
              key: password
        args: ["--requirepass", "$(REDIS_PASSWORD)"]
        volumeMounts:
        - mountPath: /data
          name: redis-storage  # Mount the PVC at /data, the default data directory for Redis
      volumes:
      - name: redis-storage
        persistentVolumeClaim:
          claimName: redis-2-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: redis-2
  namespace: hello-app
spec:
  ports:
    - port: 6379
      targetPort: 6379
  selector:
    app: redis-2
  type: ClusterIP
```
• Apply the StatefulSet and Service:

```sh
kubectl apply -f redis-2-deployment.yaml
```

### 2.8 Deploy hello2 (go) Deployment and Service
Save the following YAML to a file named hello1-deployment.yml
```yaml
# hello2-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello2
  namespace: hello-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: hello2
  template:
    metadata:
      labels:
        app: hello2
    spec:
      imagePullSecrets:
        - name: regcred
      containers:     
      - name: hello2
        image: docker.io/chumpol01/hello2-app:latest  # Replace with the correct image path
        ports:
        - containerPort: 8000
        env:
        - name: REDIS_HOST
          value: "redis-2:6379"  # Updated to use redis-2 service
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: redis-secret
              key: password
        - name: REDIS_DB
          value: "0"
```
Apply the deployment:

```sh
kubectl apply -f hello-1-deployment.yaml

```
Save the following YAML to a file named hello-2-service.yml
```yaml
# hello2-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: hello2-service
  namespace: hello-app
spec:
  type: ClusterIP
  selector:
    app: hello2
  ports:
    - protocol: TCP
      port: 8000          # Service port for hello2
      targetPort: 8000 
```
Apply the services:

```sh
kubectl apply -f hello-2-service.yml
```
### 2.9 Verify the Deployment and Services for hello2

Check the status of your pods to confirm that they’re running:
```sh
kubectl get pods -n hello-app
kubectl get svc -n hello-app
```

## Step 3 Set Up Ingress for hello1 and hello2
The Ingress configuration routes requests to /hello1 and /hello2 to hello1-service and hello2-service, respectively.

Save the following YAML to a file named hello-app-ingress.yaml:
```yaml
# hello-app-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hello-app-ingress
  namespace: hello-app
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /$2
spec:
  ingressClassName: nginx
  rules:
    - host: hello-app.local   # Hostname for accessing the services (update for your local testing)
      http:
        paths:
          - path: /hello1(/|$)(.*)
            pathType: ImplementationSpecific
            backend:
              service:
                name: hello1-service
                port:
                  number: 8000
          - path: /hello2(/|$)(.*)
            pathType: ImplementationSpecific
            backend:
              service:
                name: hello2-service
                port:
                  number: 8000
```
Apply the Ingress configuration:

```yaml
kubectl apply -f hello-app-ingress.yaml
```

## Step 4 Test the Setup
4.1.	Ensure hello-app.local is mapped to 127.0.0.1 in your /etc/hosts file:
4.2.	Access the applications in your browser or using curl:

	•	hello1: http://hello-app.local/hello1/hello1
 
	•	hello2: http://hello-app.local/hello2/hello2
 
You should see responses indicating successful routing, such as:

	•	"Hello-1 from hello1 service" for /hello1/hello1
 
	•	"Hello-2 from hello2 service" for /hello2/hello2

4.3. View Console log in pod by
for hell01 app

```sh
kubectl logs -f <pod-hello-1-name> -n hello-app
```
Log detail show

```Plaintext
Redis client connected
Session ID: 35
Session ID: 36
Session ID: 38
```


for hello2 app
```sh
kubectl logs -f <pod-hello-2-name> -n hello-app
```

Log detail show

```Plaintext
⇨ http server started on [::]:8000
{"time":"2024-11-01T06:49:41.456978881Z","level":"INFO","prefix":"echo","file":"main.go","line":"35","message":"Session: 2024-11-01 06:49:41.453995506 +0000 UTC - 21"}
{"time":"2024-11-01T06:49:52.709038095Z","level":"INFO","prefix":"echo","file":"main.go","line":"35","message":"Session: 2024-11-01 06:49:52.707511637 +0000 UTC - 23"}
```

## Troubleshooting

• 404 Not Found: If you receive a 404, verify the IngressClass and Ingress configuration. Ensure the Ingress resource uses the nginx class and that paths match correctly.
• Ingress Controller Logs: Check the logs of the NGINX Ingress controller pod for routing or backend errors: 

```sh
kubectl logs -n ingress-nginx <nginx-ingress-pod-name>
```
 
