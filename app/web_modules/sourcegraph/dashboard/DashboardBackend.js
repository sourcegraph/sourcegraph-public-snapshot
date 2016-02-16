import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import Dispatcher from "sourcegraph/Dispatcher";
import defaultXhr from "sourcegraph/util/xhr";

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

	__onDispatch(action) {
		switch (action.constructor) {
		case DashboardActions.WantAddRepos:
			console.log("got here", action.repos);
			DashboardBackend.xhr({
				uri: `.ui/.repo-mirror`,
				method: "POST",
				json: {
					Repos: action.repos,
				},
			}, function(err, resp, body) {
				if (err) {
					console.error(err);
					return;
				}
				// TODO dispath repo added action here.
			});
			break;
		case DashboardActions.WantAddUsers:
			console.log("got here", action.emails);
			DashboardBackend.xhr({
				uri: `.ui/.invite-bulk`,
				method: "POST",
				json: {
					Emails: action.emails,
				},
			}, function(err, resp, body) {
				if (err) {
					console.error(err);
					return;
				}
				// TODO dispath user invited action here.
			});
			break;
		}
	},
};

Dispatcher.register(DashboardBackend.__onDispatch);

export default DashboardBackend;
