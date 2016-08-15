// tslint:disable: typedef ordered-imports

import * as React from "react";

import {Container} from "sourcegraph/Container";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import "./TreeBackend"; // for side effects
import {TreeStore} from "sourcegraph/tree/TreeStore";

interface Props {
	repo: string;
	rev: string;
	path: string;
};

type State = any;

export class TreeEntryCommit extends Container<Props, State> {
	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
		state.commits = TreeStore.commits;
	}

	stores(): FluxUtils.Store<any>[] {
		return [TreeStore];
	}

	onStateTransition(prevState: State, nextState: State): void {
		if (prevState.repo !== nextState.repo || prevState.rev !== nextState.rev || prevState.path !== nextState.path) {
			Dispatcher.Backends.dispatch(new TreeActions.WantCommit(nextState.repo, nextState.rev, nextState.path));
		}
	}

	render(): JSX.Element | null {
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
