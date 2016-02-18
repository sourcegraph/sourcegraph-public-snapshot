import React from "react";

import Component from "sourcegraph/Component";

import * as OnboardingActions from "sourcegraph/dashboard/OnboardingActions";
import Dispatcher from "sourcegraph/Dispatcher";

class LinkGitHubWelcome extends Component {
	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		const doGitHubLink = this.state.progress.currentStep === 0;
		const imageURL = doGitHubLink ?
			"https://assets-cdn.github.com/images/modules/logos_page/GitHub-Mark.png" :
			this.state.currentUser.AvatarURL;
		const username = this.state.currentUser.Name || this.state.currentUser.Login;

		return (
			<div className="github-link-welcome">
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
					<p className="header-text welcome-header">Welcome {username.split(" ")[0]}!</p>}
				{doGitHubLink ?
					<p className="normal-text">In order to get you started we need to connect with your GitHub account. No worries, we won't change anything within your files.</p> :
					<p className="normal-text">You successfully connected<br />your GitHub account.</p>}
				<footer>
					<button className="btn btn-block btn-primary btn-lg"
						onClick={(e) => {
							if (doGitHubLink) {
								window.location.href = this.state.progress.githubLinkURL;
							} else {
								Dispatcher.dispatch(new OnboardingActions.AdvanceProgressStep());
							}
						}}>{doGitHubLink ? "Grant Permission" : "Next"}</button>
				</footer>
			</div>
		);
	}
}

LinkGitHubWelcome.propTypes = {
	progress: React.PropTypes.object.isRequired,
	currentUser: React.PropTypes.object,
};

export default LinkGitHubWelcome;
