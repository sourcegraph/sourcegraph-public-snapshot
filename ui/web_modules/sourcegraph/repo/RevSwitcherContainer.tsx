import * as React from "react";
import {RouteParams} from "sourcegraph/app/routeParams";
import {Container} from "sourcegraph/Container";
import "sourcegraph/repo/RepoBackend";
import {RepoStore} from "sourcegraph/repo/RepoStore";
import {RevSwitcher} from "sourcegraph/repo/RevSwitcher";
import {Store} from "sourcegraph/Store";

interface Props {
	repo: string;
	rev: string;
	commitID: string;
	repoObj?: any;
	isCloning: boolean;

	// to construct URLs
	routes: any[];
	routeParams: RouteParams;
}

type State = any;

// RevSwitcherContainer is for standalone RevSwitchers that need to
// be able to respond to changes in RepoStore independently.
export class RevSwitcherContainer extends Container<Props, State> {
	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
		state.branches = RepoStore.branches;
		state.tags = RepoStore.tags;
	}

	stores(): Store<any>[] {
		return [RepoStore];
	}

	render(): JSX.Element | null {
		return (
			<RevSwitcher
				branches={this.state.branches}
				tags={this.state.tags}
				{...this.props} />
			);
	}
}
