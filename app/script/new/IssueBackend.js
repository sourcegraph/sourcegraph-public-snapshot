import * as IssueActions from "./IssueActions";
import Dispatcher from "./Dispatcher";
import defaultXhr from "./util/xhr";

const IssueBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {
		case IssueActions.CreateIssue:
			IssueBackend.xhr({
				uri: `/${action.repo}/.threads/new`,
				method: "POST",
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
				action.callback(body);
			});
			break;
		}
	},
};

Dispatcher.register(IssueBackend.__onDispatch);

export default IssueBackend;
