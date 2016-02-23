import React from "react";

import Container from "sourcegraph/Container";

import OnboardingStore from "sourcegraph/dashboard/OnboardingStore";
import ImportGitHubReposMenu from "sourcegraph/dashboard/ImportGitHubReposMenu";
import ImportGitHubUsersMenu from "sourcegraph/dashboard/ImportGitHubUsersMenu";
import LinkGitHubWelcome from "sourcegraph/dashboard/LinkGitHubWelcome";
import ProgressBar from "sourcegraph/dashboard/ProgressBar";

import Dispatcher from "sourcegraph/Dispatcher";

import * as OnboardingActions from "sourcegraph/dashboard/OnboardingActions";

class OnboardingContainer extends Container {
	constructor(props) {
		super(props);
		this._dismissOnboardingModals = this._dismissOnboardingModals.bind(this);
	}

	componentDidMount() {
		super.componentDidMount();
		document.addEventListener("keydown", this._dismissOnboardingModals, false);
	}

	componentWillUnmount() {
		super.componentWillUnmount();
		document.removeEventListener("keydown", this._dismissOnboardingModals, false);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.progress = OnboardingStore.progress;
		state.currentUser = OnboardingStore.currentUser;
	}

	_dismissOnboardingModals(event) {
		// keyCode 27 is the escape key
		if (event.keyCode === 27) {
			Dispatcher.dispatch(new OnboardingActions.AdvanceProgressToStep(50));
		}
	}

	stores() { return [OnboardingStore]; }

	render() {
		if (this.state.progress.currentStep >= this.state.progress.numSteps) return null;

		return (
			<div className="onboarding-container">
				<div className="modal"
					tabIndex="-1">
					<div className="modal-dialog">
						<div className="modal-content">
							<div className={`modal-header modal-header-${this.state.progress.currentStep}`}>
								<ProgressBar numSteps={this.state.progress.numSteps} currentStep={this.state.progress.currentStep}/>
							</div>
							<div className="modal-body">
								{this.state.progress.currentStep <= 1 &&
									<LinkGitHubWelcome progress={this.state.progress} currentUser={this.state.currentUser}/>
								}
								{this.state.progress.currentStep === 2 &&
									<div>
										<p className="header-text normal-header">
											Select Repositories to Mirror
										</p>
										<p className="normal-text">
											Sourcegraph's Code Intelligence currently supports Go and Java (with more languages coming soon!)
										</p>
										<ImportGitHubReposMenu />
									</div>
								}
								{this.state.progress.currentStep === 3 &&
									<div>
										<p className="header-text normal-header">
											Invite People from GitHub
										</p>
										<p className="normal-text">
											Sourcegraph is more fun with people. You can invite your GitHub Connections, or do it the old fashioned way.
										</p>
										<ImportGitHubUsersMenu />
									</div>
								}
								{(this.state.progress.currentStep === 2 || this.state.progress.currentStep === 3) &&
									<p className="next-step">
										<a onClick={(e) => {
											e.preventDefault();
											Dispatcher.dispatch(new OnboardingActions.AdvanceProgressStep());
										}}>i'll do that later</a>
									</p>
								}
							</div>
						</div>
					</div>
				</div>
			</div>
		);
	}
}

OnboardingContainer.propTypes = {
};

export default OnboardingContainer;
