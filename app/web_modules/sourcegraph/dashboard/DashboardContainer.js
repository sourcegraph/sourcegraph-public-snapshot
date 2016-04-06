import React from "react";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import "./DashboardBackend"; // for side effects
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import context from "sourcegraph/context";

import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";

class DashboardContainer extends Container {

	constructor(props) {
		super(props);
	}

	componentWillUnmount() {
		super.componentWillUnmount();
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.exampleRepos = DashboardStore.exampleRepos;
		state.repos = DashboardStore.repos;
		state.remoteRepos = DashboardStore.remoteRepos;
		state.hasLinkedGitHub = DashboardStore.hasLinkedGitHub;
		state.githubRedirect = props.location && props.location.query ? (props.location.query["github-onboarding"] || false) : false;
	}

	onStateTransition(prevState, nextState) {
		if (!nextState.repos && nextState.repos !== prevState.repos) {
			Dispatcher.Backends.dispatch(new DashboardActions.WantRepos());
		}
		if (!nextState.remoteRepos && nextState.remoteRepos !== prevState.remoteRepos) {
			Dispatcher.Backends.dispatch(new DashboardActions.WantRemoteRepos());
		}
	}

	stores() { return [DashboardStore]; }

	render() {
		return (<div styleName="container">
			{!context.currentUser &&
				<div styleName="anon-section">
					<img styleName="logo" src={`${context.assetsRoot}/img/sourcegraph-logo.svg`}/>
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
