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
				// console.log("want add repos!!!!!");
				// console.log(action.repos);
				// console.log(window.location.href);
				//
				// let formData = new FormData();
				// action.repos.forEach(repo => formData.append("RepoURI[]", repo.URI));
				// formData.append("csrf_token", window.csrfToken);

				// $.ajax({
				// 	url: window.addRepoURL,
				// 	data: formData,
				// 	type: 'POST',
				// 	success: (data) => {
				// 		console.log("successful!");
				// 	}
				// });
				// DashboardBackend.xhr.post(window.addRepoURL)
				// console.log("about to send form data!");
				// formData.submit(window.addRepoURL, (err, res) => {
				// 	console.log("inside the callback!", err, res);
				// });
				// DashboardBackend.xhr({
				// 	// uri: `/.api/repos/${action.repo}/.builds/${action.buildID}`,
				// 	uri: window.addRepoURL, // TODO: put this somewhere more sensible
				// 	json: {},
				// 	headers: authHeaders(),
				// }, function(err, resp, body) {
				// 	if (err) {
				// 		console.error(err);
				// 		return;
				// 	}
				// 	Dispatcher.dispatch(new DashboardActions.ReposAdded(action.repos));
				// });
				break;
			}

		case DashboardActions.WantAddTeammates:
			{
				setTimeout(() => Dispatcher.dispatch(new DashboardActions.TeammatesAdded(action.teammates)));
				// Only fetch log lines newer than those we've already fetched.
				// let minID = 0;
				// let log = DashboardBackend.dashboardStore.logs.get(action.repo, action.buildID, action.taskID);
				// if (log !== null) {
				// 	minID = log.maxID;
				// }
				//
				// let url = `/${action.repo}/.builds/${action.buildID}/tasks/${action.taskID}/log`;
				// if (minID) {
				// 	url += `?MinID=${minID}`;
				// }
				//
				// DashboardBackend.xhr({
				// 	uri: url,
				// }, function(err, resp, body) {
				// 	if (err) {
				// 		console.error(err);
				// 		return;
				// 	}
				// 	if (resp.statusCode !== 200) {
				// 		console.error(`HTTP status ${resp.statusCode} received while fetching logs from ${url}`);
				// 		return;
				// 	}
				// 	let maxID = resp.headers["x-sourcegraph-log-max-id"];
				// 	if (maxID) {
				// 		maxID = parseInt(maxID, 10);
				// 	}
				// 	Dispatcher.dispatch(new DashboardActions.LogFetched(action.repo, action.buildID, action.taskID, minID, maxID, body));
				// });
				break;
			}
		}
	},
};

Dispatcher.register(DashboardBackend.__onDispatch);

export default DashboardBackend;
