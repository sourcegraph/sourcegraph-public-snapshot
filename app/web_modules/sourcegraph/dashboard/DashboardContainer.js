import React from "react";
import Helmet from "react-helmet";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import "./DashboardBackend"; // for side effects
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";
import EventLogger from "sourcegraph/util/EventLogger";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import context from "sourcegraph/app/context";

import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";

class DashboardContainer extends Container {
	constructor(props) {
		super(props);
	}

	componentDidMount() {
		super.componentDidMount();
		if (this.state.githubRedirect) {
			EventLogger.logEvent("LinkGitHubCompleted");
		}
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.exampleRepos = DashboardStore.exampleRepos || null;
		state.repos = DashboardStore.repos || null;
		state.remoteRepos = DashboardStore.remoteRepos || null;
		state.hasLinkedGitHub = null; // special condition to avoid flashing CTA
		if (DashboardStore.hasLinkedGitHub === false) state.hasLinkedGitHub = false; // show CTA
		if (DashboardStore.hasLinkedGitHub) state.hasLinkedGitHub = true; // don't show CTA
		state.githubRedirect = props.location && props.location.query ? (props.location.query["github-onboarding"] || false) : false;
	}

	onStateTransition(prevState, nextState) {repos
		if (nextState.repos === null && nextState. !== prevState.repos) {
			Dispatcher.Backends.dispatch(new DashboardActions.WantRepos());
		}
		if (nextState.remoteRepos === null && nextState.remoteRepos !== prevState.remoteRepos) {
			Dispatcher.Backends.dispatch(new DashboardActions.WantRemoteRepos());
		}
	}

	stores() { return [DashboardStore]; }

	render() {
		return (<div styleName="container">
			<Helmet title="Home" />
			{!context.currentUser &&
				<div styleName="anon-section">
					<img styleName="logo" src={`${context.assetsRoot || ""}/img/sourcegraph-logo.svg`}/>
					<div styleName="anon-title">Understand and use code better</div>
					<div styleName="anon-header-sub">
						Use Sourcegraph to search, browse, and cross-reference code.
						<br />
						Works with both public and private GitHub repositories written in Go.
					</div>
				</div>
			}
			<div styleName="repos">
				<DashboardRepos repos={(this.state.repos || []).concat(this.state.remoteRepos || [])}
					exampleRepos={this.state.exampleRepos}
					hasLinkedGitHub={this.state.hasLinkedGitHub} />
			</div>
		</div>);
	}
}

DashboardContainer.propTypes = {
};


export default CSSModules(DashboardContainer, styles);
