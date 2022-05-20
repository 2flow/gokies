## Build
```shell
docker build -t container-name:latest .
```

## Run
```shell
docker run --publish 80:80 --name runner-name container-name:latest
```

## Deoploy
```
docker build -t container-name:latest .
az acr login --name kiska
docker tag container-name:latest kiska.azurecr.io/container-name:latest
docker push kiska.azurecr.io/container-name:latest
```