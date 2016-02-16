import React from "react";
import update from "react/lib/update";

import Container from "sourcegraph/Container";
import GitHubUsersStore from "sourcegraph/dashboard/GitHubUsersStore";
import SelectableListWidget from "sourcegraph/dashboard/SelectableListWidget";
// import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
// import Dispatcher from "sourcegraph/Dispatcher";

class ImportGitHubUsersMenu extends Container {
	constructor(props) {
		super(props);
		this.state = {
			currentOrg: null,
			selectedUsers: {},
			selectAll: false,
		};
		this._handleAddUsers = this._handleAddUsers.bind(this);
		this._handleSelect = this._handleSelect.bind(this);
		this._handleSelectAll = this._handleSelectAll.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.orgs = GitHubUsersStore.orgs;
		if (!state.currentOrg) state.currentOrg = GitHubUsersStore.orgs[0];
		state.items = GitHubUsersStore.usersByOrg.get(state.currentOrg).map((user) => ({
			name: user.RemoteAccount.Name ? `${user.RemoteAccount.Login} (${user.RemoteAccount.Name})` : user.RemoteAccount.Login,
			key: user.RemoteAccount.Login,
		}));
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
		// TODO:
		console.log(this.state.selectedUsers);
		// Dispatcher.dispatch(new DashboardActions.WantAddUsers(Object.keys(this.state.selectedUsers)));
	}

	stores() { return [GitHubUsersStore]; }

	render() {
		return (
			<SelectableListWidget items={this.state.items}
				currentCategory={this.state.currentOrg}
				menuCategories={this.state.orgs}
				onMenuClick={(org) => this.setState({currentOrg: org, selectAll: false})}
				selections={this.state.selectedUsers}
				onSelect={this._handleSelect}
				onSelectAll={this._handleSelectAll}
				selectAll={this.state.selectAll}
				onSubmit={this._handleAddUsers}
				searchPlaceholderText={"Search GitHub contacts"}
				menuLabel="organizations" />
		);
	}
}

export default ImportGitHubUsersMenu;
