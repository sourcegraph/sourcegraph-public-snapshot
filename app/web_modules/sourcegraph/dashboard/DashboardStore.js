import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";

import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

export class DashboardStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
	}

	toJSON() {
		return {
			repos: this.repos,
			remoteRepos: this.remoteRepos,
			hasLinkedGitHub: this.hasLinkedGitHub,
		};
	}

	reset(data) {
		this.exampleRepos = deepFreeze([{
			URI: "github.com/docker/docker",
			Owner: "docker", Name: "docker",
			Language: "Go",
			Description: "Docker - the open-source application container engine http://www.docker.com",
		},
		{
			URI: "github.com/drone/drone",
			Owner: "drone",
			Name: "drone",
			Language: "Go",
			Description: "Drone is a Continuous Delivery platform built on Docker, written in Go",
		},
		{
			URI: "github.com/golang/go",
			Owner: "golang",
			Name: "go",
			Language: "Go",
			Description: "The Go programming language https://golang.org",
		},
		{
			URI: "github.com/influxdata/influxdb",
			Owner: "influxdata",
			Name: "influxdb",
			Language: "Go",
			Description: "Scalable datastore for metrics, events, and real-time analytics https://influxdata.com",
		},
		{
			URI: "github.com/gorilla/mux",
			Owner: "gorilla",
			Name: "mux",
			Language: "Go",
			Description: "A powerful URL router and dispatcher for golang. http://www.gorillatoolkit.org/pkg/mux",
		}]);

		this.repos = data && data.repos ? data.repos : null;
		this.remoteRepos = data && data.remoteRepos ? data.remoteRepos : null;
		this.hasLinkedGitHub = data ? data.hasLinkedGitHub : null;
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case DashboardActions.ReposFetched:
			this.repos = deepFreeze((action.data && action.data.Repos) || []);
			break;

		case DashboardActions.RemoteReposFetched:
			this.remoteRepos = deepFreeze((action.data && action.data.RemoteRepos) || []);
			this.hasLinkedGitHub = action.data && action.data.HasLinkedGitHub;
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new DashboardStore(Dispatcher.Stores);
