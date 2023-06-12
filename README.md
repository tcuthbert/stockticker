# Stockticker

## Overview

### Usage

```
❯ ./stockticker -h
Usage of ./build/stockticker:
  -api-keyfile string
    	file containing key data for the API
  -api-url string
    	url of the stock ticker API (default "https://www.alphavantage.co")
  -listen-addr string
    	server listen address (default ":5000")
  -num-days int
    	last <num-days> of stock data (default 10)
  -symbol string
    	stock symbol to lookup (default "MSFT")
```

### Starting the server

```
SYMBOL=IBM NDAYS=7 ./stockticker -api-keyfile ./apikey.txt
webserver: 2023/06/12 11:34:12 Server is ready to handle requests at: :5000
webserver: 2023/06/12 11:34:28 INFO: response time=1.131412083s
```

### Interacting with the API

```
❯ curl -s http://localhost:5000
{"Meta Data":{"1. Information":"Daily Time Series with Splits and Dividend Events","2. Symbol":"IBM","3. Last Refreshed":"2023-06-09","4. Output Size":"Compact","5. Time Zone":"US/Eastern"},"Time Series (Daily)":{"2023-06-01":{"1. open":"128.44","2. high":"130.145","3. low":"127.78","4. close":"129.82","6. volume":"4136086"},"2023-06-02":{"1. open":"130.38","2. high":"133.12","3. low":"130.15","4. close":"132.42","6. volume":"5375796"},"2023-06-05":{"1. open":"133.12","2. high":"133.58","3. low":"132.27","4. close":"132.64","6. volume":"3993516"},"2023-06-06":{"1. open":"132.43","2. high":"132.94","3. low":"131.88","4. close":"132.69","6. volume":"3297951"},"2023-06-07":{"1. open":"132.5","2. high":"134.44","3. low":"132.19","4. close":"134.38","6. volume":"5772024"},"2023-06-08":{"1. open":"134.69","2. high":"135.98","3. low":"134.01","4. close":"134.41","6. volume":"4128939"},"2023-06-09":{"1. open":"134.36","2. high":"136.1","3. low":"134.17","4. close":"135.3","6. volume":"3981748"}},"ClosingAverage":"133.09"}
```

## Project goals

- Handling broken client connections using context timeouts
- Makefile driving common project related tasks
- Kubernetes kustomize base/overlay pattern for multi-stage deployments
- Automated docker image builds using GitHub workflow
- [Semantically versioned][semver] release management

## Future improvements

##### APIHandler refactor

- I opted for using a callback handler instead of a struct/interface for
  simplicity
- Given more time I would have looked into making the APIHandler
  satisfy the [http.Handler](https://pkg.go.dev/net/http#Handler) interface,
  which should reduce some of the boilerplate code

##### Testing

- The webserver needs more testing. I'm only checking for the basic case that
  it's returning status ok
- Ditto for the apiclient

##### Documentation

- I ran out of time and didn't write docstrings detailing each component like I normally would

## Getting started

##### Dependencies

- Go 1.20 or above, see [install](https://go.dev/doc/install) for details
- [microk8s](https://microk8s.io/) or [minikube](https://minikube.sigs.k8s.io/docs/start/) cluster [^1]
- make, git, [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/)

git and make installation on Ubuntu:

```
sudo apt update
sudo apt-get install -y git make
```

## Building

##### Binary

```
❯ make build run
GOARCH=amd64 GOOS=linux go build -ldflags "-X main.VERSION=v0.1.0-4-g66f1784 -X main.COMMIT=66f1784cf6a8ef01e7fe1821b7820d8b92555745 -X main.BRANCH=readme" -o /home/tom/Projects/stockticker//build/stockticker .
/home/tom/Projects/stockticker//build/stockticker
webserver: 2023/06/12 11:39:22 Server is ready to handle requests at: :5000
```

##### Docker container

```
❯ make build-docker
docker build --no-cache -t ghcr.io/tcuthbert/stockticker:1.0.0 .
```

###### Running the docker image

```
❯ make docker-run DOCKER_FLAGS=""
docker build  -t ghcr.io/tcuthbert/stockticker:1.1.0-1-g07490f7-dirty .
[+] Building 9.6s (20/20) FINISHED
docker run --env-file .env --rm  --restart no -p 5000:5000 ghcr.io/tcuthbert/stockticker:1.1.0-1-g07490f7-dirty
webserver: 2023/06/12 06:28:04 Server is ready to handle requests at: :5000
```

## Testing

```
❯ make test
go test ./...
?   	github.com/tcuthbert/stockticker	[no test files]
ok  	github.com/tcuthbert/stockticker/apiclient	0.003s
ok  	github.com/tcuthbert/stockticker/apiresponse	(cached)
ok  	github.com/tcuthbert/stockticker/webserver	1.006s
```

## Deployment

#### Dump the k8s config for a given stage

```
: STAGE=(prod|dev)
make k8s-kustomize-${STAGE:=dev}
```

#### Deploy stage into k8s cluster

```
: STAGE=(prod|dev)
make k8s-deploy-${STAGE:=dev}
kubectl create -k kubernetes/dev/
configmap/stockticker-configmap created
secret/stockticker-apikey-5h4bmcf57t created
service/stockticker-service created
deployment.apps/stockticker-deployment created
ingress.networking.k8s.io/stockticker-ingress created
```

##### Testing the k8s ingress

```
❯ curl -s -H 'Host: stockticker.com' http://localhost/
{"Meta Data":{"1. Information":"Daily...
```

[^1]: I used microk8s for this project

[semver]: https://semver.org/
