import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as BuildActions from "sourcegraph/build/BuildActions";

function keyFor(repo, build, task) {
	let key = `${repo}#${build}`;
	if (typeof task !== "undefined") {
		key += `.${task}`;
	}
	return key;
}

export class BuildStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.builds = deepFreeze({
			content: {},
			get(repo, build) {
				return this.content[keyFor(repo, build)] || null;
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
							log: (action.minID ? existingLog.log : "") + (action.log === null ? "" : action.log),
						},
					}),
				}));
				break;
			}

		case BuildActions.TasksFetched:
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

export default new BuildStore(Dispatcher);
