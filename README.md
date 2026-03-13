# JSON Server Controller

> A Kubernetes operator that deploys an opinionated JSON Server using a Custom Resource.

---

## Overview

The JSON Server Controller watches for `JSonServer` custom resources in your cluster and automatically provisions a fully working REST API — no manual configuration required. Each resource you create results in three Kubernetes objects being created on your behalf:

| Resource     | Purpose                                              |
| ------------ | ---------------------------------------------------- |
| `ConfigMap`  | Holds the `db.json` data file consumed by the server |
| `Deployment` | Runs the json-server container                       |
| `Service`    | Exposes the REST API endpoint inside the cluster     |

The shape of your JSON directly determines the API routes that are available. Define your data, apply the resource, and the API is ready.

---

## How It Works

Given a JSON configuration like this:

```json
{
  "people": [
    { "id": 1, "name": "Person A" },
    { "id": 2, "name": "Person B" }
  ]
}
```

The controller automatically produces the following REST endpoints:

```
GET    /people
GET    /people/1
POST   /people
PUT    /people/1
DELETE /people/1
```

Each top-level key in your JSON becomes a resource collection, with full CRUD support out of the box.

---

## Architecture

A `JSonServer` resource moves through the following pipeline before the API becomes available:

```
JSonServer CR
     │
     ▼
Admission Webhook          ← validates naming conventions
     │
     ▼
Controller                 ← reconciliation loop
     │
     ▼
Service Layer              ← orchestration
     │
     ▼
Ensurer
     ├─ ConfigMap  (db.json)
     ├─ Deployment (json-server container)
     └─ Service    (REST API endpoint)
```

---

## Admission Webhook

Before any resource reaches the controller, it passes through a **Validating Admission Webhook**. This enforces a simple naming convention: every `JSonServer` name must begin with `app-`.

**Valid:**

```yaml
metadata:
  name: app-my-server
```

**Invalid — rejected before reaching the controller:**

```yaml
metadata:
  name: my-server
```

This convention makes API resources immediately identifiable in the cluster and prevents naming collisions.

### TLS & Certificate Management

Admission webhooks are required to communicate over HTTPS. Certificates are handled automatically by **cert-manager**, so there is nothing to configure manually.

The following objects are created at deploy time:

- `Issuer` — a self-signed cert-manager issuer
- `Certificate` — issues the webhook TLS certificate
- `Secret` (`webhook-server-cert`) — stores the certificate for mounting

**Certificate flow:**

```
cert-manager → Issuer → Certificate → Secret → Mounted into controller → Webhook TLS (:9443)
```

The certificate is mounted inside the controller pod at:

```
/tmp/json-server-controller-webhook-server/serving-certs
```

**Verify the webhook is registered:**

```bash
kubectl get validatingwebhookconfigurations
```

**Verify the certificate was issued:**

```bash
kubectl get certificate.cert-manager.io -n json-server-controller-system
```

---

## CI & Deployment

### Continuous Integration

The project includes a minimal CI pipeline. On every push, the following stages run in sequence:

```
fmt → vet → build → docker image
```

Built images are pushed to `ttl.sh` for short-lived, ephemeral distribution. The image tag includes the commit SHA and an expiry suffix:

```
ttl.sh/json-server-controller:<commit-sha>-1h
```

The `-1h` suffix means the image is automatically deleted after one hour — suitable for demos and testing, not long-term storage.

### ArgoCD Deployment

The controller is deployed using **Argo CD**. An `Application` manifest is included in the repository and points directly at the `config/default` path:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: json-server-controller
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/ntlaletsi70/json-server-controller
    targetRevision: main
    path: config/default
  destination:
    server: https://kubernetes.default.svc
    namespace: json-server-controller-system
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

To install, a single command is all that is needed:

```bash
kubectl apply -f deploy/controller_manager.yaml
```

Once applied, Argo CD will watch the repository, sync the manifests, and deploy the controller. With `automated` sync enabled, every new commit to `main` will trigger an update automatically.

### Deployment Flow

```
Developer
    │
    ▼
git push
    │
    ▼
GitHub Actions
(fmt → vet → build → push image to ttl.sh)
    │
    ▼
Argo CD
(watches repository, syncs on change)
    │
    ▼
Kubernetes Cluster
    │
    ▼
JSON Server Controller
```

---

## Setup

### Prerequisites

Before getting started, ensure the following are available in your environment:

