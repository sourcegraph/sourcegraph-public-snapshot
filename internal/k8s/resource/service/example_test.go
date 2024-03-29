package service

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

func ExampleService() {
	s, _ := NewService("test", "sourcegraph")

	js, _ := json.Marshal(s)
	fmt.Println(string(js))

	ys, _ := yaml.Marshal(s)
	fmt.Println(string(ys))

	// Output:
	// {"metadata":{"name":"test","namespace":"sourcegraph","creationTimestamp":null,"labels":{"app":"test","deploy":"sourcegraph"}},"spec":{"selector":{"app":"test"},"type":"ClusterIP"},"status":{"loadBalancer":{}}}
	// metadata:
	//   creationTimestamp: null
	//   labels:
	//     app: test
	//     deploy: sourcegraph
	//   name: test
	//   namespace: sourcegraph
	// spec:
	//   selector:
	//     app: test
	//   type: ClusterIP
	// status:
	//   loadBalancer: {}
}
