# VDDK Builder Project

## Overview
This project provides a Go-based HTTPS server that accepts file uploads, builds a container image using the uploaded file, and pushes it to a container registry. It also allows checking if an image already exists in the registry and supports overriding the default image name during upload.

## Prerequisites
- Go 1.19+
- Podman or Docker
- OpenShift CLI (oc)
- OpenSSL (for generating certificates)

## Commands
Run the following commands from the project root:

- **Build locally:** `make build-local`
- **Run a local registry:** `make run-registry`
- **Run locally:** `make run-local`
- **Build container image:** `make build-image`
- **Push container image:** `make push-image`
- **Deploy to OpenShift:** `make deploy`
- **Clean up:** `make clean`

## HTTPS Endpoints

### 1. **File Upload Endpoint**
Uploads a `.tar.gz` file to the server for building a container image.

**Endpoint:**
```http
POST /upload
```

**Parameters:**
- **Form Data:**
  - `file`: Path to the `.tar.gz` file to upload.
- **Query Parameters:**
  - `image` (optional): Override the default image name to push a custom image.

**Example Command:**
```bash
curl -k -F "file=@/path/to/file.tar.gz" "https://localhost:8443/upload?image=vddk-7"
```

If `image` is not provided, the default image name from the server configuration will be used.

### 2. **Check Image Endpoint**
Checks if a container image already exists in the configured registry.

**Endpoint:**
```http
GET /check-image
```

**Parameters:**
- **Query Parameters:**
  - `image`: The image name to check in the registry.

**Example Command:**
```bash
curl -k "https://localhost:8443/check-image?image=vddk-7"
```

**Responses:**
- `200 OK`: Image exists in the registry.
- `404 Not Found`: Image does not exist.
- `500 Internal Server Error`: Unexpected error during the check.

## Testing Locally
### Step 1: Run a Local Registry
Start a local container registry to push images:
```bash
make run-registry
```

### Step 2: Run the Server Locally
Build the server and start it locally with generated certificates:
```bash
make run-local
```

### Step 3: Upload a File
Upload a `.tar.gz` archive to trigger the image build:
```bash
curl -k -F "file=@/path/to/file.tar.gz" https://localhost:8443/upload
```

### Step 4: Check If an Image Exists
Verify if an image is already available in the registry:
```bash
curl -k "https://localhost:8443/check-image?image=vddk-7"
```

## Deployment on OpenShift
### Grant Permissions to Push Images
To allow the pod to push images to the OpenShift internal registry, grant the `system:image-builder` role to the service account:
```bash
oc policy add-role-to-user system:image-builder system:serviceaccount:<namespace>:default -n <namespace>
```
Replace `<namespace>` with the namespace where your pod is running.

### Deploy the Server
To deploy the server to an OpenShift cluster, run:
```bash
make deploy
```

Verify the deployment:
```bash
oc get pods
oc get services
```

### Expose the Service
To access the server from outside the cluster, expose the service using a route:
```bash
oc expose service vddk-builder-service --name=vddk-builder-route
```

Verify the route:
```bash
oc get route vddk-builder-route
```

The output will include the public URL, which can be used to call the service.

### Example Call from Outside the Cluster
1. **Upload a File**:
   ```bash
   curl -k -F "file=@/path/to/file.tar.gz" https://<route-url>/upload?image=vddk-7
   ```

2. **Check If an Image Exists**:
   ```bash
   curl -k "https://<route-url>/check-image?image=vddk-7"
   ```

Replace `<route-url>` with the URL from `oc get route`.

## Cleaning Up
- Stop and remove the local registry:
  ```bash
  make clean-registry
  ```
- Remove OpenShift resources:
  ```bash
  make clean
  ```

## Full Workflow
To build, push, and deploy the server, run:
```bash
make all
```
