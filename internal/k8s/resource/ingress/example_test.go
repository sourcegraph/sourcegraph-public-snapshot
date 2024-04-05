package ingress

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

func ExampleIngress() {
	i, _ := NewIngress("test", "sourcegraph")

	ji, _ := json.Marshal(i)
	fmt.Println(string(ji))

	yc, _ := yaml.Marshal(i)
	fmt.Println(string(yc))

	// Output:
	// {"metadata":{"name":"test","namespace":"sourcegraph","creationTimestamp":null,"labels":{"deploy":"sourcegraph"}},"spec":{},"status":{"loadBalancer":{}}}
	// metadata:
	//   creationTimestamp: null
	//   labels:
	//     deploy: sourcegraph
	//   name: test
	//   namespace: sourcegraph
	// spec: {}
	// status:
	//   loadBalancer: {}
}
