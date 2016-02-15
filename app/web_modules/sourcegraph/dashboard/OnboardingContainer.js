import React from "react";

import Container from "sourcegraph/Container";

import OnboardingStore from "sourcegraph/dashboard/OnboardingStore";
import GitHubReposStore from "sourcegraph/dashboard/GitHubReposStore";
import GitHubUsersStore from "sourcegraph/dashboard/GitHubUsersStore";
import SelectableListWidget from "sourcegraph/dashboard/SelectableListWidget";
import LinkGitHubWelcome from "sourcegraph/dashboard/LinkGitHubWelcome";
import ProgressBar from "sourcegraph/dashboard/ProgressBar";

import Dispatcher from "sourcegraph/Dispatcher";

import * as OnboardingActions from "sourcegraph/dashboard/OnboardingActions";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

class OnboardingContainer extends Container {
	constructor(props) {
		super(props);
		this._handleMenuClick = this._handleMenuClick.bind(this);
		this._handleItemSelect = this._handleItemSelect.bind(this);
		this._handleSelectAll = this._handleSelectAll.bind(this);
		this._handleWidgetSubmit = this._handleWidgetSubmit.bind(this);
	}

	componentDidMount() {
		super.componentDidMount();
	}

	componentWillUnmount() {
		super.componentWillUnmount();
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.progress = OnboardingStore.progress;
		state.currentUser = OnboardingStore.currentUser;
		switch (state.progress.currentStep) {
		case 2:
			state.currentOrg = GitHubReposStore.currentOrg;
			state.orgs = GitHubReposStore.orgs;
			state.selectAll = GitHubReposStore.selectAll;
			state.selections = GitHubReposStore.selectedRepos;
			state.items = GitHubReposStore.reposByOrg.get(state.currentOrg)
				.filter(repo => repo.Repo.Private)
				.map(repo => ({name: repo.Repo.Name, key: repo.Repo.URI}));
			break;

		case 3:
			state.currentOrg = GitHubUsersStore.currentOrg;
			state.orgs = GitHubUsersStore.orgs;
			state.selectAll = GitHubUsersStore.selectAll;
			state.selections = GitHubUsersStore.selectedUsers;
			state.items = GitHubUsersStore.usersByOrg.get(state.currentOrg).map((user) => ({
				name: user.RemoteAccount.Name ? `${user.RemoteAccount.Login} (${user.RemoteAccount.Name})` : user.RemoteAccount.Login,
				key: user.RemoteAccount.Login,
			}));
			break;

		default:
			break;
		}
	}

	_handleMenuClick(org) {
		switch (this.state.progress.currentStep) {
		case 2:
			Dispatcher.dispatch(new DashboardActions.SelectRepoOrg(org));
			break;
		case 3:
			Dispatcher.dispatch(new DashboardActions.SelectUserOrg(org));
			break;
		default:
			break;
		}
	}

	_handleItemSelect(itemKey, select) {
		switch (this.state.progress.currentStep) {
		case 2:
			Dispatcher.dispatch(new DashboardActions.SelectRepo(itemKey, select));
			break;
		case 3:
			Dispatcher.dispatch(new DashboardActions.SelectUser(itemKey, select));
			break;
		default:
			break;
		}
	}

	_handleSelectAll(items, selectAll) {
		switch (this.state.progress.currentStep) {
		case 2:
			Dispatcher.dispatch(new DashboardActions.SelectRepos(items.map(item => item.key), selectAll));
			break;
		case 3:
			Dispatcher.dispatch(new DashboardActions.SelectUsers(items.map(item => item.key), selectAll));
			break;
		default:
			break;
		}
	}

	_handleWidgetSubmit(items) {
		switch (this.state.progress.currentStep) {
		case 2:
			Dispatcher.dispatch(new DashboardActions.WantAddRepos(items));
			break;
		case 3:
			Dispatcher.dispatch(new DashboardActions.WantAddUsers(items));
			break;
		default:
			break;
		}
	}

	stores() { return [OnboardingStore, GitHubReposStore, GitHubUsersStore]; }

	render() {
		if (this.state.progress.currentStep >= this.state.progress.numSteps) return null;

		const panelBody = this.state.progress.currentStep <= 1 ?
			<LinkGitHubWelcome progress={this.state.progress} currentUser={this.state.currentUser}/> : (
			<div>
				<p className="header-text normal-header">
					Select Repositories
				</p>
				<p className="normal-text">
					Sourcegraph's Code Intelligence currently supports Go and Java (with more languages coming soon!)
				</p>
				<SelectableListWidget items={this.state.items}
					currentCategory={this.state.currentOrg}
					menuCategories={this.state.orgs}
					onMenuClick={this._handleMenuClick}
					selections={this.state.selections}
					selectAll={this.state.selectAll}
					onSelect={this._handleItemSelect}
					onSelectAll={this._handleSelectAll}
					onSubmit={this._handleWidgetSubmit}
					searchPlaceholderText={`Search GitHub ${this.state.progress.currentStep === 2 ? "repositories" : "contacts"}`}
					menuLabel="organizations" />
					<p className="next-step">
						<a onClick={(e) => {
							e.preventDefault();
							Dispatcher.dispatch(new OnboardingActions.AdvanceProgressStep());
						}}>i'll do that later</a>
					</p>
			</div>);

		return (
			<div className="onboarding-container">
				<div className="modal"
					tabIndex="-1">
					<div className="modal-dialog">
						<div className="modal-content">
							<div className={`modal-header modal-header-${this.state.progress.currentStep}`}>
								<ProgressBar numSteps={this.state.progress.numSteps} currentStep={this.state.progress.currentStep}/>
							</div>
							<div className="modal-body">
								{panelBody}
							</div>
						</div>
					</div>
				</div>
			</div>
		);
	}
}

OnboardingContainer.propTypes = {
};

export default OnboardingContainer;
