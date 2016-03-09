import React from "react";
import Container from "sourcegraph/Container";
import "sourcegraph/repo/RepoBackend";
import RepoStore from "sourcegraph/repo/RepoStore";
import RevSwitcher from "sourcegraph/repo/RevSwitcher";

// RevSwitcherContainer is for standalone RevSwitchers that need to
// be able to respond to changes in RepoStore independently.
class RevSwitcherContainer extends Container {
	reconcileState(state, props) {
		Object.assign(state, props);
		state.branches = RepoStore.branches;
		state.tags = RepoStore.tags;
	}

	stores() { return [RepoStore]; }

	render() {
		let childProps = this.props;
		delete childProps.repoStore;
		return <RevSwitcher branches={this.state.branches} tags={this.state.tags} {...childProps} />;
	}
}

RevSwitcherContainer.propTypes = {
	// All of the same properties as RevSwitcher, minus branches and tags.
};

export default RevSwitcherContainer;
