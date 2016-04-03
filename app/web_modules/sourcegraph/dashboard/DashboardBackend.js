import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import Dispatcher from "sourcegraph/Dispatcher";
import {defaultFetch, checkStatus} from "sourcegraph/util/xhr";

const DashboardBackend = {
	fetch: defaultFetch,

	__onDispatch(action) {
		switch (action.constructor) {
		case DashboardActions.WantHome:
			{
				let home = DashboardStore.home.get();
				if (home === null) {
					DashboardBackend.fetch("/.api/home")
						.then(checkStatus)
						.then((resp) => resp.json())
						.catch((err) => {
							console.error(err);
							return {Error: true};
						})
						.then((data) => Dispatcher.Stores.dispatch(new DashboardActions.HomeFetched(data)));
				}
				break;
			}
		}
	},
};

Dispatcher.Backends.register(DashboardBackend.__onDispatch);

export default DashboardBackend;
