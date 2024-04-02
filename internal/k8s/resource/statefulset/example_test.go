package statefulset

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

func ExampleStatefulSet() {
	ss, _ := NewStatefulSet("test", "sourcegraph")

	jss, _ := json.Marshal(ss)
	fmt.Println(string(jss))

	yss, _ := yaml.Marshal(ss)
	fmt.Println(string(yss))

	// Output:
	// {"metadata":{"name":"test","namespace":"sourcegraph","creationTimestamp":null,"labels":{"deploy":"sourcegraph"}},"spec":{"replicas":1,"selector":{"matchLabels":{"app":"test"}},"template":{"metadata":{"creationTimestamp":null},"spec":{"containers":null}},"serviceName":"test","podManagementPolicy":"OrderedReady","updateStrategy":{"type":"RollingUpdate"},"revisionHistoryLimit":10,"minReadySeconds":10},"status":{"replicas":0,"availableReplicas":0}}
	// metadata:
	//   creationTimestamp: null
	//   labels:
	//     deploy: sourcegraph
	//   name: test
	//   namespace: sourcegraph
	// spec:
	//   minReadySeconds: 10
	//   podManagementPolicy: OrderedReady
	//   replicas: 1
	//   revisionHistoryLimit: 10
	//   selector:
	//     matchLabels:
	//       app: test
	//   serviceName: test
	//   template:
	//     metadata:
	//       creationTimestamp: null
	//     spec:
	//       containers: null
	//   updateStrategy:
	//     type: RollingUpdate
	// status:
	//   availableReplicas: 0
	//   replicas: 0
}
