Deploying hello1 (Node.js) and hello2 (Go) Applications on Kubernetes

This guide covers how to deploy two applications, hello1 (Node.js) and hello2 (Go), on Kubernetes. It includes setting up the necessary Deployments, Services, and Ingress configurations to route traffic to each app using an NGINX Ingress controller.

Prerequisites

	•	Kubernetes cluster (Docker Desktop, Minikube, or any Kubernetes setup)
	•	kubectl configured to interact with your cluster
	•	NGINX Ingress Controller installed
	•	Modify your /etc/hosts file to include:
