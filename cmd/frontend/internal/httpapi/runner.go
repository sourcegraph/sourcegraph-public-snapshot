package httpapi

import (
	"net/http"
)

const config = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: sourcegraph-runner
  labels:
    k8s-app: sourcegraph-runner
spec:
  replicas: 1
  selector:
    matchLabels:
      k8s-app: sourcegraph-runner
  template:
    metadata:
      name: sourcegraph-runner
      labels:
        k8s-app: sourcegraph-runner
    spec:
      containers:
        - name: sourcegraph-runner
          image: sourcegraph/runner:1.0.0
          imagePullPolicy: IfNotPresent
          #resources:
          #  requests:
          #    cpu: '4'
          #    memory: 4Gi
          env:
            - name: DOCKER_HOST
              value: tcp://localhost:2375
            - name: MAX_CONCURRENT_TASKS
              value: '8'
            - name: SG_TOKEN
              valueFrom:
                secretKeyRef:
                  name: sg-token
                  key: token
            - name: SG_URL
              valueFrom:
                secretKeyRef:
                  name: sg-token
                  key: sourcegraphUrl
        - name: dind
          image: docker:18.06-dind
          # Needed, cannot run docker otherwise :|
          securityContext:
            privileged: true
          volumeMounts:
            - name: dind-storage
              mountPath: /var/lib/docker
      volumes:
        - name: dind-storage
          emptyDir: {}
`

func runnerKubeconfigServe(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/yaml")
	_, err := w.Write([]byte(config))
	return err
}
