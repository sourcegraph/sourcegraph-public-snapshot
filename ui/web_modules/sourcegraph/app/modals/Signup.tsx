import * as React from "react";

import { Router } from "sourcegraph/app/router";
import { Button } from "sourcegraph/components";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { LocationStateModal } from "sourcegraph/components/Modal";
import { GitHubLogo } from "sourcegraph/components/symbols";
import { colors, layout, whitespace } from "sourcegraph/components/utils";
import { ghCodeAction } from "sourcegraph/user/Signup";

const dividerSx = {
	margin: 0,
	borderColor: colors.blueGrayL2(0.5),
};

export class Signup extends React.Component<{}, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };

	render(): JSX.Element {
		const buttonSx = {
			marginTop: whitespace[5],
			padding: 0,
			textAlign: "left",
		};
		const privateCode = ghCodeAction(this.context.router, true);
		const publicCode = ghCodeAction(this.context.router, false);
		return <SignupModalContainer modalName="join">
			<div style={{
				margin: "auto",
				maxWidth: 320,
				padding: `${whitespace[5]} ${whitespace[3]}`,
				textAlign: "center",
			}}>
				To sign up, please authorize <br {...layout.hide.notSm } /> private code with GitHub:
				{privateCode.form}
				<Button onClick={privateCode.submit} color="blue" block={true} style={buttonSx}>
					<span style={{
						background: colors.blueD2(0.4),
						display: "inline-block",
						padding: whitespace[2],
					}}>
						<GitHubLogo width={24} />
					</span>
					<span style={{
						marginLeft: whitespace[3],
						marginRight: whitespace[3],
					}}>Authorize with GitHub</span>
				</Button>
				<p style={{ color: colors.blueGrayL1() }}>or</p>
				{publicCode.form}
				<a onClick={publicCode.submit}>Only authorize public code</a>
			</div>
			<hr style={dividerSx} />
			<div style={{ padding: "1.5rem", textAlign: "center" }}>
				Already have an account? <LocationStateToggleLink href="/login" modalName="login" location={location}>
					Log in
				</LocationStateToggleLink>
			</div>
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
		const url = { ...this.context.router.location, query: "", state: "" };
		this.context.router.push(url);
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
