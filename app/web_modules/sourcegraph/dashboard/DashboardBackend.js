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
					console.error(err);
					return;
				}
				// TODO dispath repo created action here.
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
					console.error(err);
					return;
				}
				Dispatcher.dispatch(new DashboardActions.MirrorReposAdded(action.repos));
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
				// TODO: proper modal.
				console.log(body, resp);
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
				// TODO dispath user invited action here.
				Dispatcher.dispatch(new DashboardActions.UsersInvited(action.emails));
			});
			break;
		}
	},
};

Dispatcher.register(DashboardBackend.__onDispatch);

export default DashboardBackend;
