import React from "react";
import update from "react/lib/update";

import Container from "sourcegraph/Container";
import "./DashboardBackend"; // for side effects
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import GitHubReposStore from "sourcegraph/dashboard/GitHubReposStore";
import GitHubUsersStore from "sourcegraph/dashboard/GitHubUsersStore";

import DashboardUsers from "sourcegraph/dashboard/DashboardUsers";
import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";

import Dispatcher from "sourcegraph/Dispatcher";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

import AlertContainer from "sourcegraph/alerts/AlertContainer";

import classNames from "classnames";

class DashboardContainer extends Container {
	constructor(props) {
		super(props);
		this.state = {
			repoName: "", // for the repo create input
		};
		this._handleRepoTextInput = this._handleRepoTextInput.bind(this);
		this._handleCreateRepo = this._handleCreateRepo.bind(this);
		this._dismissWelcome = this._dismissWelcome.bind(this);
	}

	componentDidMount() {
		super.componentDidMount();
		if (this.state.onboarding.linkGitHubRedirect) setTimeout(this._dismissWelcome, 3000);
	}

	componentWillUnmount() {
		super.componentWillUnmount();
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.repos = (DashboardStore.repos || [])
			.map(repo => update(repo, {$merge: {ExistsLocally: true}}))
			.concat(GitHubReposStore.remoteRepos.getDashboard());
		state.users = (DashboardStore.users || [])
			.map(user => ({LocalAccount: user}))
			.concat(GitHubUsersStore.users.getUnique());
		state.currentUser = DashboardStore.currentUser;
		state.onboarding = DashboardStore.onboarding;
		state.onWaitlist = DashboardStore.onWaitlist;
		state.isMothership = DashboardStore.isMothership;
		state.allowStandaloneRepos = !DashboardStore.allowMirrors;
		state.allowGitHubMirrors = DashboardStore.allowMirrors;
		state.allowStandaloneUsers = !DashboardStore.allowMirrors;
		state.allowGitHubUsers = DashboardStore.allowMirrors;
	}

	_handleRepoTextInput(e) {
		this.setState(update(this.state, {
			repoName: {$set: e.target.value},
		}));
	}

	_handleCreateRepo() {
		if (this.state.repoName === "") return;
		Dispatcher.dispatch(new DashboardActions.WantCreateRepo(this.state.repoName));
		this.setState({showCreateRepoWell: false, repoName: ""});
	}

	_dismissWelcome() {
		this.setState({dismissWelcome: true});
	}

	stores() { return [DashboardStore, GitHubReposStore, GitHubUsersStore]; }

	render() {
		const username = this.state.currentUser.Name || this.state.currentUser.Login;
		const linkGitHubAvatar = this.state.onboarding.linkGitHub ?
			"https://assets-cdn.github.com/images/modules/logos_page/GitHub-Mark.png" :
			this.state.currentUser.AvatarURL;
		const welcomeLabel = this.state.onboarding.linkGitHub ? "Link your GitHub account to get started." : `Welcome ${username.split(" ")[0]}!`;

		return (
			<div className="dashboard-container row">
				<AlertContainer />
				<div className="dash-repos col-lg-9 col-md-8">
					{(this.state.onboarding.linkGitHub || (this.state.onboarding.linkGitHubRedirect && !this.state.dismissWelcome)) &&
						<div className="well link-github-well">
							<div className="avatar-container">
								<div className="avatar-md">
									<img className={`avatar-md ${this.state.onboarding.linkGitHub ? "avatar-github" : ""}`} src={linkGitHubAvatar} />
									{this.state.onboarding.linkGitHubRedirect &&
										<div className="github-link-success-icon">
											<span className="check-icon"><i className="fa fa-check"></i></span>
										</div>
									}
								</div>
							</div>
							<strong className="link-github-label">{welcomeLabel}</strong>
							{this.state.onboarding.linkGitHub &&
								<button className="btn btn-primary link-github-button"
									onClick={() => window.location.href = this.state.onboarding.linkGitHubURL}>
									Connect
								</button>
							}
						</div>
					}
					<div className="dash-repos-header">
						<h3 className="your-repos">Repositories</h3>
						{!this.state.allowGitHubMirrors && (this.state.currentUser.Admin || this.state.currentUser.Write) &&
							<button className="btn add-repo-btn btn-default"
								onClick={_ => this.setState({showCreateRepoWell: !this.state.showCreateRepoWell})}>
								<i className={classNames("sg-icon", {
									"sg-icon-plus-box": !this.state.showCreateRepoWell,
									"sg-icon-close": this.state.showCreateRepoWell,
								})} />
							</button>
						}
					</div>
					{this.state.showCreateRepoWell && <div className="well add-repo-well">
						<div className="form-inline">
							<div className={classNames("form-group", {
								"has-error": false, // TODO: add repo name validation
							})}>
								<input className="form-control create-repo-input"
									placeholder="Repository name"
									type="text"
									value={this.state.repoName}
									onKeyPress={(e) => { if ((e.keyCode || e.which) === 13) this._handleCreateRepo(); }}
									onChange={this._handleRepoTextInput} />
							</div>
							<button className={classNames("btn btn-primary create-repo-btn", {
								disabled: this.state.repoName === "",
							})}
								onClick={this._handleCreateRepo}>CREATE</button>
						</div>
					</div>}
					<div>
						<DashboardRepos repos={this.state.repos}
							onWaitlist={this.state.onWaitlist}
							allowGitHubMirrors={this.state.allowGitHubMirrors}
							linkGitHub={this.state.onboarding.linkGitHub} />
					</div>
				</div>
				<div className="dash-users col-lg-3 col-md-4">
					<DashboardUsers users={this.state.users}
						currentUser={this.state.currentUser}
						onboarding={this.state.onboarding}
						allowStandaloneUsers={this.state.allowStandaloneUsers}
						allowGitHubUsers={this.state.allowGitHubUsers} />
				</div>
			</div>
		);
	}
}

DashboardContainer.propTypes = {
};

export default DashboardContainer;
