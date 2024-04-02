package deployment

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

func ExampleDeployment() {
	d, _ := NewDeployment("test", "sourcegraph")

	jd, _ := json.Marshal(d)
	fmt.Println(string(jd))

	yd, _ := yaml.Marshal(d)
	fmt.Println(string(yd))

	// Output:
	// {"metadata":{"name":"test","namespace":"sourcegraph","creationTimestamp":null,"labels":{"deploy":"sourcegraph"}},"spec":{"replicas":1,"selector":{"matchLabels":{"app":"test"}},"template":{"metadata":{"creationTimestamp":null},"spec":{"containers":null}},"strategy":{"type":"Recreate"},"minReadySeconds":10,"revisionHistoryLimit":10},"status":{}}
	// metadata:
	//   creationTimestamp: null
	//   labels:
	//     deploy: sourcegraph
	//   name: test
	//   namespace: sourcegraph
	// spec:
	//   minReadySeconds: 10
	//   replicas: 1
	//   revisionHistoryLimit: 10
	//   selector:
	//     matchLabels:
	//       app: test
	//   strategy:
	//     type: Recreate
	//   template:
	//     metadata:
	//       creationTimestamp: null
	//     spec:
	//       containers: null
	// status: {}
}
