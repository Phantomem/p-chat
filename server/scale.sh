
kubectl scale deployment/my-app --replicas=5

kubectl autoscale deployment/my-app --min=2 --max=10 --cpu-percent=80