// @flow weak

import React from "react";
import Helmet from "react-helmet";
import Container from "sourcegraph/Container";
import CSSModules from "react-css-modules";
import styles from "./CommitMain.css";
import * as BuildActions from "sourcegraph/build/BuildActions";
import BuildStore from "sourcegraph/build/BuildStore";
import "sourcegraph/build/BuildBackend";
import RepoStore from "sourcegraph/repo/RepoStore";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import "sourcegraph/repo/RepoBackend";
import Commit from "sourcegraph/vcs/Commit";
import Dispatcher from "sourcegraph/Dispatcher";

class CommitMain extends Container {
	static propTypes = {
		repo: React.PropTypes.string,
		rev: React.PropTypes.string,
		commitID: React.PropTypes.string,
	};

	stores() { return [BuildStore, RepoStore]; }

	reconcileState(state, props) {
		Object.assign(state, props);

		state.commit = RepoStore.commits.get(state.repo, state.rev);
	}

	onStateTransition(prevState, nextState) {
		if (prevState.commitID !== nextState.commitID || prevState.repo !== nextState.repo || prevState.rev !== nextState.rev) {
			Dispatcher.Backends.dispatch(new RepoActions.WantCommit(nextState.repo, nextState.rev));
		}
	}

	render() {
		if (!this.state.commitID) return null;
		return (
			<div styleName="container">
				<Helmet title={`@${this.state.commitID.slice(0, 6)}`} />
				{this.state.commit && !this.state.commit.Error && <Commit repo={this.state.repo} commit={this.state.commit} />}
			</div>
		);
	}
}

export default CSSModules(CommitMain, styles);
