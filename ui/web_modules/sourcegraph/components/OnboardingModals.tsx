import * as React from "react";

import { Router, RouterLocation } from "sourcegraph/app/router";
import { LocationStateModal, setLocationModalState } from "sourcegraph/components/Modal";
import { ChromeExtensionOnboarding } from "sourcegraph/dashboard/ChromeExtensionOnboarding";
import { GitHubPrivateAuthOnboarding } from "sourcegraph/dashboard/GitHubPrivateAuthOnboarding";

interface Props {
	location: RouterLocation;
}

export class OnboardingModals extends React.Component<Props, {}>  {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };

	_completeChromeStep(): void {
		setLocationModalState(this.context.router, this.props.location, "github", true);
	}

	_completeGitHubStep(): void {
		setLocationModalState(this.context.router, this.props.location, "github", false);
	}

	render(): JSX.Element | null {
		return (
			<div>
				<LocationStateModal modalName="chrome">
					<div style={{ maxWidth: "800px", marginLeft: "auto", marginRight: "auto" }}>
						<ChromeExtensionOnboarding completeStep={this._completeChromeStep.bind(this)} location={this.props.location} />
					</div>
				</LocationStateModal>
				<LocationStateModal modalName="github">
					<div style={{ maxWidth: "800px", marginLeft: "auto", marginRight: "auto" }}>
						<GitHubPrivateAuthOnboarding completeStep={this._completeGitHubStep.bind(this)} location={this.props.location} />
					</div>
				</LocationStateModal>
			</div>
		);
	}
}
