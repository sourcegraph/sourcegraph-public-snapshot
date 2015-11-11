import * as DefActions from "./DefActions";
import DefStore from "./DefStore";
import Dispatcher from "./Dispatcher";
import defaultXhr from "xhr";

// TODO preloading
const DefBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {
		case DefActions.WantDef:
			let def = DefStore.defs.get(action.url);
			if (def === null) {
				DefBackend.xhr({
					uri: `/.ui${action.url}`,
					headers: {
						"X-Definition-Data-Only": "yes",
					},
					json: {},
				}, function(err, resp, body) {
					if (err) {
						console.error(err);
						return;
					}
					Dispatcher.dispatch(new DefActions.DefFetched(action.url, body));
				});
			}
			break;

		case DefActions.WantExample:
			let example = DefStore.examples.get(action.defURL, action.index);
			if (example === null) {
				DefBackend.xhr({
					uri: `/.ui${action.defURL}/.examples?TokenizedSource=true&PerPage=1&Page=${action.index + 1}`,
					json: {},
				}, function(err, resp, body) {
					if (err) {
						console.error(err);
						return;
					}
					if (body === null || body.Error) {
						Dispatcher.dispatch(new DefActions.NoExampleAvailable(action.defURL, action.index));
						return;
					}
					Dispatcher.dispatch(new DefActions.ExampleFetched(action.defURL, action.index, body[0]));
				});
			}
			break;

		case DefActions.WantDiscussions:
			let discussions = DefStore.discussions.get(action.defURL);
			if (discussions === null) {
				DefBackend.xhr({
					uri: `/.ui${action.defURL}/.discussions?order=Date`,
					json: {},
				}, function(err, resp, body) {
					if (err) {
						console.error(err);
						return;
					}
					Dispatcher.dispatch(new DefActions.DiscussionsFetched(action.defURL, body.Discussions ? body.Discussions.map(normalizeDiscussion) : []));
				});
			}
			break;

		case DefActions.CreateDiscussion:
			DefBackend.xhr({
				uri: `/.ui${action.defURL}/.discussions/create`,
				method: "POST",
				json: {
					Title: action.title,
					Description: action.description,
				},
			}, function(err, resp, body) {
				if (err) {
					console.error(err);
					return;
				}
				Dispatcher.dispatch(new DefActions.DiscussionsFetched(action.defURL, [normalizeDiscussion(body)].concat(DefStore.discussions.get(action.defURL))));
				action.callback(body);
			});
			break;

		case DefActions.CreateDiscussionComment:
			DefBackend.xhr({
				uri: `/.ui${action.defURL}/.discussions/${action.discussionID}/.comment`,
				method: "POST",
				json: {
					Body: action.body,
				},
			}, function(err, resp, body) {
				if (err) {
					console.error(err);
					return;
				}
				let list = DefStore.discussions.get(action.defURL).map((d) =>
					d.ID === action.discussionID ? Object.assign(d, {Comments: d.Comments.concat([body])}) : d
				);
				Dispatcher.dispatch(new DefActions.DiscussionsFetched(action.defURL, list));
				action.callback();
			});
			break;

		}
	},
};

function normalizeDiscussion(d) {
	d.Comments = d.Comments || []; // TODO fix this in backend
	return d;
}

Dispatcher.register(DefBackend.__onDispatch);

export default DefBackend;
