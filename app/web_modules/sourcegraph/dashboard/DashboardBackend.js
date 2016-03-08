import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import Dispatcher from "sourcegraph/Dispatcher";
import defaultXhr from "sourcegraph/util/xhr";

const DashboardBackend = {
	xhr: defaultXhr,

	__onDispatch(action) {
		switch (action.constructor) {
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
