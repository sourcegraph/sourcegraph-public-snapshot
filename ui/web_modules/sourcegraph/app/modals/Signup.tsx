import { hover } from "glamor";
import * as React from "react";

import { Router } from "sourcegraph/app/router";
import { Button } from "sourcegraph/components";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { LocationStateModal } from "sourcegraph/components/Modal";
import * as styles from "sourcegraph/components/styles/modal.css";
import { GitHubLogo } from "sourcegraph/components/symbols";
import { Close } from "sourcegraph/components/symbols/Primaries";
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

const sx = {
	maxWidth: 500,
	marginLeft: "auto",
	marginRight: "auto",
	padding: 0,
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
				<div style={{ padding: `${whitespace[3]} 0` }}>
					<strong style={{ marginLeft: "1.5rem" }}>Sign up</strong>
					{!this.props.sticky && <a onClick={this.close}
						{...hover({ color: `${colors.blueGray()} !important` }) }
						style={{
							color: colors.blueGrayL1(),
							float: "right",
							padding: `0 ${whitespace[3]}`,
						}}>
						<Close width={24} />
					</a>}
				</div>
				<hr style={dividerSx} />
				{this.props.children}
			</div>
		</LocationStateModal>;
	}
}
