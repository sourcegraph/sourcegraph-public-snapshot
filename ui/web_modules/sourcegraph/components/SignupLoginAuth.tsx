import * as React from "react";

import { Router, RouterLocation } from "sourcegraph/app/router";
import { GitHubAuthButton } from "sourcegraph/components";
import { colors, typography, whitespace } from "sourcegraph/components/utils";
import { addQueryObjToURL, ghCodeAction } from "sourcegraph/user/Auth";

interface Props {
	children?: React.ReactNode[];
	newUserReturnTo?: string | RouterLocation;
	returnTo?: string | RouterLocation;
}

const dividerSx = {
	margin: 0,
	borderColor: colors.blueGrayL2(0.5),
};

const containerSx = {
	margin: "auto",
	maxWidth: 320,
	padding: `${whitespace[5]} ${whitespace[3]}`,
	textAlign: "center",
};

const subtextSx = {
	color: colors.blueGray(),
	padding: whitespace[4],
	textAlign: "center",
	...typography.small
};

export class SignupLoginAuth extends React.Component<Props, {}> {

	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };

	render(): JSX.Element {
		const publicCode = ghCodeAction(this.context.router, false);
		const { children, returnTo, newUserReturnTo } = this.props;
		const location = this.context.router.location;
		const signUpFlowURL = addQueryObjToURL(
			location,
			newUserReturnTo || location,
			{ tour: "signup", private: true },
		);

		return <div>
			<div style={containerSx}>
				{children &&
					<div style={{ marginBottom: whitespace[5] }}>{children}</div>}
				<GitHubAuthButton
					id="github-auth-btn"
					scope="private"
					block={true}
					returnTo={returnTo || location}
					newUserReturnTo={signUpFlowURL}>Authorize with GitHub</GitHubAuthButton>
				<p style={{ color: colors.blueGrayL1() }}>or</p>
				{publicCode.form}
				<a onClick={publicCode.submit}>Only authorize public code</a>
			</div>
			<hr style={dividerSx} />
			<div style={subtextSx}>
				By signing in, you agree to our <a href="/privacy" target="_blank">privacy policy</a> and <a href="/terms" target="_blank">terms</a>.
			</div>
		</div>;
	}
}
