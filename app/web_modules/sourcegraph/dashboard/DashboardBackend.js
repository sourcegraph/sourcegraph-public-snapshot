import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import Dispatcher from "sourcegraph/Dispatcher";
import defaultXhr from "xhr";

// function authHeaders() {
// 	let hdr = {};
// 	if (typeof document !== "undefined" && document.head.dataset && document.head.dataset.currentUserOauth2AccessToken) {
// 		let auth = `x-oauth-basic:${document.head.dataset.currentUserOauth2AccessToken}`;
// 		hdr.authorization = `Basic ${btoa(auth)}`;
// 	}
// 	return hdr;
// }

const DashboardBackend = {
	xhr: defaultXhr,
	dashboardStore: DashboardStore,

	__onDispatch(action) {
		switch (action.constructor) {
		case DashboardActions.WantAddRepos:
			{
				console.log("got here", action.repos);
				// Dispatcher.dispatch(new DashboardActions.ReposAdded());
				break;
			}

		case DashboardActions.WantAddTeammates:
			{
				console.log("got there");
				// Dispatcher.dispatch(new DashboardActions.UsersAdded());
				break;
			}
		}
	},
};

Dispatcher.register(DashboardBackend.__onDispatch);

export default DashboardBackend;
