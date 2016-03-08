import React from "react";

import Container from "sourcegraph/Container";
import "./DashboardBackend"; // for side effects
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import GitHubUsersStore from "sourcegraph/dashboard/GitHubUsersStore";

import DashboardUsers from "sourcegraph/dashboard/DashboardUsers";
import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";

import AlertContainer from "sourcegraph/alerts/AlertContainer";

class DashboardContainer extends Container {
	constructor(props) {
		super(props);
		this._username = this._username.bind(this);
		this._userAvatar = this._userAvatar.bind(this);
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
		state.repos = DashboardStore.repos;
		state.users = GitHubUsersStore.users.getUnique();
		state.currentUser = DashboardStore.currentUser || {}; // empty if anonymous user
		state.onboarding = DashboardStore.onboarding;
	}

	_username() {
		return this.state.currentUser.Name || this.state.currentUser.Login || "";
	}

	_userAvatar() {
		return this.state.onboarding.linkGitHub ?
			"https://assets-cdn.github.com/images/modules/logos_page/GitHub-Mark.png" :
			(this.state.currentUser.AvatarURL || "");
	}

	_dismissWelcome() {
		this.setState({dismissWelcome: true});
	}

	stores() { return [DashboardStore, GitHubUsersStore]; }

	render() {
		return (
			<div className="dashboard-container row">
				<AlertContainer />
				<div className="dash-repos col-lg-9 col-md-8">
					{(this.state.currentUser.Login && this.state.onboarding.linkGitHub ||
						(this.state.onboarding.linkGitHubRedirect && !this.state.dismissWelcome)) &&
						<div className="well link-github-well">
							<div className="avatar-container">
								<div className="avatar-md">
									<img className={`avatar-md ${this.state.onboarding.linkGitHub ? "avatar-github" : ""}`} src={this._userAvatar()} />
									{this.state.onboarding.linkGitHubRedirect &&
										<div className="github-link-success-icon">
											<span className="check-icon"><i className="fa fa-check"></i></span>
										</div>
									}
								</div>
							</div>
							<strong className="link-github-label">{this.state.onboarding.linkGitHub ?
								"Link your GitHub account to get started." :
								`Welcome ${this._username().split(" ")[0]}!`
							}</strong>
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
							linkGitHub={this.state.onboarding.linkGitHub} />
					</div>
				</div>
				<div className="dash-users col-lg-3 col-md-4">
					<DashboardUsers users={this.state.users}
						currentUser={this.state.currentUser}
						onboarding={this.state.onboarding} />
				</div>
			</div>
		);
	}
}

DashboardContainer.propTypes = {
};

export default DashboardContainer;
