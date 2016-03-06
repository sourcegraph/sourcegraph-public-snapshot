import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import * as AlertActions from "sourcegraph/alerts/AlertActions";
import Dispatcher from "sourcegraph/Dispatcher";
import defaultXhr from "sourcegraph/util/xhr";

const DashboardBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {
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
				Dispatcher.dispatch(new AlertActions.AddAlert(false,
					`Please send <a href="${body.Link}">this invitation link</a> to <strong>${action.email}</strong>.`));
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
