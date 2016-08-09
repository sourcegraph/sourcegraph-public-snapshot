// tslint:disable

import * as React from "react";
import Container from "sourcegraph/Container";
import "sourcegraph/repo/RepoBackend";
import RepoStore from "sourcegraph/repo/RepoStore";
import "sourcegraph/tree/TreeBackend";
import TreeStore from "sourcegraph/tree/TreeStore";
import RevSwitcher from "sourcegraph/repo/RevSwitcher";

type Props = {
	repo: string,
	rev?: string,
	commitID: string,
	repoObj?: any,
	isCloning: boolean,

	// srclibDataVersions is TreeStore.srclibDataVersions.
	srclibDataVersions?: any,

	// to construct URLs
	routes: any[],
	routeParams: any,
};

// RevSwitcherContainer is for standalone RevSwitchers that need to
// be able to respond to changes in RepoStore independently.
class RevSwitcherContainer extends Container<Props, any> {
	reconcileState(state, props) {
		Object.assign(state, props);
		state.branches = RepoStore.branches;
		state.tags = RepoStore.tags;
		state.srclibDataVersions = TreeStore.srclibDataVersions;
	}

	stores() { return [RepoStore, TreeStore]; }

	render(): JSX.Element | null {
		return (
			<RevSwitcher
				branches={this.state.branches}
				tags={this.state.tags}
				srclibDataVersions={this.state.srclibDataVersions}
				{...this.props} />
			);
	}
}

export default RevSwitcherContainer;
