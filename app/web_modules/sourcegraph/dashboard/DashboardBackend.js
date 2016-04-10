import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";
import {trackPromise} from "sourcegraph/app/status";

const DashboardBackend = {
	fetch: defaultFetch,

	__onDispatch(action) {
		switch (action.constructor) {
		case DashboardActions.WantRepos:
			{
				let repos = DashboardStore.repos;
				if (repos === null) {
					trackPromise(
						DashboardBackend.fetch("/.api/repos")
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => Dispatcher.Stores.dispatch(new DashboardActions.ReposFetched(data)))
					);
				}
				break;
			}

		case DashboardActions.WantRemoteRepos:
			{
				let remoteRepos = DashboardStore.remoteRepos;
				if (remoteRepos === null) {
					trackPromise(
						DashboardBackend.fetch("/.api/remote-repos")
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => Dispatcher.Stores.dispatch(new DashboardActions.RemoteReposFetched(data)))
						);
				}
				break;
			}

		default:
			break;
		}
	},
};

Dispatcher.Backends.register(DashboardBackend.__onDispatch);

export default DashboardBackend;
