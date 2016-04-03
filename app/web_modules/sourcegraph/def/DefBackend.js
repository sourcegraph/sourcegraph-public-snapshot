// @flow weak

import * as DefActions from "sourcegraph/def/DefActions";
import DefStore from "sourcegraph/def/DefStore";
import Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";

const DefBackend = {
	fetch: defaultFetch,

	__onDispatch(action) {
		switch (action.constructor) {
		case DefActions.WantDef:
			{
				let def = DefStore.defs.get(action.repo, action.rev, action.def);
				if (def === null) {
					DefBackend.fetch(`/.api/repos/${action.repo}${action.rev ? `@${action.rev}` : ""}/-/def/${action.def}`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => {
								console.error(err);
								return {Error: true};
							})
							.then((data) => Dispatcher.Stores.dispatch(new DefActions.DefFetched(action.repo, action.rev, action.def, data)));
				}
				break;
			}

		case DefActions.WantDefs:
			{
				let defs = DefStore.defs.list(action.repo, action.rev, action.query);
				if (defs === null) {
					DefBackend.fetch(`/.api/defs?RepoRevs=${encodeURIComponent(action.repo)}@${encodeURIComponent(action.rev)}&Nonlocal=true&Query=${encodeURIComponent(action.query)}`)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => {
								console.error(err);
								return {Error: true};
							})
							.then((data) => {
								Dispatcher.Stores.dispatch(new DefActions.DefsFetched(action.repo, action.rev, action.query, data));
							});
				}
				break;
			}

		case DefActions.WantRefs:
			{
				let refs = DefStore.refs.get(action.repo, action.rev, action.def, action.file);
				if (refs === null) {
					let url = `/.api/repos/${action.repo}${action.rev ? `@${action.rev}` : ""}/-/def/${action.def}/-/refs`;
					if (action.file) url += `?Files=${encodeURIComponent(action.file)}`;
					DefBackend.fetch(url)
							.then(checkStatus)
							.then((resp) => resp.json())
							.catch((err) => {
								console.error(err);
								return null;
							})
							.then((data) => {
								Dispatcher.Stores.dispatch(new DefActions.RefsFetched(action.repo, action.rev, action.def, action.file, data));
							});
				}
				break;
			}
		}
	},
};

Dispatcher.Backends.register(DefBackend.__onDispatch);

export default DefBackend;
