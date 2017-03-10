import * as React from "react";
import { Router } from "sourcegraph/app/router";
import { SignupLoginAuth } from "sourcegraph/components";
import { LocationStateModal, dismissModal } from "sourcegraph/components/Modal";
import { layout } from "sourcegraph/components/utils";

export class Signup extends React.Component<{}, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };

	render(): JSX.Element {
		return <SignupModalContainer modalName="join">
			<SignupLoginAuth>
				To sign up, please authorize <br {...layout.hide.notSm } /> private code with GitHub:
			</SignupLoginAuth>
		</SignupModalContainer>;
	}
}

interface Props {
	modalName: string;
	sticky?: boolean;
}

export class SignupModalContainer extends React.Component<Props, {}> {

	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };

	close = () => {
		dismissModal(this.props.modalName, this.context.router)();
	}

	render(): JSX.Element {
		return <LocationStateModal
			title="Sign up"
			padded={false}
			onDismiss={this.close}
			modalName={this.props.modalName}
			sticky={this.props.sticky}>
			{this.props.children}
		</LocationStateModal>;
	}
}
