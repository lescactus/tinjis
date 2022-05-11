# Preface

We're really happy that you're considering to join us! Here's a challenge that will help us understand your skills and serve as a starting discussion point for the interview.

We're not expecting that everything will be done perfectly as we value your time. You're encouraged to point out possible improvements during the interview though!

Have fun!

## The challenge

Pleo runs most of its infrastructure in Kubernetes. It's a bunch of microservices talking to each other and performing various tasks like verifying card transactions, moving money around, paying invoices ...

We would like to see that you both:
- Know how to create a small microservice
- Know how to wire it together with other services running in Kubernetes

We're providing you with a small service (Antaeus) written in Kotlin that's used to charge a monthly subscription to our customers. The trick is, this service needs to call an external payment provider to make a charge and this is where you come in.

You're expected to create a small payment microservice that Antaeus can call to pay the invoices. You can use the language of your choice. Your service should randomly succeed/fail to pay the invoice.

On top of that, we would like to see Kubernetes scripts for deploying both Antaeus and your service into the cluster. This is how we will test that the solution works.

## Instructions

Start by forking this repository. :)

1. Build and test Antaeus to make sure you know how the API works. We're providing a `docker-compose.yml` file that should help you run the app locally.
2. Create your own service that Antaeus will use to pay the invoices. Use the `PAYMENT_PROVIDER_ENDPOINT` env variable to point Antaeus to your service.
3. Your service will be called if you invoke `/rest/v1/invoices/pay` call on Antaeus. You can probably figure out which call returns the current status invoices by looking at the code ;)
4. Kubernetes: Provide deployment scripts for both Antaeus and your service. Don't forget about Service resources so we can call Antaeus from outside the cluster and check the results.
    - Bonus points if your scripts use liveness/readiness probes.
5. **Discussion bonus points:** Use the README file to discuss how this setup could be improved for production environments. We're especially interested in:
    1. How would a new deployment look like for these services? What kind of tools would you use?
    2. If a developers needs to push updates to just one of the services, how can we grant that permission without allowing the same developer to deploy any other services running in K8s?
    3. How do we prevent other services running in the cluster to talk to your service. Only Antaeus should be able to do it.

## How to run

If you want to run Antaeus locally, we've prepared a docker compose file that should help you do it. Just run:
```
docker-compose up
```
and the app should build and start running (after a few minutes when gradle does its job)

## How we'll test the solution

1. We will use your scripts to deploy both services to our Kubernetes cluster.
2. Run the pay endpoint on Antaeus to try and pay the invoices using your service.
3. Fetch all the invoices from Antaeus and confirm that roughly 50% (remember, your app should randomly fail on some of the invoices) of them will have status "PAID".

---
## Solution

To increase readability and avoid mixing the code of Antaeus and the 3rd party payment provider (very originally named `payment`), the repository has been rearranged with the following structure:

```
.
├── antaeus                     // Antaeus service related code 
│   ├── deploy                  // Kubernetes manifests for antaeus
│   ├── gradle
│   ├── pleo-antaeus-xxx
│   ├── ...
│   ├── docker-compose.yml      // Docker compose file for antaeus
│   ├── Dockerfile              // Dockerfile for antaeus
|   └── ...
├── payment                     // Payment service related code
│   ├── deploy                  // Kubernetes manifests for payment
│   ├── docker-compose.yml      // Docker compose file for payment
│   ├── Dockerfile              // Dockerfile for payment
│   ├── main.go                 // Payment source code
|   └── ...
├── docker-compose.yml          // Docker compose file for both antaeus and payment
└── README.md
```

### Building `payment`

<details>
<summary>Click to expand</summary>

#### From source with go

