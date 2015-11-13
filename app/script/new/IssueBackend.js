import * as IssueActions from "./IssueActions";
import Dispatcher from "./Dispatcher";
import defaultXhr from "xhr";

const IssueBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {
		case IssueActions.CreateIssue:
			IssueBackend.xhr({
				uri: `/${action.repo}/.issues/new`,
				method: "POST",
				headers: {
					"X-Csrf-Token": window._csrfToken,
				},
				json: {
					title: action.title,
					body: action.body,
					reference: {
						Repo: {
							URI: action.repo,
						},
						CommitID: action.commitID,
						Path: action.path,
						StartLine: action.startLine,
						EndLine: action.endLine,
					},
				},
			}, function(err, resp, body) {
				if (err) {
					console.error(err);
					return;
				}
				action.callback();
			});
			break;
		}
	},
};

Dispatcher.register(IssueBackend.__onDispatch);

export default IssueBackend;
