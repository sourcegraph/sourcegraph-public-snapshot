// tslint:disable: typedef ordered-imports

import {Store} from "sourcegraph/Store";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {deepFreeze} from "sourcegraph/util/deepFreeze";
import * as BuildActions from "sourcegraph/build/BuildActions";
import {updatedAt} from "sourcegraph/build/Build";
import "sourcegraph/build/BuildBackend";

function keyFor(repo, build, task?) {
	let key = `${repo}#${build}`;
	if (typeof task !== "undefined") {
		key += `.${task}`;
	}
	return key;
}

function keyForList(repo, search) {
	return `${repo}${search}`;
}

class BuildStoreClass extends Store<any> {
	builds: any;
	buildLists: any;
	logs: any;
	tasks: any;

	constructor(dispatcher) {
		super(dispatcher);
	}

	reset() {
		this.builds = deepFreeze({
			content: {},
			_fetchedForCommit: {},
			get(repo, build) {
				return this.content[keyFor(repo, build)] || null;
			},

			// listNewestByCommitID returns the latest builds for the given repo
			// and commit ID. If you used WantNewestBuildForCommit, this may
			// only be 1 build. After the fetch completes (regardless of success),
			// listNewestByCommitID always returns a non-null value (an empty array
			// if there are no builds).
			listNewestByCommitID(repo, commitID) {
				if (!this._fetchedForCommit[keyFor(repo, commitID)]) {
					return null;
				}
				const builds = (Object as any).values(this.content).filter((b) =>
					b.Repo === repo && b.CommitID === commitID
				);
				if (builds === null) {
					return null;
				}
				return builds.sort((a, b) => {
					// These date strings ("2016-03-07T22:51:43.202747Z") lexically sort.
					const ta = updatedAt(a);
					const tb = updatedAt(b);
					if (ta === tb) {
						return 0;
					}
					if (ta > tb) {
						return -1; // Newest first.
					}
					return 1;
				});
			},
		});
		this.buildLists = deepFreeze({
			content: {},
			get(repo, search) {
				return this.content[keyForList(repo, search)] || null;
			},
		});
		this.logs = deepFreeze({
			content: {},
			get(repo, build, task) {
				return this.content[keyFor(repo, build, task)] || null;
			},
		});
		this.tasks = deepFreeze({
			content: {},
			get(repo, build) {
				return this.content[keyFor(repo, build)] || null;
			},
		});
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case BuildActions.BuildFetched:
			this.builds = deepFreeze(Object.assign({}, this.builds, {
				content: Object.assign({}, this.builds.content, {
					[keyFor(action.repo, action.buildID)]: action.build,
				}),
			}));
			break;

		case BuildActions.BuildsFetched:
			this.buildLists = deepFreeze(Object.assign({}, this.buildLists, {
				content: Object.assign({}, this.buildLists.content, {
					[keyForList(action.repo, action.search)]: action.builds.Builds || [],
				}),
			}));
			break;

		case BuildActions.BuildsFetchedForCommit:
			this.builds = deepFreeze(Object.assign({}, this.builds, {
				_fetchedForCommit: Object.assign({}, this.builds._fetchedForCommit, {
					[keyFor(action.repo, action.commitID)]: true,
				}),
			}));
			action.builds.forEach((b) => {
				this.__onDispatch(new BuildActions.BuildFetched(action.repo, b.ID, b));
			});
			break;

		case BuildActions.LogFetched:
			{
				// Append to existing log if we're fetching the portion
				// right after the existing log data.
				let existingLog = this.logs.get(action.repo, action.buildID, action.taskID);
				if (!existingLog) {
					existingLog = {log: "", maxID: 0};
				}
				// TODO(sqs): Handle nonsequential log fetches
				// (current log ends at ${existingLog.maxID}, fetch
				// range begins at ${action.minID}. Trigger a fetch of
				// the full range next time.
				this.logs = deepFreeze(Object.assign({}, this.logs, {
					content: Object.assign({}, this.logs.content, {
						[keyFor(action.repo, action.buildID, action.taskID)]: {
							maxID: action.maxID,
							log: existingLog.log + (action.log === null ? "" : action.log),
						},
					}),
				}));
				break;
			}

		case BuildActions.TasksFetched:
			if (JSON.stringify(action.tasks) === JSON.stringify(this.tasks.get(action.repo, action.buildID))) {
				return;
			}
			this.tasks = deepFreeze(Object.assign({}, this.tasks, {
				content: Object.assign({}, this.tasks.content, {
					[keyFor(action.repo, action.buildID)]: action.tasks,
				}),
			}));
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export const BuildStore = new BuildStoreClass(Dispatcher.Stores);
