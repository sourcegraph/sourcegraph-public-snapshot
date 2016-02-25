import React from "react";
import update from "react/lib/update";

import Container from "sourcegraph/Container";
import GitHubReposStore from "sourcegraph/dashboard/GitHubReposStore";
import SelectableListWidget from "sourcegraph/dashboard/SelectableListWidget";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import Dispatcher from "sourcegraph/Dispatcher";

class ImportGitHubReposMenu extends Container {
	constructor(props) {
		super(props);
		this.state = {
			currentOrg: null,
			selectedRepos: {},
			selectAll: false,
			showLoading: false,
		};
		this._canMirror = this._canMirror.bind(this);
		this._handleAddMirrors = this._handleAddMirrors.bind(this);
		this._handleSelect = this._handleSelect.bind(this);
		this._handleSelectAll = this._handleSelectAll.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.orgs = GitHubReposStore.orgs;
		state.onWaitlist = GitHubReposStore.onWaitlist;
		if (!state.currentOrg) state.currentOrg = GitHubReposStore.orgs[0];
		state.items = GitHubReposStore.reposByOrg.get(state.currentOrg)
			.filter(repo => this._canMirror(repo, state.onWaitlist))
			.map(repo => ({name: repo.Repo.Name, key: repo.Repo.URI, isPrivate: Boolean(repo.Repo.Private)}));
		state.unselectableItems = GitHubReposStore.reposByOrg.get(state.currentOrg)
			.filter(repo => !this._canMirror(repo, state.onWaitlist))
			.map((repo, i) => ({name: repo.Repo.Name, key: `${i}`, reason: (() => {
				if (state.onWaitlist && repo.Repo.Private) return "private repositories coming soon";
				if (repo.ExistsLocally) return "already mirrored";
				return `${repo.Repo.Language || ""} coming soon`;
			})(),
			}));
		state.showLoading = GitHubReposStore.showLoading;
	}

	_canMirror(repo, onWaitlist) {
		if (onWaitlist) {
			if (repo.Repo.Private) return false;
		}
		if (repo.ExistsLocally) return false;
		return repo.Repo.Language === "Go" || repo.Repo.Language === "Java";
	}

	_handleSelect(repoURI, select) {
		this.setState(update(this.state, {
			selectedRepos: {$merge: {[repoURI]: select}},
		}));
	}

	_handleSelectAll(items, selectAll) {
		let selected = {};
		for (let item of items) {
			selected[item.key] = selectAll;
		}
		this.setState(update(this.state, {
			selectAll: {$set: selectAll},
			selectedRepos: {$merge: selected},
		}));
	}

	_handleAddMirrors(items) {
		let repos = this.state.items.filter(repo => this.state.selectedRepos[repo.key]).map(repo => ({
			URI: repo.key,
			Private: repo.isPrivate,
		}));
		Dispatcher.dispatch(new DashboardActions.WantAddMirrorRepos(repos));
	}

	stores() { return [GitHubReposStore]; }

	render() {
		return (
			<SelectableListWidget items={this.state.items}
				unselectableItems={this.state.unselectableItems}
				currentCategory={this.state.currentOrg}
				menuCategories={this.state.orgs}
				onMenuClick={(org) => this.setState({currentOrg: org, selectAll: false})}
				onSelect={this._handleSelect}
				onSelectAll={this._handleSelectAll}
				selections={this.state.selectedRepos}
				selectAll={this.state.selectAll}
				menuLabel="Organizations"
				searchPlaceholderText="Find a repository..."
				onSubmit={this._handleAddMirrors}
				showLoading={this.state.showLoading} />
		);
	}
}

export default ImportGitHubReposMenu;
