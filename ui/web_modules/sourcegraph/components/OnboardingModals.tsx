import {Location} from "history";
import * as React from "react";
import {InjectedRouter} from "react-router";
import {LocationStateModal, setLocationModalState} from "sourcegraph/components/Modal";
import {ChromeExtensionOnboarding} from "sourcegraph/dashboard/ChromeExtensionOnboarding";
import {GitHubPrivateAuthOnboarding} from "sourcegraph/dashboard/GitHubPrivateAuthOnboarding";

interface Props {
	location: Location;
}

export class OnboardingModals extends React.Component<Props, {}>  {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: InjectedRouter };

	_completeChromeStep(): void {
		setLocationModalState(this.context.router, this.props.location, "github", true);
	}

	_completeGitHubStep(): void {
		setLocationModalState(this.context.router, this.props.location, "github", false);
	}

	render(): JSX.Element | null {
		return (
			<div>
				<LocationStateModal modalName="chrome" location={this.props.location} router={this.context.router}>
					<div style={{maxWidth: "800px", marginLeft: "auto", marginRight: "auto"}}>
						<ChromeExtensionOnboarding completeStep={this._completeChromeStep.bind(this)} location={this.props.location}/>
					</div>
				</LocationStateModal>
				<LocationStateModal modalName="github" location={this.props.location} router={this.context.router}>
					<div style={{maxWidth: "800px", marginLeft: "auto", marginRight: "auto"}}>
						<GitHubPrivateAuthOnboarding completeStep={this._completeGitHubStep.bind(this)} repos={[]} location={this.props.location}/>
					</div>
				</LocationStateModal>
			</div>
		);
	}
}
