import Store from "sourcegraph/Store";
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DefActions from "sourcegraph/def/DefActions";

function defsListKeyFor(repo, rev, query) {
	return `${repo}#${rev}#${query}`;
}

function refsKeyFor(defURL, repo, file) {
	return `${defURL}:${repo}:${file || ""}`;
}

export class DefStore extends Store {
	reset(data) {
		this.defs = deepFreeze({
			content: data && data.defs ? data.defs : {},
			get(url) {
				return this.content[url] || null;
			},
			list(repo, rev, query) {
				return this.content[defsListKeyFor(repo, rev, query)] || null;
			},
		});
		this.activeDef = null;
		this.highlightedDef = null;
		this.refs = deepFreeze({
			content: {},
			get(defURL, repo, file) {
				return this.content[refsKeyFor(defURL, repo, file)] || null;
			},
		});
		this.refLocations = deepFreeze({
			content: {},
			get(defURL) {
				return this.content[defURL] || null;
			},
		});

		this.defOptionsURLs = null;
		this.defOptionsLeft = 0;
		this.defOptionsTop = 0;
	}

	__onDispatch(action) {
		switch (action.constructor) {
		case DefActions.DefFetched:
			this.defs = deepFreeze(Object.assign({}, this.defs, {
				content: Object.assign({}, this.defs.content, {
					[action.url]: action.def,
				}),
			}));
			break;

		case DefActions.DefsFetched:
			this.defs = deepFreeze(Object.assign({}, this.defs, {
				content: Object.assign({}, this.defs.content, {
					[defsListKeyFor(action.repo, action.rev, action.query)]: action.defs,
				}),
			}));
			break;

		case DefActions.HighlightDef:
			this.highlightedDef = action.url;
			break;

		case DefActions.RefLocationsFetched:
			this.refLocations = deepFreeze(Object.assign({}, this.refLocations, {
				content: Object.assign({}, this.refLocations.content, {
					[action.defURL]: getRankedRefLocations(action.locations),
				}),
			}));
			break;

		case DefActions.RefsFetched:
			this.refs = deepFreeze(Object.assign({}, this.refs, {
				content: Object.assign({}, this.refs.content, {
					[refsKeyFor(action.defURL, action.repo, action.file)]: action.refs,
				}),
			}));
			break;

		case DefActions.SelectMultipleDefs:
			this.defOptionsURLs = action.urls;
			this.defOptionsLeft = action.left;
			this.defOptionsTop = action.top;
			break;

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

function getRankedRefLocations(locations) {
	if (locations.length <= 2) {
		return locations;
	}
	let dashboardRepos = DashboardStore.repos;
	let repos = [];

	// The first repo of locations is the current repo.
	repos.push(locations[0]);

	let otherRepos = [];
	let i = 1;
	for (; i < locations.length; i++) {
		if (locations[i].Repo in dashboardRepos) {
			repos.push(locations[i]);
		} else {
			otherRepos.push(locations[i]);
		}
	}
	Array.prototype.push.apply(repos, otherRepos);
	return repos;
}

export default new DefStore(Dispatcher.Stores);
