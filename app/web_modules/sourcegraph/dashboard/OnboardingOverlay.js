import React from "react";

import Component from "sourcegraph/Component";

import ProgressBar from "sourcegraph/dashboard/ProgressBar";
import OnboardingWidget from "sourcegraph/dashboard/OnboardingWidget";

import * as OnboardingActions from "sourcegraph/dashboard/OnboardingActions";
import Dispatcher from "sourcegraph/Dispatcher";

class OnboardingOverlay extends Component {
	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		const doGitHubLink = this.state.progress.currentStep === 0;

		const imageURL = doGitHubLink ?
			"https://assets-cdn.github.com/images/modules/logos_page/GitHub-Mark.png" : "http://placekitten.com/g/115/115";
		const welcome = (<div className="github-link-welcome">
			<div className="avatar-container">
				<div className="avatar-lg">
					<img className={`avatar-lg ${doGitHubLink ? "avatar-github" : ""}`} src={imageURL} />
					{doGitHubLink ? null : (
						<div className="github-link-success-icon">
							<span className="check-icon"><i className="fa fa-check"></i></span>
						</div>
					)}
				</div>
			</div>
			{doGitHubLink ?
				<p className="header-text welcome-header">Connect with your GitHub<br />account</p> :
				<p className="header-text welcome-header">Welcome Johnathan!</p>}
			{doGitHubLink ?
				<p className="normal-text">In order to get you started we need to connect with your GitHub account. No worries, we won't change anything within your files.</p> :
				<p className="normal-text">You successfully connected<br />your GitHub account.</p>}
			<div className="footer">
				<button className="btn btn-block btn-primary btn-lg"
					onClick={(e) => {
						if (doGitHubLink) {
							window.location.href = this.state.progress.githubLinkURL;
						} else {
							Dispatcher.dispatch(new OnboardingActions.AdvanceProgressStep());
						}
					}}>{doGitHubLink ? "Grant Permission" : "Next"}</button>
			</div>
		</div>);

		return (
			<div className="onboarding-overlay">
				<div className="panel panel-default">
					<div className={`panel-heading panel-heading-${this.state.progress.currentStep}`}>
						<ProgressBar numSteps={this.state.progress.numSteps} currentStep={this.state.progress.currentStep}/>
					</div>
					<div className="panel-body">
						{this.state.progress.currentStep <= 1 ? welcome :
							<div>
							<p className="header-text normal-header">Select Repositories</p>
							<p className="normal-text">Sourcegraph's Code Intelligence currently supports Go and Java (with more languages coming soon!)</p>
								<OnboardingWidget items={this.state.items}
									currentType={this.state.currentType}
									currentOrg={this.state.currentOrg}
									orgs={this.state.orgs}
									selections={this.state.selections}
									selectAll={this.state.selectAll}
									menuLabel="organizations" />
							</div>
						}
					</div>
				</div>
			</div>
		);
	}
}

OnboardingOverlay.propTypes = {
	progress: React.PropTypes.object.isRequired,
	items: React.PropTypes.arrayOf(React.PropTypes.shape({
		index: React.PropTypes.number,
		name: React.PropTypes.string,
	})),
	orgs: React.PropTypes.arrayOf(React.PropTypes.string),
	currentOrg: React.PropTypes.string,
	currentType: React.PropTypes.string,
	selections: React.PropTypes.object.isRequired,
	selectAll: React.PropTypes.bool.isRequired,
};

export default OnboardingOverlay;
