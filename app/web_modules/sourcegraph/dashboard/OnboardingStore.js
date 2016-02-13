import {Store} from "flux/utils";

import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import * as OnboardingActions from "sourcegraph/dashboard/OnboardingActions";
import EntityTypes from "sourcegraph/dashboard/EntityTypes";

import update from "react/lib/update";

export class OnboardingStore extends Store {
	constructor(dispatcher) {
		super(dispatcher);
		this.progress = deepFreeze(window.progress);

		this.selectAll = false;
		let selectedRepos = {};
		let selectedUsers = {};
		if (window.mirrorRepos) window.mirrorRepos.forEach(repo => selectedRepos[repo.index] = false);
		if (window.users) window.users.forEach(user => selectedUsers[user.index] = false);
		this.selectedRepos = deepFreeze(selectedRepos);
		this.selectedUsers = deepFreeze(selectedUsers);

		let orgs = {};
		if (window.mirrorRepos) window.mirrorRepos.forEach(repo => orgs[repo.org] = 1);
		if (window.users) window.users.forEach(user => orgs[user.org] = 1);
		this.orgs = Object.keys(orgs);
		this.currentOrg = this.orgs.length > 0 ? this.orgs[0] : null;
		this.currentType = EntityTypes.REPO;
	}

	__onDispatch(action) {
		switch (action.constructor) {

		case DashboardActions.ReposAdded:
			break;

		case DashboardActions.UsersAdded:
			break;

		case OnboardingActions.AdvanceProgressStep:
			this.progress = update(this.progress, {
				currentStep: {$set: this.progress.currentStep + 1},
			});
			this.selectAll = false;
			if (this.progress.currentStep === 3) {
				this.currentType = EntityTypes.USER;
			}
			break;

		case OnboardingActions.SelectOrg:
			this.currentOrg = action.org;
			this.selectAll = false;
			break;

		case OnboardingActions.SelectItems:
			{
				let updateQuery = {};
				action.items.forEach(item => updateQuery[item.index] = {$set: action.selectAll});
				if (action.type === EntityTypes.REPO) {
					this.selectedRepos = update(this.selectedRepos, updateQuery);
				} else if (action.type === EntityTypes.USER) {
					this.selectedUsers = update(this.selectedUsers, updateQuery);
				}
				this.selectAll = action.selectAll;
				break;
			}

		case OnboardingActions.SelectItem:
			{
				let updateQuery = {};
				updateQuery[action.itemIndex] = {$set: action.select};
				if (action.type === EntityTypes.REPO) {
					this.selectedRepos = update(this.selectedRepos, updateQuery);
				} else if (action.type === EntityTypes.USER) {
					this.selectedUsers = update(this.selectedUsers, updateQuery);
				}
				break;
			}

		default:
			return; // don't emit change
		}

		this.__emitChange();
	}
}

export default new OnboardingStore(Dispatcher);
