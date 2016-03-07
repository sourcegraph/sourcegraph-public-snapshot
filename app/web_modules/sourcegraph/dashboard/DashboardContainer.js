import React from "react";

import Container from "sourcegraph/Container";
import "./DashboardBackend"; // for side effects
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import GitHubReposStore from "sourcegraph/dashboard/GitHubReposStore";
import GitHubUsersStore from "sourcegraph/dashboard/GitHubUsersStore";

import DashboardUsers from "sourcegraph/dashboard/DashboardUsers";
import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";

import AlertContainer from "sourcegraph/alerts/AlertContainer";

class DashboardContainer extends Container {
	constructor(props) {
		super(props);
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
		state.repos = (DashboardStore.repos || []).concat(GitHubReposStore.repos);
		state.remoteRepos = (GitHubReposStore.remoteRepos || []);
		state.users = (DashboardStore.users || [])
			.map(user => ({LocalAccount: user}))
			.concat(GitHubUsersStore.users.getUnique());
		state.currentUser = DashboardStore.currentUser || {}; // empty if anonymous user
		state.onboarding = DashboardStore.onboarding;
		state.onWaitlist = DashboardStore.onWaitlist;
		state.isMothership = DashboardStore.isMothership;
		state.allowStandaloneRepos = !DashboardStore.allowMirrors;
		state.allowGitHubMirrors = DashboardStore.allowMirrors;
		state.allowStandaloneUsers = !DashboardStore.allowMirrors;
		state.allowGitHubUsers = DashboardStore.allowMirrors;
	}

	_dismissWelcome() {
		this.setState({dismissWelcome: true});
	}

	stores() { return [DashboardStore, GitHubReposStore, GitHubUsersStore]; }

	render() {
		const username = this.state.currentUser.Name || this.state.currentUser.Login || "";
		const linkGitHubAvatar = this.state.onboarding.linkGitHub ?
			"https://assets-cdn.github.com/images/modules/logos_page/GitHub-Mark.png" :
			(this.state.currentUser.AvatarURL || "");
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
					</div>
					<div>
						<DashboardRepos repos={this.state.repos}
							remoteRepos={this.state.remoteRepos}
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
