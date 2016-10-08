import * as Dispatcher from "sourcegraph/Dispatcher";
import * as OrgActions from "sourcegraph/org/OrgActions";
import "sourcegraph/org/OrgBackend";
import {Store} from "sourcegraph/Store";
import {deepFreeze} from "sourcegraph/util/deepFreeze";

class OrgStoreClass extends Store<any> {
	orgs: any;
	members: any;

	reset(): void {
		this.orgs = null;
		this.members = deepFreeze({
			org: {},
			get(orgName: string): any {
				return this.org[orgName] || null;
			},
		});
	}

	__onDispatch(action: OrgActions.Action): void {
		if (action instanceof OrgActions.OrgsFetched) {
			this.orgs = action.data || [];

		} else if (action instanceof OrgActions.OrgMembersFetched) {
			this.members = deepFreeze(Object.assign({}, this.members, {
				org: Object.assign({}, this.members.org, {
					[action.orgName]: action.data || [],
				}),
			}));

		} else {
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export const OrgStore = new OrgStoreClass(Dispatcher.Stores);
