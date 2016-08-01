import * as RepoActions from "sourcegraph/repo/RepoActions";
import Dispatcher from "sourcegraph/Dispatcher";

// updateRepoCloning updates the cloning status of a repository based on the response status code.
// 200 sets it as "not cloning", and 202 sets it to "cloning in progress".
//
// updateRepoCloning is intended to be chained in a fetch call. For example:
//   fetch(...).then(updateRepoCloning(actions.repo)) ...
export function updateRepoCloning(repo: string): (resp: Response) => Promise<Response> | Response {
	return (resp) => {
		if (resp.status === 200) {
			Dispatcher.Stores.dispatch(new RepoActions.RepoCloning(repo, false));
		} else if (resp.status === 202) {
			Dispatcher.Stores.dispatch(new RepoActions.RepoCloning(repo, true));
		}
		return resp;
	};
}
