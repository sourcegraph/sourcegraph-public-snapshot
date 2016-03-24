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
						Dispatcher.dispatch(new DefActions.DefFetched(action.url, body));
					});
				}
				break;
			}

		case DefActions.WantDefs:
			{
				let defs = DefStore.defs.list(action.repo, action.rev, action.query);
				if (defs === null) {
					DefBackend.xhr({
						uri: `/.api/.defs?RepoRevs=${encodeURIComponent(action.repo)}@${encodeURIComponent(action.rev)}&Nonlocal=true&Query=${encodeURIComponent(action.query)}`,
						json: {},
					}, function(err, resp, body) {
						if (err) {
							console.error(err);
							return;
						}
						Dispatcher.dispatch(new DefActions.DefsFetched(action.repo, action.rev, action.query, body));
					});
				}
				break;
			}

		case DefActions.WantExamples:
			{
				let examples = DefStore.examples.get(action.defURL, action.file, action.page);
				if (examples === null) {
					let files = action.file ? `&files=${action.file}` : "";
					DefBackend.xhr({
						uri: `/.api/repos${action.defURL}/.examples?PerPage=10&Page=${action.page}${files}`,
						json: {},
					}, function(err, resp, body) {
						if (!err && (resp.statusCode !== 200)) err = `HTTP ${resp.statusCode}`;
						if (err) {
							console.error(err);
							return;
						}
						if (!body || !body.Examples || body.Examples.length === 0) {
							Dispatcher.dispatch(new DefActions.NoExamplesAvailable(action.defURL, action.page));
							return;
						}
						Dispatcher.dispatch(new DefActions.ExamplesFetched(action.defURL, action.file, action.page, body));
					});
				}
				break;
			}

		case DefActions.WantRefs:
			{
				{
					DefBackend.xhr({
						uri: `/.ui${action.defURL}/.refs`,
						json: {},
					}, function(err, resp, body) {
						if (err) {
							console.error(err);
							return;
						}
						Dispatcher.dispatch(new DefActions.RefsFetched(action.defURL, body));
					});
				}
				break;
			}
		}
	},
};

Dispatcher.register(DefBackend.__onDispatch);

export default DefBackend;
