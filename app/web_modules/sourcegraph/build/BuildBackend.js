import * as BuildActions from "sourcegraph/build/BuildActions";
import BuildStore from "sourcegraph/build/BuildStore";
import Dispatcher from "sourcegraph/Dispatcher";
import defaultXhr from "sourcegraph/util/xhr";

const BuildBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {
		case BuildActions.WantBuild:
			{
				let build = BuildStore.builds.get(action.repo, action.buildID);
				if (build === null || action.force) {
					BuildBackend.xhr({
						uri: `/.api/repos/${action.repo}/.builds/${action.buildID}`,
						json: {},
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

		case BuildActions.WantNewestBuildForCommit:
			{
				let builds = BuildStore.builds.listNewestByCommitID(action.repo, action.commitID);
				if (builds === null || action.force) {
					BuildBackend.xhr({
						uri: `/.api/builds?Sort=updated_at&Direction=desc&PerPage=1&Repo=${encodeURIComponent(action.repo)}&CommitID=${encodeURIComponent(action.commitID)}`,
						json: {},
					}, function(err, resp, body) {
						if (!err && (resp.statusCode !== 200 && resp.statusCode !== 201)) err = `HTTP ${resp.statusCode}`;
						if (err) {
							console.error(err);
							return;
						}
						Dispatcher.dispatch(new BuildActions.BuildsFetchedForCommit(action.repo, action.commitID, body.Builds || []));
					});
				}
				break;
			}

		case BuildActions.CreateBuild:
			{
				BuildBackend.xhr({
					uri: `/.api/repos/${action.repo}/.builds`,
					method: "post",
					json: {
						CommitID: action.commitID,
						Branch: action.branch,
						Config: {Queue: true},
					},
				}, function(err, resp, body) {
					if (!err && (resp.statusCode !== 200 && resp.statusCode !== 201)) err = `HTTP ${resp.statusCode}`;
					if (err) {
						console.error(err);
						return;
					}
					Dispatcher.dispatch(new BuildActions.BuildFetched(action.repo, body.ID, body));
				});
				break;
			}

		case BuildActions.WantLog:
			{
				// Only fetch log lines newer than those we've already fetched.
				let minID = 0;
				let log = BuildStore.logs.get(action.repo, action.buildID, action.taskID);
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
					if (resp.statusCode !== 200) {
						console.error(`HTTP status ${resp.statusCode} received while fetching logs from ${url}`);
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
				let tasks = BuildStore.tasks.get(action.repo, action.buildID);
				if (tasks === null || action.force) {
					BuildBackend.xhr({
						uri: `/.api/repos/${action.repo}/.builds/${action.buildID}/.tasks?PerPage=1000`,
						json: {},
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
