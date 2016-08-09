// tslint:disable

import * as DefActions from "sourcegraph/def/DefActions";
import {DefStore} from "sourcegraph/def/DefStore";
import * as Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";
import {updateRepoCloning} from "sourcegraph/repo/cloning";
import {singleflightFetch} from "sourcegraph/util/singleflightFetch";
import {encodeDefPath} from "sourcegraph/def/index";
import get from "lodash.get";

export const DefBackend = {
	fetch: singleflightFetch(defaultFetch),

	__onDispatch(action) {
		switch (action.constructor) {
		case DefActions.WantDef:
			{
				let def = DefStore.defs.get(action.repo, action.rev, action.def);
				if (def === null) {
					DefBackend.fetch(`/.api/repos/${action.repo}${action.rev ? `@${action.rev}` : ""}/-/def/${encodeDefPath(action.def)}`)
						.then(updateRepoCloning(action.repo))
						.then(checkStatus)
						.then((resp) => resp.json())
						.catch((err) => ({Error: err}))
						.then((data) => Dispatcher.Stores.dispatch(new DefActions.DefFetched(action.repo, action.rev, action.def, data)));
				}
				break;
			}

		case DefActions.WantDefAuthors:
			{
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
				break;
			}

		case DefActions.WantDefs:
			{
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
				break;
			}

		case DefActions.WantRefLocations:
			{
				let a = action as DefActions.WantRefLocations;
				let refLocations = DefStore.getRefLocations(a.resource);
				if (refLocations === null) {
					let url = a.url();
					DefBackend.fetch(url)
						.then(checkStatus)
						.then((resp) => resp.json())
						.catch((err) => ({Error: err}))
						.then((data) => {
							if (!data || !data.Error) {
								Dispatcher.Stores.dispatch(new DefActions.RefLocationsFetched(action, data));
							}
						});
				}
				break;
			}

		case DefActions.WantExamples:
			{
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
				break;
			}

		case DefActions.WantRefs:
			{
				let refs = DefStore.refs.get(action.repo, action.commitID, action.def, action.refRepo, action.refFile);
				if (refs === null) {
					let url = `/.api/repos/${action.repo}${action.commitID ? `@${action.commitID}` : ""}/-/def/${action.def}/-/refs?Repo=${encodeURIComponent(action.refRepo)}`;
					if (action.refFile) url += `&Files=${encodeURIComponent(action.refFile)}`;
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
				break;
			}

		case DefActions.WantHoverInfo:
			{
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
				break;
			}
		}
	},
};

function convNotFound(data) {
	if (get(data, "Error.response.status") === 404) {
		return {};
	}
	return data;
}

Dispatcher.Backends.register(DefBackend.__onDispatch);
