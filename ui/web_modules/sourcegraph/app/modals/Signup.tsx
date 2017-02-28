import * as React from "react";

import { Router } from "sourcegraph/app/router";
import { Button } from "sourcegraph/components";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { LocationStateModal } from "sourcegraph/components/Modal";
import * as styles from "sourcegraph/components/styles/modal.css";
import { GitHubLogo } from "sourcegraph/components/symbols";
import { Close } from "sourcegraph/components/symbols/Primaries";
import { colors } from "sourcegraph/components/utils";
import { ghCodeAction } from "sourcegraph/user/Signup";

export class Signup extends React.Component<{}, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };

	render(): JSX.Element {
		const buttonSx = {
			padding: 0,
			textAlign: "left",
		};
		const privateCode = ghCodeAction(this.context.router, true);
		const publicCode = ghCodeAction(this.context.router, false);
		return <SignupModalContainer modalName="join">
			<div style={{ padding: 20 }}>
				To sign up, please authorize with GitHub:
			</div>
			{privateCode.form}
			<Button onClick={privateCode.submit} color="blue" style={buttonSx}>
				<GitHubLogo style={{ padding: 10, background: colors.blueD2() }} />
				<span style={{ paddingLeft: 10, paddingRight: 20 }}>Authorize with GitHub</span>
			</Button>
			<div style={{ color: colors.gray(.8) }}>
				or
			</div>
			{publicCode.form}
			<a onClick={publicCode.submit}>Only authorize public code</a>
			<hr style={{ marginLeft: -30, marginRight: -30, marginBottom: 20 }} />
			Already have an account? <LocationStateToggleLink href="/login" modalName="login" location={location}>
				Log in
			</LocationStateToggleLink>
		</SignupModalContainer>;
	}
}

const sx = {
	maxWidth: "420px",
	marginLeft: "auto",
	marginRight: "auto",
	padding: 0,
	textAlign: "center",
};

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
		return <LocationStateModal modalName={this.props.modalName} sticky={this.props.sticky}>
			<div className={styles.modal} style={sx}>
				<div style={{ padding: 15, fontWeight: 800, textAlign: "left" }}>
					Sign up
					{!this.props.sticky && <a style={{ float: "right" }} onClick={close}>
						<Close style={{ color: colors.black(.5) }} />
					</a>}
				</div>
				<hr style={{ margin: 0 }} />
				<div style={{ padding: 30, paddingTop: 0 }}>
					{this.props.children}
				</div>
			</div>
		</LocationStateModal>;
	}
}
