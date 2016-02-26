import React from "react";
import update from "react/lib/update";

import Container from "sourcegraph/Container";
import GitHubUsersStore from "sourcegraph/dashboard/GitHubUsersStore";
import SelectableListWidget from "sourcegraph/dashboard/SelectableListWidget";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import Dispatcher from "sourcegraph/Dispatcher";

class ImportGitHubUsersMenu extends Container {
	constructor(props) {
		super(props);
		this.state = {
			currentOrg: null,
			selectedUsers: {},
			selectAll: false,
			showLoading: false,
		};
		this._handleAddUsers = this._handleAddUsers.bind(this);
		this._handleSelect = this._handleSelect.bind(this);
		this._handleSelectAll = this._handleSelectAll.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.orgs = GitHubUsersStore.orgs;
		if (!state.currentOrg) state.currentOrg = GitHubUsersStore.orgs[0];
		state.items = GitHubUsersStore.getByOrg[state.currentOrg]
			.filter(this._canSelect)
			.map((user) => ({
				name: user.RemoteAccount.Name ? `${user.RemoteAccount.Login} (${user.RemoteAccount.Name})` : user.RemoteAccount.Login,
				key: user.RemoteAccount.Login,
			}));
		state.unselectableItems = GitHubUsersStore.users.getByOrg(state.currentOrg)
			.filter(user => !this._canSelect(user))
			.map((user) => ({
				name: user.RemoteAccount.Name ? `${user.RemoteAccount.Login} (${user.RemoteAccount.Name})` : user.RemoteAccount.Login,
				key: user.RemoteAccount.Login,
				reason: this._unselectableReason(user),
			}));
		state.showLoading = GitHubUsersStore.showLoading;
		state.allItems = GitHubUsersStore.users.users;
	}

	_canSelect(user) {
		if (!user.Email) return false;
		if (user.LocalAccount) return false;
		if (user.IsInvited) return false;
		return true;
	}

	_unselectableReason(user) {
		if (!user.Email) return "cannot find email";
		if (user.LocalAccount) return "user account exists";
		if (user.IsInvited) return "user already invited";
		return ""; // unknown reason
	}

	_handleSelect(login, select) {
		this.setState(update(this.state, {
			selectedUsers: {$merge: {[login]: select}},
		}));
	}

	_handleSelectAll(items, selectAll) {
		let selected = {};
		for (let item of items) {
			selected[item.key] = selectAll;
		}
		this.setState(update(this.state, {
			selectAll: {$set: selectAll},
			selectedUsers: {$merge: selected},
		}));
	}

	_handleAddUsers(items) {
		let emails = [];
		for (let login of Object.keys(this.state.selectedUsers)) {
			if (this.state.selectedUsers[login]) {
				let user = GitHubUsersStore.users.get(login);
				if (user && user.Email) emails.push(user.Email);
			}
		}
		if (emails.length > 0) {
			Dispatcher.dispatch(new DashboardActions.WantInviteUsers(emails));
		} else {
			console.log("No emails for selected users");
		}
		Dispatcher.dispatch(new DashboardActions.DismissUsersModal());
	}

	stores() { return [GitHubUsersStore]; }

	render() {
		return (
			<SelectableListWidget items={this.state.items}
				unselectableItems={this.state.unselectableItems}
				allItems={this.state.allItems}
				currentCategory={this.state.currentOrg}
				menuCategories={this.state.orgs}
				onMenuClick={(org) => this.setState({currentOrg: org, selectAll: false})}
				selections={this.state.selectedUsers}
				onSelect={this._handleSelect}
				onSelectAll={this._handleSelectAll}
				selectAll={this.state.selectAll}
				onSubmit={this._handleAddUsers}
				searchPlaceholderText={"Search GitHub contacts"}
				menuLabel="organizations"
				showLoading={this.state.showLoading} />
		);
	}
}

export default ImportGitHubUsersMenu;
