import {browserHistory} from "react-router";

import * as BuildActions from "sourcegraph/build/BuildActions";
import BuildStore from "sourcegraph/build/BuildStore";
import Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";
import {trackPromise} from "sourcegraph/app/status";
import {urlToBuild} from "sourcegraph/build/routes";

const BuildBackend = {
	fetch: defaultFetch,

	__onDispatch(action) {
		switch (action.constructor) {
		case BuildActions.WantBuilds:
			{
				let builds = BuildStore.buildLists.get(action.repo, action.search);
				if (builds === null || action.force) {
					let endpoint = !action.repo ? "/.api/builds" : `/.api/repos/${action.repo}/-/builds`;
					if (action.search) endpoint = `${endpoint}${action.search}`;
					trackPromise(
						BuildBackend.fetch(endpoint)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => Dispatcher.Stores.dispatch(new BuildActions.BuildsFetched(action.repo, data, action.search)))
					);
				}
				break;
			}

		case BuildActions.WantBuild:
			{
				let build = BuildStore.builds.get(action.repo, action.buildID);
				if (build === null || action.force) {
					trackPromise(
						BuildBackend.fetch(`/.api/repos/${action.repo}/-/builds/${action.buildID}`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => Dispatcher.Stores.dispatch(new BuildActions.BuildFetched(action.repo, action.buildID, data)))
					);
				}
				break;
			}

		case BuildActions.WantNewestBuildForCommit:
			{
				let builds = BuildStore.builds.listNewestByCommitID(action.repo, action.commitID);
				if (builds === null || action.force) {
					trackPromise(
						BuildBackend.fetch(`/.api/builds?Sort=updated_at&Direction=desc&PerPage=1&Repo=${encodeURIComponent(action.repo)}&CommitID=${encodeURIComponent(action.commitID)}`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => Dispatcher.Stores.dispatch(new BuildActions.BuildsFetchedForCommit(action.repo, action.commitID, data.Builds || [])))
					);
				}
				break;
			}

		case BuildActions.CreateBuild:
			{
				trackPromise(
					BuildBackend.fetch(`/.api/repos/${action.repo}/-/builds`, {
						method: "POST",
						body: JSON.stringify({
							CommitID: action.commitID,
							Branch: action.branch,
						}),
					})
						.then(checkStatus)
						.then((resp) => resp.json())
						.catch((err) => ({Error: err}))
						.then((data) => {
							Dispatcher.Stores.dispatch(new BuildActions.BuildFetched(action.repo, data.ID, data));
							if (!data.Error && action.redirectAfterCreation) {
								browserHistory.push(urlToBuild(action.repo, data.ID));
							}
						})
				);
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

				let url = `/.api/repos/${action.repo}/-/builds/${action.buildID}/tasks/${action.taskID}/log`;
				if (minID) {
					url += `?MinID=${minID}`;
				}

				trackPromise(
					BuildBackend.fetch(url)
						.then(checkStatus)
						.catch((err) => ({Error: err}))
						.then((resp) => {
							resp.text().then((text) => {
								let maxID = resp.headers.get("x-sourcegraph-log-max-id");
								if (maxID) {
									maxID = parseInt(maxID, 10);
								}
								Dispatcher.Stores.dispatch(new BuildActions.LogFetched(action.repo, action.buildID, action.taskID, minID, maxID, text));
							});
						})
				);
				break;
			}

		case BuildActions.WantTasks:
			{
				let tasks = BuildStore.tasks.get(action.repo, action.buildID);
				if (tasks === null || action.force) {
					trackPromise(
						BuildBackend.fetch(`/.api/repos/${action.repo}/-/builds/${action.buildID}/tasks?PerPage=1000`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => Dispatcher.Stores.dispatch(new BuildActions.TasksFetched(action.repo, action.buildID, data)))
					);
				}
				break;
			}
		}
	},
};

Dispatcher.Backends.register(BuildBackend.__onDispatch);

export default BuildBackend;
