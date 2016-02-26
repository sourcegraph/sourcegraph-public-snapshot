import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import Dispatcher from "sourcegraph/Dispatcher";
import defaultXhr from "sourcegraph/util/xhr";

const DashboardBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {
		case DashboardActions.WantCreateRepo:
			DashboardBackend.xhr({
				uri: `/.ui/.repo-create?RepoURI=${action.name}`,
				method: "POST",
				json: {},
			}, function(err, resp, body) {
				if (err) {
					// TODO: some proper error handling
					console.error(err);
					return;
				}
				if (resp.statusCode !== 200) {
					// TODO: some proper error handling
					console.log(resp);
					return;
				}
				Dispatcher.dispatch(new DashboardActions.RepoCreated(body));
			});
			break;
		case DashboardActions.WantAddMirrorRepos:
			DashboardBackend.xhr({
				uri: `/.ui/.repo-mirror`,
				method: "POST",
				json: {
					Repos: action.repos,
				},
			}, function(err, resp, body) {
				if (err) {
					// TODO: some proper error handling
					console.error(err);
					return;
				}
				if (resp.statusCode !== 200) {
					// TODO: some proper error handling
					console.log(resp);
					return;
				}
				Dispatcher.dispatch(new DashboardActions.MirrorReposAdded(body));
			});
			break;
		case DashboardActions.WantAddMirrorRepo:
			console.log("want add mirror repo", action.repo);
			DashboardBackend.xhr({
				uri: `/.ui/.repo-mirror`,
				method: "POST",
				json: {
					Repos: [action.repo],
				},
			}, function(err, resp, body) {
				if (err) {
					// TODO: some proper error handling
					console.error(err);
					return;
				}
				if (resp.statusCode !== 200) {
					// TODO: some proper error handling
					console.log(resp);
					return;
				}
				Dispatcher.dispatch(new DashboardActions.MirrorRepoAdded(action.repo, body));
			});
			break;
		case DashboardActions.WantInviteUser:
			DashboardBackend.xhr({
				uri: `/.ui/.invite`,
				method: "POST",
				json: {
					Email: action.email,
					Permission: action.permission,
				},
			}, function(err, resp, body) {
				if (err) {
					console.error(err);
					return;
				}
				if (resp.statusCode !== 200) {
					// TODO: some proper error handling
					console.log(resp);
					return;
				}
				Dispatcher.dispatch(new DashboardActions.UserInvited({
					Name: action.email,
					Admin: action.permission === "admin",
					Write: action.permission === "write",
				}));
				// TODO: proper modal....but soon we're sending emails anyway.
				alert(`Invite link: ${body.Link}`);
			});
			break;
		case DashboardActions.WantInviteUsers:
			DashboardBackend.xhr({
				uri: `/.ui/.invite-bulk`,
				method: "POST",
				json: {
					Emails: action.emails,
				},
			}, function(err, resp, body) {
				if (err) {
					console.error(err);
					return;
				}
				if (resp.statusCode !== 200) {
					// TODO: some proper error handling
					console.log(resp);
					return;
				}
				Dispatcher.dispatch(new DashboardActions.UsersInvited(body));
			});
			break;
		}
	},
};

Dispatcher.register(DashboardBackend.__onDispatch);

export default DashboardBackend;
