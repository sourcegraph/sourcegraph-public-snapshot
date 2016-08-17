import get from "lodash.get";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as DefActions from "sourcegraph/def/DefActions";
import {DefStore} from "sourcegraph/def/DefStore";
import {encodeDefPath} from "sourcegraph/def/index";
import {updateRepoCloning} from "sourcegraph/repo/cloning";
import {singleflightFetch} from "sourcegraph/util/singleflightFetch";
import {toQuery} from "sourcegraph/util/toQuery";
import {checkStatus, defaultFetch} from "sourcegraph/util/xhr";

export const DefBackend = {
	fetch: singleflightFetch(defaultFetch),

	__onDispatch(payload: DefActions.Action): void {
		if (payload instanceof DefActions.WantDef) {
			const action = payload;
			let def = DefStore.defs.get(action.repo, action.rev, action.def);
			if (def === null) {
				DefBackend.fetch(`/.api/repos/${action.repo}${action.rev ? `@${action.rev}` : ""}/-/def/${encodeDefPath(action.def)}`)
					.then(updateRepoCloning(action.repo))
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then((data) => Dispatcher.Stores.dispatch(new DefActions.DefFetched(action.repo, action.rev, action.def, data)));
			}
		}

		if (payload instanceof DefActions.WantDefAuthors) {
			const action = payload;
			let authors = DefStore.authors.get(action.repo, action.commitID, action.def);
			if (authors === null) {
				DefBackend.fetch(`/.api/repos/${action.repo}${action.commitID ? `@${action.commitID}` : ""}/-/def/${action.def}/-/authors`)
						.then(checkStatus)
						.then((resp) => resp.json())
						.catch((err) => {
							console.error(err);
							return {Error: true};
						})
						.then((data) => Dispatcher.Stores.dispatch(new DefActions.DefAuthorsFetched(action.repo, action.commitID, action.def, data)));
			}
		}

		if (payload instanceof DefActions.WantDefs) {
			const action = payload;
			let defs = DefStore.defs.list(action.repo, action.commitID, action.query, action.filePathPrefix);
			if (defs === null) {
				DefBackend.fetch(`/.api/defs?RepoRevs=${encodeURIComponent(action.repo)}@${encodeURIComponent(action.commitID)}&Nonlocal=true&Query=${encodeURIComponent(action.query)}&FilePathPrefix=${action.filePathPrefix ? encodeURIComponent(action.filePathPrefix) : ""}`)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then((data) => {
						Dispatcher.Stores.dispatch(new DefActions.DefsFetched(action.repo, action.commitID, action.query, action.filePathPrefix, data, action.overlay));
					});
			}
		}

		if (payload instanceof DefActions.WantRefLocations) {
			const action = payload;
			let refLocations = DefStore.getRefLocations(action.resource);
			if (refLocations === null) {
				let q = toQuery({
					Repos: action.resource.repos,
				});
				if (q) {
					q = `?${q}`;
				}
				DefBackend.fetch(`/.api/repos/${action.resource.repo}${action.resource.commitID ? `@${action.resource.commitID}` : ""}/-/def/${action.resource.def}/-/ref-locations${q}`)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then((data) => {
						if (!data || !data.Error) {
							Dispatcher.Stores.dispatch(new DefActions.RefLocationsFetched(action, data));
						}
					});
			}
		}

		if (payload instanceof DefActions.WantLocalRefLocations) {
			const action = payload;
			let refLocations = DefStore.getRefLocations(action.resource);
			if (refLocations === null) {
				let q = toQuery({
					Repos: action.resource.repos,
				});
				if (q) {
					q = `&${q}`;
				}
				DefBackend.fetch(`/.api/repos/${action.resource.repo}${action.resource.commitID ? `@${action.resource.commitID}` : ""}/-/def/${action.resource.def}/-/local-refs?file=${action.pos.file}&line=${action.pos.line}&character=${action.pos.character}${q}`)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then((data) => {
						if (!data || !data.Error) {
							Dispatcher.Stores.dispatch(new DefActions.LocalRefLocationsFetched(action, data));
						}
					});
			}
		}

		if (payload instanceof DefActions.WantExamples) {
			const action = payload;
			let a = action as DefActions.WantExamples;
			let examples = DefStore.getExamples(a.resource);
			if (examples === null) {
				let url = a.url();
				DefBackend.fetch(url)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then(convNotFound)
					.then((data) => {
						if (!data || !data.Error) {
							Dispatcher.Stores.dispatch(new DefActions.ExamplesFetched(action, data));
						}
					});
			}
		}

		if (payload instanceof DefActions.WantRefs) {
			const action = payload;
			let refs = DefStore.refs.get(action.repo, action.commitID, action.def, action.refRepo, action.refFile);
			if (refs === null) {
				let url = `/.api/repos/${action.repo}${action.commitID ? `@${action.commitID}` : ""}/-/def/${action.def}/-/refs?Repo=${encodeURIComponent(action.refRepo)}`;
				if (action.refFile) {
					url += `&Files=${encodeURIComponent(action.refFile)}`;
				}
				DefBackend.fetch(url)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then((data) => {
						if (get(data, "Error.response.status") === 404) {
							data = [];
						}
						Dispatcher.Stores.dispatch(new DefActions.RefsFetched(action.repo, action.commitID, action.def, action.refRepo, action.refFile, data));
					});
			}
		}

		if (payload instanceof DefActions.WantHoverInfo) {
			const action = payload;
			let info = DefStore.hoverInfos.get(action.pos);
			if (info === null) {
				let rev = action.pos.commit ? `@${action.pos.commit}` : "";
				let url = `/.api/repos/${action.pos.repo}${rev}/-/hover-info?file=${action.pos.file}&line=${action.pos.line}&character=${action.pos.character}`;
				DefBackend.fetch(url)
					.then(checkStatus)
					.then((resp) => resp.json())
					.catch((err) => ({Error: err}))
					.then((data) => {
						if (get(data, "Error.response.status") === 404) {
							return; // TODO we may want to indicate error in the UI
						}
						Dispatcher.Stores.dispatch(new DefActions.HoverInfoFetched(action.pos, data));
					});
			}
		}

		if (payload instanceof DefActions.WantJumpDef) {
			const action = payload;
			let rev = action.pos.commit ? `@${action.pos.commit}` : "";
			let url = `/.api/repos/${action.pos.repo}${rev}/-/jump-def?file=${action.pos.file}&line=${action.pos.line}&character=${action.pos.character}`;
			DefBackend.fetch(url)
				.then(checkStatus)
				.then((resp) => resp.json())
				.catch((err) => ({Error: err}))
				.then((data) => {
					Dispatcher.Stores.dispatch(new DefActions.JumpDefFetched(action.pos, data));
				});
		}
	},
};

function convNotFound(data: any): any {
	if (get(data, "Error.response.status") === 404) {
		return {};
	}
	return data;
}

Dispatcher.Backends.register(DefBackend.__onDispatch);
