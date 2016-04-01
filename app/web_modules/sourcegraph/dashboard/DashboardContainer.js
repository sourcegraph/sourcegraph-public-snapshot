import React from "react";

import Container from "sourcegraph/Container";
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";
import context from "sourcegraph/context";
import Styles from "./styles/Dashboard.css";

class DashboardContainer extends Container {


	constructor(props) {
		super(props);
		const logoUrl = !global.document ? "" : document.getElementById("DashboardContainer").dataset.logo;
		const signup_url = !global.document ? "" : document.getElementById("DashboardContainer").dataset.signupurl;
		this.state = {logo: logoUrl, signup_url: signup_url};
		this._username = this._username.bind(this);
		this._userAvatar = this._userAvatar.bind(this);
	}

	componentWillUnmount() {
		super.componentWillUnmount();
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.repos = DashboardStore.repos;
		state.onboarding = DashboardStore.onboarding;
	}

	_username() {
		return context.currentUser.Name || context.currentUser.Login || "";
	}

	_userAvatar() {
		return this.state.onboarding.linkGitHub ?
			"https://assets-cdn.github.com/images/modules/logos_page/GitHub-Mark.png" :
			(context.currentUser.AvatarURL || "");
	}

	stores() { return [DashboardStore]; }

	render() {
		return (
			<div>
				<div className={Styles.dash_repos}>
				{!context.currentUser &&
					<div className={Styles.anon_section}>
						<img className={Styles.logo} src={this.state.logo} />
						<div className={Styles.anon_title}>Understand and use code better</div>
						<div className={Styles.anon_header_sub}>
							Use Sourcegraph to search, browse, and cross-reference code.
							<br />
							Works with both public and private GitHub repositories written in Go.
						</div>
					</div>
				}
				<div className={Styles.repos}>
					<DashboardRepos repos={this.state.repos}
						linkGitHub={this.state.onboarding.linkGitHub}
						linkGitHubURL={this.state.onboarding.linkGitHubURL || ""}
						signup={this.state.signup_url} />
				</div>
			</div>
		</div>);
	}
}

DashboardContainer.propTypes = {
};


export default DashboardContainer;
