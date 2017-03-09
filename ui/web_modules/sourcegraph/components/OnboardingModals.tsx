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
		setLocationModalState(this.context.router, "github", true);
	}

	_completeGitHubStep(): void {
		setLocationModalState(this.context.router, "github", false);
	}

	render(): JSX.Element | null {
		return (
			<div>
				<LocationStateModal modalName="chrome" style={{ maxWidth: 800 }} title="Get the Chrome extension">
					<ChromeExtensionOnboarding completeStep={this._completeChromeStep.bind(this)} location={this.props.location} />
				</LocationStateModal>
				<LocationStateModal modalName="github" style={{ maxWidth: 800 }} title="Authorize private code">
					<GitHubPrivateAuthOnboarding completeStep={this._completeGitHubStep.bind(this)} location={this.props.location} />
				</LocationStateModal>
			</div>
		);
	}
}
