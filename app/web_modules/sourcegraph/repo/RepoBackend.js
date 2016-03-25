import * as RepoActions from "sourcegraph/repo/RepoActions";
import RepoStore from "sourcegraph/repo/RepoStore";
import Dispatcher from "sourcegraph/Dispatcher";
import defaultXhr from "sourcegraph/util/xhr";

const RepoBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {

		case RepoActions.WantRepo:
			{
				let repo = RepoStore.repos.get(action.repo);
				if (repo === null) {
					RepoBackend.xhr({
						uri: `/.api/repos/${action.repo}`,
						json: {},
					}, function(err, resp, body) {
						if (!err && resp.statusCode !== 200) err = `HTTP ${resp.statusCode}`;
						if (err) {
							console.error(err);
							return;
						}
						Dispatcher.dispatch(new RepoActions.FetchedRepo(action.repo, body));
					});
				}
				break;
			}

		case RepoActions.WantBranches:
			{
				let branches = RepoStore.branches.list(action.repo);
				if (branches === null) {
					RepoBackend.xhr({
						uri: `/.api/repos/${action.repo}/.branches`,
						json: {},
					}, function(err, resp, body) {
						if (!err && resp.statusCode !== 200) err = `HTTP ${resp.statusCode}`;
						if (err) {
							console.error(err);
							return;
						}
						Dispatcher.dispatch(new RepoActions.FetchedBranches(action.repo, body.Branches || [], err));
					});
				}
				break;
			}

		case RepoActions.WantTags:
			{
				let tags = RepoStore.tags.list(action.repo);
				if (tags === null) {
					RepoBackend.xhr({
						uri: `/.api/repos/${action.repo}/.tags`,
						json: {},
					}, function(err, resp, body) {
						if (!err && resp.statusCode !== 200) err = `HTTP ${resp.statusCode}`;
						if (err) {
							console.error(err);
							return;
						}
						Dispatcher.dispatch(new RepoActions.FetchedTags(action.repo, body.Tags || [], err));
					});
				}
				break;
			}
		}
	},
};

Dispatcher.register(RepoBackend.__onDispatch);

export default RepoBackend;
