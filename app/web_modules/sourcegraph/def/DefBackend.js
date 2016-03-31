import * as DefActions from "sourcegraph/def/DefActions";
import DefStore from "sourcegraph/def/DefStore";
import Dispatcher from "sourcegraph/Dispatcher";
import defaultXhr from "sourcegraph/util/xhr";

const DefBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {
		case DefActions.WantDef:
			{
				let def = DefStore.defs.get(action.url);
				if (def === null) {
					DefBackend.xhr({
						uri: `/.api/repos${action.url}`,
						json: {},
					}, function(err, resp, body) {
						if (resp.statusCode !== 200) body = {Error: true};
						if (err) {
							console.error(err);
							return;
						}
						Dispatcher.Stores.dispatch(new DefActions.DefFetched(action.url, body));
					});
				}
				break;
			}

		case DefActions.WantDefs:
			{
				let defs = DefStore.defs.list(action.repo, action.rev, action.query);
				if (defs === null) {
					DefBackend.xhr({
						uri: `/.api/defs?RepoRevs=${encodeURIComponent(action.repo)}@${encodeURIComponent(action.rev)}&Nonlocal=true&Query=${encodeURIComponent(action.query)}`,
						json: {},
					}, function(err, resp, body) {
						if (err) {
							console.error(err);
							return;
						}
						Dispatcher.Stores.dispatch(new DefActions.DefsFetched(action.repo, action.rev, action.query, body));
					});
				}
				break;
			}

		case DefActions.WantRefs:
			{
				let refs = DefStore.refs.get(action.defURL, action.file);
				if (refs === null) {
					let url = `/.ui${action.defURL}/-/refs`;
					if (action.file) url += `?Files=${encodeURIComponent(action.file)}`;
					DefBackend.xhr({
						uri: url,
						json: {},
					}, function(err, resp, body) {
						if (err) {
							console.error(err);
							return;
						}
						Dispatcher.Stores.dispatch(new DefActions.RefsFetched(action.defURL, action.file, body));
					});
				}
				break;
			}
		}
	},
};

Dispatcher.Backends.register(DefBackend.__onDispatch);

export default DefBackend;
