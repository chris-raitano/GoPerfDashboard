# deploy ingress-nginx ingress controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.2.0/deploy/static/provider/cloud/deploy.yaml
# deploy app resources
kubectl apply -f ./testRunner/deploy.dev.yaml
kubectl apply -f ./server/deploy.dev.yaml
