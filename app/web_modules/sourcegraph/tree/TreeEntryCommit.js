import React from "react";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import "./TreeBackend"; // for side effects
import TreeStore from "sourcegraph/tree/TreeStore";

class TreeEntryCommit extends Container {
	reconcileState(state, props) {
		Object.assign(state, props);
		state.commits = TreeStore.commits;
	}

	stores() { return [TreeStore]; }

	onStateTransition(prevState, nextState) {
		if (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.path !== nextState.path) {
			Dispatcher.Backends.dispatch(new TreeActions.WantCommit(nextState.repo, nextState.rev, nextState.path));
		}
	}

	render() {
		let commit = this.state.commits.get(this.state.repo, this.state.rev, this.state.path);
		if (commit === null) {
			return null;
		}

		commit = commit.Commits[0];

		let sig = commit.Author || commit.Committer;

		return (
				<div className="commit">
					<time>{sig.Date}</time>
					<div className="message">
						{commit.Message}
					</div>
				</div>
		);
	}
}

TreeEntryCommit.propTypes = {
	repo: React.PropTypes.string.isRequired,
	rev: React.PropTypes.string.isRequired,
	path: React.PropTypes.string.isRequired,
};

export default TreeEntryCommit;
