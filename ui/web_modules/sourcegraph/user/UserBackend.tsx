import * as Dispatcher from "sourcegraph/Dispatcher";
import * as UserActions from "sourcegraph/user/UserActions";
import {checkStatus, defaultFetch} from "sourcegraph/util/xhr";

class UserBackendClass {
	fetch: any;

	constructor() {
		this.fetch = defaultFetch;
	}

	__onDispatch(payload: UserActions.Action): void {
		if (payload instanceof UserActions.SubmitBetaSubscription) {
			const action = payload;
			this.fetch(`/.api/beta-subscription`, {
				method: "POST",
				body: JSON.stringify({
					Email: action.email,
					FirstName: action.firstName,
					LastName: action.lastName,
					Languages: action.languages,
					Editors: action.editors,
					Message: action.message,
				}),
			})
				.then(checkStatus)
				.then((resp) => resp.json())
				.catch((err) => ({Error: err}))
				.then(function(data: any): void {
					Dispatcher.Stores.dispatch(new UserActions.BetaSubscriptionCompleted(data));
				});
		}
	}
};

export const UserBackend = new UserBackendClass();
Dispatcher.Backends.register(UserBackend.__onDispatch.bind(UserBackend));
