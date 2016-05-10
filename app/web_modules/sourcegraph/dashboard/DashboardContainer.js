import React from "react";
import Helmet from "react-helmet";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import "./DashboardBackend"; // for side effects
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";
import GlobalSearch from "sourcegraph/search/GlobalSearch";
import {EventLocation} from "sourcegraph/util/EventLogger";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";
import {urlToGitHubOAuth, urlToPrivateGitHubOAuth} from "sourcegraph/util/urlTo";

import OnboardingModals from "./OnboardingModals";
import HomeContainer from "./HomeContainer";
import GitHubAuthButton from "sourcegraph/user/GitHubAuthButton";

class DashboardContainer extends Container {
	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
		user: React.PropTypes.object,
		signedIn: React.PropTypes.bool.isRequired,
		githubToken: React.PropTypes.object,
		eventLogger: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this.state = {
			showChromeExtensionCTA: false,
		};
	}

	componentDidMount() {
		super.componentDidMount();
		setTimeout(() => this.setState({
			showChromeExtensionCTA: global.chrome && global.document && !document.getElementById("chrome-extension-installed"),
		}), 0);
	}

	reconcileState(state, props, context) {
		Object.assign(state, props);
		state.repos = DashboardStore.repos || null;
		state.remoteRepos = DashboardStore.remoteRepos || null;

		state.signedIn = context.signedIn;
		state.githubToken = context.githubToken;
		state.user = context.user;

		if (props.location && props.location.state) {
			state.onboardingExperience = props.location.state["_onboarding"] && state.signedIn ? props.location.state["_onboarding"] : null;
		}
	}

	onStateTransition(prevState, nextState) {
		if (nextState.repos === null && nextState.repos !== prevState.repos) {
			Dispatcher.Backends.dispatch(new DashboardActions.WantRepos());
		}
		if (nextState.remoteRepos === null && nextState.remoteRepos !== prevState.remoteRepos) {
			Dispatcher.Backends.dispatch(new DashboardActions.WantRemoteRepos());
		}
	}

	stores() { return [DashboardStore]; }

	renderCTAButtons() {
		return (
			<div>
				{!this.context.githubToken && <div styleName="cta">
				<a href={urlToGitHubOAuth} onClick={() => this.context.eventLogger.logEventForPage("InitiateGitHubOAuth2Flow", EventLocation.Dashboard, {scopes: "", upgrade: true})}>
						<GitHubAuthButton>Link GitHub account</GitHubAuthButton>
					</a>
				</div>}
				{this.context.githubToken && (!this.context.githubToken.scope || !(this.context.githubToken.scope.includes("repo") && this.context.githubToken.scope.includes("read:org") && this.context.githubToken.scope.includes("user:email"))) && <div styleName="cta">
					<a href={urlToPrivateGitHubOAuth}
						onClick={() => this.context.eventLogger.logEventForPage("InitiateGitHubOAuth2Flow", EventLocation.Dashboard, {scopes: "read:org,repo,user:email", upgrade: true})}>
						<GitHubAuthButton >Use with private repositories</GitHubAuthButton>
					</a>
				</div>}
			</div>
		);
	}

	render() {
		return (
			<div styleName="container">
				{this.state.onboardingExperience && <OnboardingModals location={this.state.location} onboardingFlow={this.state.onboardingExperience} canShowChromeExtensionCTA={this.state.showChromeExtensionCTA}/>}

				<Helmet title="Home" />

				{!this.context.signedIn && <HomeContainer location={this.props.location} />}

				{this.context.user && this.context.user.Admin &&
					<GlobalSearch query={this.props.location.query.q || ""}/>
				}

				{this.context.signedIn &&
					<div styleName="anon-section">
						{this.renderCTAButtons()}
					</div>
				}

				{this.context.signedIn && <div styleName="repos">
					<DashboardRepos repos={(this.state.repos || []).concat(this.state.remoteRepos || [])} />
				</div>}
			</div>
		);
	}
}

export default CSSModules(DashboardContainer, styles);
