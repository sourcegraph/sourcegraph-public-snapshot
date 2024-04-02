package pvc

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

func ExamplePersistentVolumeClaim() {
	p, _ := NewPersistentVolumeClaim("test", "sourcegraph")

	jp, _ := json.Marshal(p)
	fmt.Println(string(jp))

	yp, _ := yaml.Marshal(p)
	fmt.Println(string(yp))

	// Output:
	// {"metadata":{"name":"test","namespace":"sourcegraph","creationTimestamp":null,"labels":{"deploy":"sourcegraph"}},"spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"10Gi"}}},"status":{}}
	// metadata:
	//   creationTimestamp: null
	//   labels:
	//     deploy: sourcegraph
	//   name: test
	//   namespace: sourcegraph
	// spec:
	//   accessModes:
	//   - ReadWriteOnce
	//   resources:
	//     requests:
	//       storage: 10Gi
	// status: {}
}
