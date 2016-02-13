import React from "react";

import Container from "sourcegraph/Container";
import "./DashboardBackend"; // for side effects
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import OnboardingStore from "sourcegraph/dashboard/OnboardingStore";
import EntityTypes from "sourcegraph/dashboard/EntityTypes";

import OnboardingOverlay from "sourcegraph/dashboard/OnboardingOverlay";

class OnboardingContainer extends Container {
	constructor(props) {
		super(props);
	}

	componentDidMount() {
		super.componentDidMount();
	}

	componentWillUnmount() {
		super.componentWillUnmount();
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.repos = DashboardStore.repos;
		state.users = DashboardStore.users;
		state.mirrorRepos = DashboardStore.mirrorRepos;
		state.progress = OnboardingStore.progress;
		state.selectedRepos = OnboardingStore.selectedRepos;
		state.selectedUsers = OnboardingStore.selectedUsers;
		state.currentOrg = OnboardingStore.currentOrg;
		state.orgs = OnboardingStore.orgs;
		state.selectAll = OnboardingStore.selectAll;
		state.currentType = OnboardingStore.currentType;
	}

	stores() { return [DashboardStore, OnboardingStore]; }

	onStateTransition(prevState, nextState) {
	}

	render() {
		let selections, items;
		if (this.state.currentType === EntityTypes.REPO) {
			items = this.state.mirrorRepos;
			selections = this.state.selectedRepos;
		} else if (this.state.currentType === EntityTypes.USER) {
			items = this.state.users;
			selections = this.state.selectedUsers;
		}
		return (
			<div className="onboarding-container">
				{this.state.progress.currentStep < this.state.progress.numSteps ?
					<OnboardingOverlay
						progress={this.state.progress}
						items={items}
						currentType={this.state.currentType}
						currentOrg={this.state.currentOrg}
						orgs={this.state.orgs}
						selections={selections}
						selectAll={this.state.selectAll} /> : null}
			</div>
		);
	}
}

OnboardingContainer.propTypes = {
};

export default OnboardingContainer;
