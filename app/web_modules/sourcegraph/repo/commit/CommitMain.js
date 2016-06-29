// @flow weak

import React from "react";
import Helmet from "react-helmet";
import Container from "sourcegraph/Container";
import CSSModules from "react-css-modules";
import styles from "./CommitMain.css";
import DeltaStore from "sourcegraph/delta/DeltaStore";
import * as DeltaActions from "sourcegraph/delta/DeltaActions";
import "sourcegraph/delta/DeltaBackend";
import BlobStore from "sourcegraph/blob/BlobStore";
import "sourcegraph/blob/BlobBackend";
import RepoStore from "sourcegraph/repo/RepoStore";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import "sourcegraph/repo/RepoBackend";
import Commit from "sourcegraph/vcs/Commit";
import Dispatcher from "sourcegraph/Dispatcher";
import FileDiffs from "sourcegraph/delta/FileDiffs";

class CommitMain extends Container {
	static propTypes = {
		repo: React.PropTypes.string,
		repoID: React.PropTypes.number,
		rev: React.PropTypes.string,
		commitID: React.PropTypes.string,
	};

	stores() { return [DeltaStore, RepoStore, BlobStore]; }

	reconcileState(state, props) {
		Object.assign(state, props);

		state.commit = RepoStore.commits.get(state.repo, state.rev);
		state.parentCommitID = state.commit && !state.commit.Error && state.commit.Parents ? state.commit.Parents[0] : null;

		state.files = state.repo && state.commitID && state.parentCommitID ? DeltaStore.files.get(state.repo, state.parentCommitID, state.repo, state.commitID) : null;

		state.annotations = BlobStore.annotations;
	}

	onStateTransition(prevState, nextState) {
		if (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev) {
			Dispatcher.Backends.dispatch(new RepoActions.WantCommit(nextState.repo, nextState.rev));
		}
		if (prevState.repoID !== nextState.repoID || prevState.repo !== nextState.repo || prevState.commitID !== nextState.commitID || prevState.parentCommitID !== nextState.parentCommitID) {
			if (nextState.repo && nextState.commitID && nextState.parentCommitID) {
				Dispatcher.Backends.dispatch(new DeltaActions.WantFiles(nextState.repo, nextState.parentCommitID, nextState.repo, nextState.commitID));
			}
		}
	}

	render() {
		if (!this.state.commitID) return null;
		return (
			<div styleName="container">
				<Helmet title={`@${this.state.commitID.slice(0, 6)}`} />
				{this.state.commit && !this.state.commit.Error && <Commit repo={this.state.repo} commit={this.state.commit} full={true} />}
				{this.state.files && !this.state.files.Error &&
				<FileDiffs files={this.state.files.FileDiffs}
					stats={this.state.files.Stats}
					baseRepo={this.state.repo}
					baseRev={this.state.parentCommitID}
					headRepo={this.state.repo}
					headRev={this.state.commitID} />}
			</div>
		);
	}
}

export default CSSModules(CommitMain, styles);
