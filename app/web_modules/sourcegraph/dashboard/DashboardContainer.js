import React from "react";

import Container from "sourcegraph/Container";
import "./DashboardBackend"; // for side effects
import DashboardStore from "sourcegraph/dashboard/DashboardStore";

import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";

import Styles from "./styles/Dashboard.css";

class DashboardContainer extends Container {


	constructor(props) {
		super(props);
		const logoUrl = !global.document ? "" : document.getElementById("DashboardContainer").dataset.logo;
		const signup_url = !global.document ? "" : document.getElementById("DashboardContainer").dataset.signupurl;
		this.state = {logo: logoUrl, signup_url: signup_url};
		this._username = this._username.bind(this);
		this._userAvatar = this._userAvatar.bind(this);
		this._dismissWelcome = this._dismissWelcome.bind(this);
	}

	componentDidMount() {
		super.componentDidMount();
		if (this.state.onboarding.linkGitHubRedirect) setTimeout(this._dismissWelcome, 5000);
	}

	componentWillUnmount() {
		super.componentWillUnmount();
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.repos = DashboardStore.repos;
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

	stores() { return [DashboardStore]; }

	render() {
		return (
			<div className="dashboard-container row">
				<div className="dash-repos col-lg-8 col-md-8 col-lg-offset-2 col-md-offset-2">
				{!this.state.currentUser.Login &&
					<div>
						<img className={Styles.logo} src={this.state.logo}/>
						<div className={Styles.anon_title}>Understand and use code better</div>
						<div className={Styles.anon_header_sub}>Use Sourcegraph to search, browse, and cross-reference code. <br />
						Works with both public and private GitHub repositories written in Go.
						</div>
					</div>
				}
					<div>
						<DashboardRepos repos={this.state.repos}
							linkGitHub={this.state.onboarding.linkGitHub}
							linkGitHubURL={this.state.onboarding.linkGitHubURL || ""}
							onboarding={this.state.onboarding} signup={this.state.signup_url} />
					</div>
				</div>
			</div>
		);
	}
}

DashboardContainer.propTypes = {
};


export default DashboardContainer;
