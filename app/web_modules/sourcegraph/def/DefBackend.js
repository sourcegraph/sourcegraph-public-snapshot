// @flow weak

import * as DefActions from "sourcegraph/def/DefActions";
import DefStore from "sourcegraph/def/DefStore";
import Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";
import {updateRepoCloning} from "sourcegraph/repo/cloning";
import {singleflightFetch} from "sourcegraph/util/singleflightFetch";
import {trackPromise} from "sourcegraph/app/status";

const DefBackend = {
	fetch: singleflightFetch(defaultFetch),

	__onDispatch(action) {
		switch (action.constructor) {
		case DefActions.WantDef:
			{
				let def = DefStore.defs.get(action.repo, action.rev, action.def);
				if (def === null) {
					trackPromise(
						DefBackend.fetch(`/.api/repos/${action.repo}${action.rev ? `@${action.rev}` : ""}/-/def/${action.def}`)
							.then(updateRepoCloning(action.repo))
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => Dispatcher.Stores.dispatch(new DefActions.DefFetched(action.repo, action.rev, action.def, data)))
					);
				}
				break;
			}

		case DefActions.WantDefAuthors:
			{
				let authors = DefStore.authors.get(action.repo, action.rev, action.def);
				if (authors === null) {
					DefBackend.fetch(`/.api/repos/${action.repo}${action.rev ? `@${action.rev}` : ""}/-/def/${action.def}/-/authors`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => {
								console.error(err);
								return {Error: true};
							})
							.then((data) => Dispatcher.Stores.dispatch(new DefActions.DefAuthorsFetched(action.repo, action.rev, action.def, data)));
				}
				break;
			}

		case DefActions.WantDefs:
			{
				let defs = DefStore.defs.list(action.repo, action.rev, action.query, action.filePathPrefix);
				if (defs === null) {
					trackPromise(
						DefBackend.fetch(`/.api/defs?RepoRevs=${encodeURIComponent(action.repo)}@${encodeURIComponent(action.rev)}&Nonlocal=true&Query=${encodeURIComponent(action.query)}&FilePathPrefix=${action.filePathPrefix ? encodeURIComponent(action.filePathPrefix) : ""}`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => {
								Dispatcher.Stores.dispatch(new DefActions.DefsFetched(action.repo, action.rev, action.query, action.filePathPrefix, data, action.overlay));
							})
					);
				}
				break;
			}

		case DefActions.WantRefLocations:
			{
				let a = (action: DefActions.WantRefLocations);
				let refLocations = DefStore.getRefLocations(a.resource);
				if (refLocations === null) {
					let url = a.url();
					trackPromise(
						DefBackend.fetch(url)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => {
								if (!data) {
									data = {}; // the nil object is serialized as null by the server
								}
								if (!data.Error) {
									Dispatcher.Stores.dispatch(new DefActions.RefLocationsFetched(action, data));
								}
							})
					);
				}
				break;
			}

		case DefActions.WantRefs:
			{
				let refs = DefStore.refs.get(action.repo, action.rev, action.def, action.refRepo, action.refFile);
				if (refs === null) {
					let url = `/.api/repos/${action.repo}${action.rev ? `@${action.rev}` : ""}/-/def/${action.def}/-/refs?Repo=${encodeURIComponent(action.refRepo)}`;
					if (action.refFile) url += `&Files=${encodeURIComponent(action.refFile)}`;
					trackPromise(
						DefBackend.fetch(url)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => ({Error: err}))
							.then((data) => {
								Dispatcher.Stores.dispatch(new DefActions.RefsFetched(action.repo, action.rev, action.def, action.refRepo, action.refFile, data));
							})
					);
				}
				break;
			}
		}
	},
};

Dispatcher.Backends.register(DefBackend.__onDispatch);

export default DefBackend;
