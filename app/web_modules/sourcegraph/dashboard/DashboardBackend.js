import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";

const DashboardBackend = {
	fetch: defaultFetch,

	__onDispatch(action) {
		switch (action.constructor) {
		case DashboardActions.WantRepos:
			{
				let repos = DashboardStore.repos;
				if (repos === null) {
					DashboardBackend.fetch("/.api/repos")
						.then(checkStatus)
						.then((resp) => resp.json())
						.catch((err) => {
							console.error(err);
							return {Error: true};
						})
						.then((data) => Dispatcher.Stores.dispatch(new DashboardActions.ReposFetched(data)));
				}
				break;
			}

		case DashboardActions.WantRemoteRepos:
			{
				let remoteRepos = DashboardStore.remoteRepos;
				if (remoteRepos === null) {
					DashboardBackend.fetch("/.api/remote-repos")
						.then(checkStatus)
						.then((resp) => resp.json())
						.catch((err) => {
							console.error(err);
							return {Error: true};
						})
						.then((data) => Dispatcher.Stores.dispatch(new DashboardActions.RemoteReposFetched(data)));
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
