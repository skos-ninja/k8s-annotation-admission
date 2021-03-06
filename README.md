# k8s-annotation-admission
[![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/skos/k8s-annotation-admission)](https://hub.docker.com/r/skos/k8s-annotation-admission)

This application allows you to perform regex validation on k8s resources to ensure that they have the required annotations and that the annotations match the regex provided. This uses the [validation webhook admission controller](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/) to perform these checks.

Annotations are provided to the application via a configuration parameter called `annotations` that is a key value pair of annotation name to regex value. e.g. in yaml `test.skos.ninja: .*`.

This application is available as a docker image from [dockerhub](https://hub.docker.com/r/skos/k8s-annotation-admission).

## Known limitations
- Currently the warning mode of the validation hook requires you to have at least 1.19 in both the server and the client in order for the warnings to be displayed.
- There's no way to limit annotations to specific resources. It's intended for you to deploy a new instance of the application per resource and then to add the hook to the existing `ValidatingWebhookConfiguration`.
- The hook allows you to run in non TLS mode however k8s mandates that all hooks must use TLS in order to be called.

## Configuration
The configuration options for the application are as follows:
```
  -a, --annotations map[string]string   Specify annotations (default "")
  -p, --port int             Specify port to run server on (default 8080)
  -c, --tls-cert string      Specify TLS certificate path
  -k, --tls-key string       Specify TLS key path
  -w, --warning              Only warn on a failed validation
```

These can be provided both as a flag for the command or via the yaml configuration that accepts the long form of the flag as the key.

## Usage
Below is an example deployment of the annotation validator that will ensure that all `Deployment` resources in the `test` namespace have any value in the annotation `test.skos.ninja`.

Deployment:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: k8s-annotation-validator
spec:
  selector:
    app: k8s-annotation-validator
  ports:
    - protocol: TCP
      port: 443
      targetPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8s-annotation-validator
  labels:
    app: k8s-annotation-validator
spec:
  selector:
    matchLabels:
      app: k8s-annotation-validator
  template:
    metadata:
      labels:
        app: k8s-annotation-validator
    spec:
      containers:
        - name: k8s-annotation-validator
          image: skos/k8s-annotation-admission:latest
          ports:
            - containerPort: 8080
          volumeMounts:
            - name: config-volume
              mountPath: /app/config/
      volumes:
        - name: config-volume
          configMap:
            name: k8s-annotation-validator
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: k8s-annotation-validator
data:
  tls.cert: |
    -----BEGIN CERTIFICATE-----
    <public cert>
    -----END CERTIFICATE-----
  tls.key: |
    -----BEGIN PRIVATE KEY-----
    <private key>
    -----END PRIVATE KEY-----
  config.yaml: |
    tls-cert: /app/config/tls.cert
    tls-key: /app/config/tls.key
    annotations:
      test.skos.ninja: .*
```

Validation webhook configuration:
```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: k8s-annotation-validator
webhooks:
- name: k8s-annotation-validator.default.svc
  sideEffects: None
  admissionReviewVersions: ["v1"]
  clientConfig:
    caBundle: "Ci0tLS0tQk...<base64-encoded PEM bundle containing the CA that signed the webhook's serving certificate>...tLS0K"
    service:
      namespace: default
      name: k8s-annotation-validator
      path: /validate
      port: 443
  namespaceSelector:
    matchExpressions:
      - key: annotations.skos.ninja/validate
        operator: In
        values: ["required"]
  rules:
    - operations: ["CREATE", "UPDATE"]
      apiGroups: ["apps"]
      apiVersions: ["v1"]
      resources: ["deployments"]
      scope: "Namespaced"
```

Example namespace annotation:
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: test
  labels:
    annotations.skos.ninja/validate: required
```
