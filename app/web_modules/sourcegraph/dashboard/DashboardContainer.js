import React from "react";
import Helmet from "react-helmet";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import "./DashboardBackend"; // for side effects
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";
import EventLogger, {EventLocation} from "sourcegraph/util/EventLogger";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import context from "sourcegraph/app/context";

import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";

import {Button} from "sourcegraph/components";
import {GitHubIcon} from "sourcegraph/components/Icons";
import {urlToGitHubOAuth} from "sourcegraph/util/urlTo";

import deepFreeze from "sourcegraph/util/deepFreeze";
import DashboardPromos from "./DashboardPromos";
// import ChromeExtensionCTA from "./ChromeExtensionCTA";

class DashboardContainer extends Container {

	componentDidMount() {
		super.componentDidMount();
		if (this.state.githubRedirect) {
			EventLogger.logEvent("LinkGitHubCompleted");
		}
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.exampleRepos = this._exampleRepos();
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

	_exampleRepos() {
		return deepFreeze([{
			URI: "github.com/golang/go",
			Owner: "golang",
			Name: "go",
			Language: "Go",
			Examples: [
				{Functions: {
					Path: "/github.com/golang/go@master/-/def/GoPackage/net/http/-/Get",
					FunctionCallCount: "2313",
					FmtStrings: {
						Name: {
							ScopeQualified: "http.Get"}, Type: {
								ScopeQualified: "(url string) (resp *Response, err error)",
							}, NameAndTypeSeparator: "", DefKeyword: "func"}},
				},
				{Functions: {
					Path: "/github.com/golang/go@master/-/def/GoPackage/fmt/-/Sprintf",
					FunctionCallCount: "1313",
					FmtStrings: {
						Name: {ScopeQualified: "fmt.Sprintf"},
						Type: {ScopeQualified: "(format string, a ...interface{}) string"}, NameAndTypeSeparator: "", DefKeyword: "func",
					},
				},
				},
			],
		},
		]);
	}

	renderCTAButtons() {
		return (<div styleName="cta-header">
				{!context.hasLinkedGitHub && <div styleName="cta">
					<a href={urlToGitHubOAuth} onClick={() => EventLogger.logEventForPage("SubmitLinkGitHub", EventLocation.Dashboard)}>
						<Button outline={true} color="warning"><GitHubIcon styleName="github-icon" />Add My GitHub Repositories</Button>
					</a>
				</div>}
			{/* NOTE: The ChromeExtensionCTA (container) is responsible for determining whether it should render the button. */}
			{/* <ChromeExtensionCTA /> */}
		</div>);
	}

	render() {
		return (<div styleName="container">
			<Helmet title="Home" />

			{!context.currentUser &&
				<div styleName="anon-section">
					<div styleName="anon-title">Index your GitHub code</div>
					<div styleName="anon-header-sub">Web-based, IDE-like code browsing and global "find usages" for Go code.</div>
				</div>
			}

			{!context.currentUser &&
				<DashboardPromos/>
			}

			{context.currentUser &&
				<div styleName="anon-section">
					<div styleName="anon-title-left">My Dashboard</div>
					{this.renderCTAButtons()}
				</div>
			}

			{!context.currentUser &&
				<div styleName="anon-section">
					<div styleName="anon-title">Jump in with live examples</div>
					<div styleName="anon-header-sub">Select a function from the Go standard library and see its usage across all open-source libraries</div>
				</div>
			}

			<div styleName="repos">
				<DashboardRepos repos={(this.state.repos || []).concat(this.state.remoteRepos || [])}
					exampleRepos={this.state.exampleRepos}/>
			</div>

			{!context.currentUser &&
				<div styleName="cta-box">
					<a href="join" onClick={() => EventLogger.logEventForPage("JoinCTAClicked", EventLocation.Dashboard, {PageLocation: "Bottom"})}>
						<Button color="info" size="large">Add Sourcegraph to my Code</Button>
					</a>
				</div>
			}
		</div>);
	}
}

DashboardContainer.propTypes = {
};


export default CSSModules(DashboardContainer, styles);