You need a working [go](https://golang.org/doc/install) toolchain (It has been developped and tested with go 1.16 and go 1.16 only, but should work with go >= 1.14). Refer to the official documentation for more information (or from your Linux/Mac/Windows distribution documentation to install it from your favorite package manager).

```sh
cd payment/

# Build from sources. Use the '-o' flag to change the compiled binary name
go build

# Default compiled binary is payment
# You can optionnaly move it somewhere in your $PATH to access it shell wide
./payment
```

#### From source with docker

If you don't have [go](https://golang.org/) installed but have docker, run the following command to build inside a docker container:

```sh
# Build from sources inside a docker container. Use the '-o' flag to change the compiled binary name
# Warning: the compiled binary belongs to root:root
docker run --rm -it -v "$PWD":/app -w /app golang:1.16 go build

# Default compiled binary is payment
# You can optionnaly move it somewhere in your $PATH to access it shell wide
./payment
```

#### From source with docker but built inside a docker image

If you don't want to pollute your computer with another program, `payment` comes with its own docker image:

```sh
docker build -t payment .

docker run --rm -p 8080:8080 payment
```

#### Unit tests

To run the test suite, run the following commands:

```sh
# Run the unit tests. Remove the '-v' flag to reduce verbosity
go test -v ./... 

# Get coverage to html format
go test -coverprofile -v /tmp/cover.out ./...
go tool cover -html=/tmp/cover.out -o /tmp/cover.out.html
```

</details>

### Kubernetes deployment

Ensure you have a properly working and accessible Kubernetes cluster with a valid `~/.kube/config`. 

To deploy `antaeus` and `payment`, run:

```sh
kubectl create -f antaeus/deploy/namespace.yaml
kubectl create -f antaeus/deploy/serviceaccount.yaml
kubectl create -f antaeus/deploy/deployment.yaml
kubectl create -f antaeus/deploy/service.yaml

kubectl create -f payment/deploy/namespace.yaml
kubectl create -f payment/deploy/serviceaccount.yaml
kubectl create -f payment/deploy/deployment.yaml
kubectl create -f payment/deploy/service.yaml
```

This will create the following resources:

* A `antaeus` namespace containing the `antaeus` deployment and its `LoadBalancer` service.

* A `payment` namespace containing the `payment` deployment and its `ClusterIP` service.

The docker images used have been built locally using the code found in this repository and pushed manually to Docker hub. See the Improvements area section to read more about what can be improved.

#### Accessing `antaeus` REST API

Since the service associated with `antaeus` is of type [`LoadBalancer`](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer), it is possible to reach it using its external ingress IP. To curl it, run the following command:

```sh
# Typically on AWS
hostname="$(kubectl get svc -n antaeus antaeus -ojsonpath='{.status.loadBalancer.ingress[0].hostname}')"
curl "http://${hostname}:80/rest/v1/xxx"


# Typically on Minikube or GCP
ip="$(kubectl get svc -n antaeus antaeus -ojsonpath='{.status.loadBalancer.ingress[0].ip}')"
curl "http://${ip}:80/rest/v1/xxx"
```

### Discussions

> How would a new deployment look like for these services? What kind of tools would you use?

* The first thing would be to separate each service in its own git repository.
* Create a CI pipeline running unit tests and building the docker image and push it to a docker registry using [semantic versioning](https://semver.org/) compliant tags.
* Create a CD pipeline that automates the deployment of the new image and run a test suite: integration tests, non-regression tests, stress tests etc ..., with automated logs and metrics analysis to detect anormalities and do environment promotion. [Flagger](https://flagger.app/) is good for that, [Harness](https://harness.io/) too.
* To deploy in Kubernetes, a [Helm chart](https://helm.sh/) or a [Kustomization](https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/) would come handy, especially when managing multiple environments.
* Follow the GitOps principle with tools such as the amazing [FluxCD](https://fluxcd.io/) or [ArgoCD](https://argoproj.github.io/).
* The usage of a `LoadBalancer` service for a single deployment should be discouraged in a production environment. Instead, an [`Ingress`](https://kubernetes.io/docs/concepts/services-networking/ingress/) would be better. Ideally, at the edge should stand an API Gateway or cloud Load Balancer doing TLS termination and redirection, authn/authz (could also be done via sidecar proxy), WAF, audit logs, request validation, etc ...


> If a developers needs to push updates to just one of the services, how can we grant that permission without allowing the same developer to deploy any other services running in K8s?

Since we are talking about **production** environments, the answer is that developers must not push updates manually in the cluster. It is the sign of a dysfunctional automated environment. However, if for *reasons* it had to happen, [RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) is the answer:

* Developers of service `antaeus` will be granted a role to read/write the `antaeus` service but not the `payment` one.

* Developers of service `payment` will be granted a role to read/write the `payment` service but not the `antaeus` one.

It could looks like:

```yaml
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: antaeusdev-rw
  namespace: antaeus
rules:
- apiGroups: [""]
  resources: ["*"]
  verbs: ["get","list","watch","create","update","patch","delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: antaeusdev-rw
  namespace: antaeus
subjects:
- kind: User
  name: antaeusdev
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: antaeusdev-rw
  apiGroup: rbac.authorization.k8s.io
```

> How do we prevent other services running in the cluster to talk to your service. Only Antaeus should be able to do it.

One of the answer is by using [Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/#networkpolicy-resource) for a L4 restrictions. It requires using a CNI implementing `NetworkPolicy`, otherwise it has no effect.

Another answer could be using a Service Mesh such as [Istio](https://istio.io/) or [Linkerd](https://linkerd.io) which provides authn/authz at the L7.

### Improvements area

<details>
<summary>Click to expand</summary>

#### `payment`

`payment` is a very minimalistic service far from being "production ready" (and it could even be written with less lines of code). Some area of improvements might include but not limited to:

* Avoid using `io.ReadAll` in `InvoiceHandler` as it loads all the request body into memory, leading to a possible memory exhaustion (and possible Denial Of Service). Better use a `bytes.Buffer` or `io.Copy` instead

* Provide structured (json) logging to stdout with tiered log levels (fatal, error, warn, info, debug, trace) and without sensitive information

* Provide metrics

* Provide APM using a tracing library such as OpenTelemetry

* Provide alerts and monitoring (Grafana) dashboards based on above-mentioned metrics and APM

* Provide incident response checklist

* Provide accurate cpu and memory requests/limits based on stress-test benchmarks

* Use of a [PodDistruptionBudget](https://kubernetes.io/docs/tasks/run-application/configure-pdb/) and scale to multiple replicas with the use of anti-affinity rules to spread accross multiple availability zones

* Be [12factor](https://12factor.net/) compliant by reading configuration at runtime from config maps or secrets (env variables or config files) - such as "log level", "tcp port", etc ...

* Provide a Swagger endpoint

* Support graceful shutdowns for interrupt signals (SIGTERM)

* Ensure the service is stateless by using an external store provider (SQL, blob, NoSQL, k/v, etc...)

* Authenticate API calls (authentication/authorization) or even better, do it either at the edge (via an API gateway for instance) or in a sidecar proxy

* Do not use the `latest` docker tag ([never](https://stevelasker.blog/2018/03/01/docker-tagging-best-practices-for-tagging-and-versioning-docker-images/)). Instead, provide [semantic versioning](https://semver.org/)

* Implement retries policies or circuit-breaker (could ideally also by done by a Service Mesh)

#### `antaeus`

* Antaeus doesn't serialize properly the `Currency` enum class. The json payload sent to the `payment` service looks like: `{"currency":{},"customer_id":1,"value":301.99}`. Is it a bug or a feature ?

* Antaeus container is running as `root` user which is a massive security risk. A potential mitigation is adding an user from the Dockerfile and swich to it with the `USER` instruction.

</details>