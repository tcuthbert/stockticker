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
SYMBOL=MSFT NDAYS=7 ./stockticker -api-keyfile ./apikey.txt
webserver: 2023/06/12 11:34:12 Server is ready to handle requests at: :5000
webserver: 2023/06/12 11:34:28 INFO: response time=1.131412083s
```

### Interacting with the API

```
❯ curl -s http://localhost:5000
{"Meta Data":{"1. Information":"Daily Time Series with Splits and Dividend Events","2. Symbol":"MSFT","3. Last Refreshed":"2023-06-09","4. Output Size":"Compact","5. Time Zone":"US/Eastern"},"Time Series (Daily)":{"2023-06-01":{"1. open":"325.93","2. high":"333.53","3. low":"324.72","4. close":"332.58","6. volume":"26773851"},"2023-06-02":{"1. open":"334.247","2. high":"337.5","3. low":"332.55","4. close":"335.4","6. volume":"25873769"},"2023-06-05":{"1. open":"335.22","2. high":"338.5599","3. low":"334.6601","4. close":"335.94","6. volume":"21307053"},"2023-06-06":{"1. open":"335.33","2. high":"335.37","3. low":"332.17","4. close":"333.68","6. volume":"20396223"},"2023-06-07":{"1. open":"331.65","2. high":"334.49","3. low":"322.5","4. close":"323.38","6. volume":"40717129"},"2023-06-08":{"1. open":"323.935","2. high":"326.64","3. low":"323.35","4. close":"325.26","6. volume":"23277708"},"2023-06-09":{"1. open":"324.99","2. high":"329.99","3. low":"324.41","4. close":"326.79","6. volume":"22528950"}},"ClosingAverage":"330.43"}
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

##### Dump the k8s config for a given stage

```
: STAGE=(prod|dev)
make k8s-kustomize-${STAGE:=dev}
```

##### Deploy stage into k8s cluster

```
: STAGE=(prod|dev)
make k8s-deploy-${STAGE:=dev}

kubectl create -k kubernetes/dev/
configmap/stockticker-configmap created
secret/stockticker-apikey-5h4bmcf57t created
service/stockticker-service created
deployment.apps/stockticker-deployment created
```

[^1]: I used microk8s for this project

[semver]: https://semver.org/
