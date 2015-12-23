import * as BuildActions from "sourcegraph/build/BuildActions";
import BuildStore from "sourcegraph/build/BuildStore";
import Dispatcher from "sourcegraph/Dispatcher";
import defaultXhr from "xhr";

function authHeaders() {
	let hdr = {};
	if (typeof document !== "undefined" && document.head.dataset && document.head.dataset.currentUserOauth2AccessToken) {
		let auth = `x-oauth-basic:${document.head.dataset.currentUserOauth2AccessToken}`;
		hdr.authorization = `Basic ${btoa(auth)}`;
	}
	return hdr;
}

const BuildBackend = {
	xhr: defaultXhr,
	buildStore: BuildStore,

	__onDispatch(action) {
		switch (action.constructor) {
		case BuildActions.WantBuild:
			{
				let build = BuildBackend.buildStore.builds.get(action.repo, action.buildID);
				if (build === null || action.force) {
					BuildBackend.xhr({
						uri: `/.api/repos/${action.repo}/.builds/${action.buildID}`,
						json: {},
						headers: authHeaders(),
					}, function(err, resp, body) {
						if (err) {
							console.error(err);
							return;
						}
						Dispatcher.dispatch(new BuildActions.BuildFetched(action.repo, action.buildID, body));
					});
				}
				break;
			}

		case BuildActions.WantLog:
			{
				// Only fetch log lines newer than those we've already fetched.
				let minID = 0;
				let log = BuildBackend.buildStore.logs.get(action.repo, action.buildID, action.taskID);
				if (log !== null) {
					minID = log.maxID;
				}

				let url = `/${action.repo}/.builds/${action.buildID}/tasks/${action.taskID}/log`;
				if (minID) {
					url += `?MinID=${minID}`;
				}

				BuildBackend.xhr({
					uri: url,
				}, function(err, resp, body) {
					if (err) {
						console.error(err);
						return;
					}
					let maxID = resp.headers["x-sourcegraph-log-max-id"];
					if (maxID) {
						maxID = parseInt(maxID, 10);
					}
					Dispatcher.dispatch(new BuildActions.LogFetched(action.repo, action.buildID, action.taskID, minID, maxID, body));
				});
				break;
			}

		case BuildActions.WantTasks:
			{
				let tasks = BuildBackend.buildStore.tasks.get(action.repo, action.buildID);
				if (tasks === null || action.force) {
					BuildBackend.xhr({
						uri: `/.api/repos/${action.repo}/.builds/${action.buildID}/.tasks`,
						json: {},
						headers: authHeaders(),
					}, function(err, resp, body) {
						if (err) {
							console.error(err);
							return;
						}
						Dispatcher.dispatch(new BuildActions.TasksFetched(action.repo, action.buildID, body));
					});
				}
				break;
			}
		}
	},
};

Dispatcher.register(BuildBackend.__onDispatch);

export default BuildBackend;
