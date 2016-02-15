import React from "react";

import Container from "sourcegraph/Container";

import OnboardingStore from "sourcegraph/dashboard/OnboardingStore";
import GitHubReposStore from "sourcegraph/dashboard/GitHubReposStore";
import GitHubUsersStore from "sourcegraph/dashboard/GitHubUsersStore";
import SelectableListWidget from "sourcegraph/dashboard/SelectableListWidget";
import LinkGitHubWelcome from "sourcegraph/dashboard/LinkGitHubWelcome";
import ProgressBar from "sourcegraph/dashboard/ProgressBar";

import Dispatcher from "sourcegraph/Dispatcher";

import * as OnboardingActions from "sourcegraph/dashboard/OnboardingActions";

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
		console.log("reconciling state for onboarding container");
		Object.assign(state, props);
		state.progress = OnboardingStore.progress;
		switch (state.progress.currentStep) {
			case 2:
				state.items = GitHubReposStore.mirrorRepos;
				state.orgs = GitHubReposStore.orgs;
				state.selectAll = GitHubReposStore.selectAll;
				state.selections = GitHubReposStore.selectedRepos;
				state.currentOrg = GitHubReposStore.currentOrg;
				break;

			case 3:
				state.items = GitHubUsersStore.mirrorRepos;
				state.orgs = GitHubUsersStore.orgs;
				state.selectAll = GitHubUsersStore.selectAll;
				state.selections = GitHubUsersStore.selectedRepos;
				state.currentOrg = GitHubUsersStore.currentOrg;
				break;

			default:
				break;
		}
	}

	stores() { return [OnboardingStore, GitHubReposStore, GitHubUsersStore]; }

	render() {
		if (this.state.progress.currentStep >= this.state.progress.numSteps) return null;

		const panelBody = this.state.progress.currentStep <= 1 ?
			<LinkGitHubWelcome progress={this.state.progress} /> : (
			<div>
				<p className="header-text normal-header">
					Select Repositories
				</p>
				<p className="normal-text">
					Sourcegraph's Code Intelligence currently supports Go and Java (with more languages coming soon!)
				</p>
				<SelectableListWidget items={this.state.items}
					currentCategory={this.state.currentOrg}
					menuCategories={this.state.orgs}
					selections={this.state.selections}
					selectAll={this.state.selectAll}
					onSubmit={(items) => console.log("wooooooo it is submitted!", items)}
					menuLabel="organizations" />
					<p>
					<a onClick={(e) => {
						e.preventDefault();
						Dispatcher.dispatch(new OnboardingActions.AdvanceProgressStep());
					}}>i'll do that later</a>
					</p>
			</div>)

		return (
			<div className="onboarding-container">
				<div className="onboarding-overlay">
					<div className="panel panel-default">
						<div className={`panel-heading panel-heading-${this.state.progress.currentStep}`}>
							<ProgressBar numSteps={this.state.progress.numSteps} currentStep={this.state.progress.currentStep}/>
						</div>
						<div className="panel-body">{panelBody}</div>
					</div>
				</div>
			</div>
		);
	}
}

OnboardingContainer.propTypes = {
};

export default OnboardingContainer;