| Requirement        | Notes                            |
| ------------------ | -------------------------------- |
| Go 1.23+           | Required to build the controller |
| Docker or Podman   | For building and pushing images  |
| `kubectl`          | Cluster access                   |
| Kubernetes cluster | Local or remote                  |
| cert-manager       | Must be installed in the cluster |
| Mage               | Build task runner                |

### Installing Mage

```bash
go install github.com/magefile/mage@latest
```

### Available Build Tasks

To list all available tasks:

```bash
mage -l
```

| Command              | Description                                               |
| -------------------- | --------------------------------------------------------- |
| `mage ensureTools`   | Install required binaries (`controller-gen`, `kustomize`) |
| `mage manifests`     | Generate CRD manifests                                    |
| `mage generate`      | Generate deepcopy code                                    |
| `mage fmt`           | Format code                                               |
| `mage vet`           | Vet code                                                  |
| `mage build:manager` | Build the controller binary                               |
| `mage dev:run`       | Run the controller locally against a live cluster         |

---

## Quick Start

The following steps take you from a fresh clone to a running API.

**1. Clone the repository**

```bash
git clone https://github.com/ntlaletsi70/json-server-controller
cd json-server-controller
```

**2. Install build tools**

Ensures required binaries such as `controller-gen` and `kustomize` are present:

```bash
mage ensureTools
```

**3. Build the controller**

```bash
mage build:manager
```

**4. Build and deploy the controller image**

```bash
IMG=ttl.sh/json-server-controller:demo mage build:image
IMG=ttl.sh/json-server-controller:demo mage deploy
```

**5. Create the example API**

```bash
kubectl apply -k config/samples/
```

**6. Forward the service port**

```bash
kubectl port-forward svc/app-my-server 3000:80
```

**7. Test the endpoint**

```bash
curl http://localhost:3000/people
```

Expected response:

```json
[
  { "id": 1, "name": "Person A" },
  { "id": 2, "name": "Person B" }
]
```

---

## Scenarios

The following scenarios are useful for verifying controller behaviour end-to-end.

---

### 1 — Resource Creation

Apply the sample resource and confirm all three dependent objects are created:

```bash
kubectl apply -k config/samples/
```

```bash
kubectl get configmap
kubectl get deployment
kubectl get svc
```

All three should appear under the name `app-my-server`.

---

### 2 — API Access

Forward the service and query the endpoint:

```bash
kubectl port-forward svc/app-my-server 3000:80
curl http://localhost:3000/people
```

---

### 3 — Multiple APIs

Each `JSonServer` resource runs its own isolated API instance. Create a second resource with a different JSON structure:

```json
{
  "products": [
    { "id": 1, "name": "Laptop" },
    { "id": 2, "name": "Phone" }
  ]
}
```

This produces a separate set of endpoints:

```
/products
/products/1
```

The two API instances run independently and do not share state.

---

### 4 — Scaling

The underlying `Deployment` can be scaled normally:

```bash
kubectl scale jsonserver app-my-server \
  --current-replicas=2 \
  --replicas=3
```

Confirm the replica count:

```bash
kubectl get deployment app-my-server
```

Expected: **3 replicas** running.

---

### 5 — Admission Validation

Attempt to create a resource that violates the naming convention:

```yaml
metadata:
  name: my-server
```

```bash
kubectl apply -f invalid.yaml
```

The request will be rejected by the webhook before it reaches the controller:

```
metadata.name must start with "app-"
```

---

### 6 — JSON Validation

If the `jsonConfig` field contains malformed JSON, the controller surfaces the error in the resource status:

```yaml
jsonConfig: |
  {
    "people": [
      { "id": 1, "name": "Person A"
    ]
  }
```

Inspect the resource status:

```bash
kubectl get jsonserver app-my-server -o yaml
```

Look for the `status.state` and `status.message` fields, which will describe the validation failure.

---

### 7 — Delete API

Deleting the `JSonServer` resource triggers cleanup of all dependent objects:

```bash
kubectl delete jsonserver app-my-server
```

The controller removes the `ConfigMap`, `Deployment`, and `Service` automatically. No manual cleanup is required.

---

## Project Structure

```
api/        CRD type definitions
cmd/        Controller entrypoint
config/     Kubernetes manifests and samples
internal/   Core controller logic
pkg/        Service layer and resource orchestration
```

### Key Components

| Path          | Role                                                            |
| ------------- | --------------------------------------------------------------- |
| `controller/` | Reconciliation loop — watches and reacts to resource changes    |
| `service/`    | Orchestration layer — coordinates resource creation and updates |
| `ensurer/`    | Creates and manages the ConfigMap, Deployment, and Service      |
| `api/`        | CRD schema definitions                                          |
