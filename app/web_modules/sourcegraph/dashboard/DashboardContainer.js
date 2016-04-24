import React from "react";
import {Link} from "react-router";
import Helmet from "react-helmet";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import "./DashboardBackend"; // for side effects
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";
import GlobalSearch from "sourcegraph/search/GlobalSearch";
import EventLogger, {EventLocation} from "sourcegraph/util/EventLogger";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import context from "sourcegraph/app/context";

import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";

import {Button} from "sourcegraph/components";
import {GitHubIcon} from "sourcegraph/components/Icons";
import {urlToGitHubOAuth} from "sourcegraph/util/urlTo";

import ChromeExtensionCTA from "./ChromeExtensionCTA";

class DashboardContainer extends Container {
	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this.state = {
			showChromeExtensionCTA: false,
		};
	}

	componentDidMount() {
		super.componentDidMount();
		if (this.state.githubRedirect) {
			EventLogger.logEvent("LinkGitHubCompleted");
		}
		setTimeout(() => this.setState({
			showChromeExtensionCTA: global.chrome && global.document && !document.getElementById("chrome-extension-installed"),
		}), 0);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.repos = DashboardStore.repos || null;
		state.remoteRepos = DashboardStore.remoteRepos || null;
		state.githubRedirect = props.location && props.location.query ? (props.location.query["github-onboarding"] || false) : false;
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
				{!context.hasLinkedGitHub && <div styleName="cta">
					<a href={urlToGitHubOAuth} onClick={() => EventLogger.logEventForPage("SubmitLinkGitHub", EventLocation.Dashboard)}>
						<Button outline={true} color="warning"><GitHubIcon style={{marginRight: "10px", fontSize: "16px"}} />Add My GitHub Repositories</Button>
					</a>
				</div>}
			</div>
		);
	}

	render() {
		return (
			<div styleName="container">
				<Helmet title="Home" />

				{!context.currentUser &&
					<div styleName="anon-section">
						<div styleName="anon-title"><img src={`${this.context.siteConfig.assetsRoot}/img/sourcegraph-logo.svg`}/></div>
						<div styleName="anon-header-sub">Save time and code better with live usage examples.</div>
					</div>
				}
				{!context.currentUser &&
					<div styleName="cta-box">
						<div styleName="cta-headline">See everywhere a Go function is called, globally.</div>
						<Link to="github.com/golang/go/-/def/GoPackage/net/http/-/NewRequest/-/info" onClick={() => EventLogger.logEvent("GoHTTPDefRefsCTAClicked")}>
							<Button color="primary" size="large" unspaced={true} lowercase={true}>See usage examples for http.NewRequest &raquo;</Button>
						</Link>
						<div styleName="cta-subline">
							<Link styleName="cta-link" to="join">Sign up for private code</Link>
							{this.state.showChromeExtensionCTA && <span>|</span>}
							{this.state.showChromeExtensionCTA && <ChromeExtensionCTA />}
						</div>
					</div>
				}

				{context.currentUser &&
					<div styleName="anon-section">
						{this.renderCTAButtons()}
					</div>
				}

				{context.currentUser && context.currentUser.Admin &&
					<GlobalSearch query={this.props.location.query.q || ""}/>
				}

				{context.currentUser && <div styleName="repos">
					<DashboardRepos repos={(this.state.repos || []).concat(this.state.remoteRepos || [])} />
				</div>}
			</div>
		);
	}
}

export default CSSModules(DashboardContainer, styles);
