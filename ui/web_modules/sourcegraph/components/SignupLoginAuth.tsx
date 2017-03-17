import * as React from "react";

import { GitHubAuthButton } from "sourcegraph/components";
import { colors, typography, whitespace } from "sourcegraph/components/utils";
import { ComponentWithRouter } from "sourcegraph/core/ComponentWithRouter";
import { githubAuthAction } from "sourcegraph/user/Auth";

interface Props {
	children?: React.ReactNode[];
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

export class SignupLoginAuth extends ComponentWithRouter<Props, {}> {
	render(): JSX.Element {
		const publicCode = githubAuthAction(this.context.router, false);

		return <div>
			<div style={containerSx}>
				{this.props.children &&
					<div style={{ marginBottom: whitespace[5] }}>{this.props.children}</div>}
				<GitHubAuthButton
					id="github-auth-btn"
					privateCode={true}
					block={true}>Authorize with GitHub</GitHubAuthButton>
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
