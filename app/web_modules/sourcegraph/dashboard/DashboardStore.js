import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";

export class DashboardStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
	}

	reset() {
		let exampleRepos = [{URI: "github.com/docker/docker", Owner: "docker", Name: "docker", Language: "Go", Description: "Docker - the open-source application container engine http://www.docker.com"},
			{URI: "github.com/drone/drone", Owner: "drone", Name: "drone", Language: "Go", Description: "Drone is a Continuous Delivery platform built on Docker, written in Go"},
			{URI: "github.com/golang/go", Owner: "golang", Name: "go", Language: "Go", Description: "The Go programming language https://golang.org"},
			{URI: "github.com/influxdata/influxdb", Owner: "influxdata", Name: "influxdb", Language: "Go", Description: "Scalable datastore for metrics, events, and real-time analytics https://influxdata.com"},
			{URI: "github.com/gorilla/mux", Owner: "gorilla", Name: "mux", Language: "Go", Description: "A powerful URL router and dispatcher for golang. http://www.gorillatoolkit.org/pkg/mux"}];

		if (typeof window !== "undefined") { // TODO(autotest) support document object.
			this.repos = deepFreeze((window.repos || []).concat(window.remoteRepos || []));
			this.onboarding = deepFreeze(window.onboarding);
			if (this.onboarding && this.onboarding.linkGitHub) {
				this.repos = deepFreeze(exampleRepos);
			}
		} else {
			this.repos = [];
			this.onboarding = {};
		}
	}

	__onDispatch(action) {
		switch (action.constructor) {

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new DashboardStore(Dispatcher.Stores);
