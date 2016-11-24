import * as React from "react";
import { InjectedRouter } from "react-router";
import { context } from "sourcegraph/app/context";
import { Component } from "sourcegraph/Component";
import { Heading } from "sourcegraph/components";
import { GitHubAuthButton } from "sourcegraph/components/GitHubAuthButton";
import { PageTitle } from "sourcegraph/components/PageTitle";
import * as styles from "sourcegraph/user/styles/accountForm.css";
import "sourcegraph/user/UserBackend"; // for side effects

interface Props {
	location: any;
}

type State = any;

export class AuthExtForm extends Component<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};
	context: {
		router: InjectedRouter;
	};

	render(): JSX.Element | null {
		if (context.user && context.hasPrivateGitHubToken()) {
			const decodeUrl = decodeURIComponent(this.props.location.query["rtg"]);
			const returnUrl = new URL(decodeUrl);

			if (returnUrl.origin.match(/https:\/\/(www\.)?github.com/)) {
				setTimeout(() => { window.location.href = decodeUrl; }, 500);
			} else {
				throw new Error("Invalid return URL");
			}

			return null;
		}

		return (
			<div className={styles.form}>
				<Heading level={3} align="center" underline="orange">Welcome to Sourcegraph</Heading>
				<GitHubAuthButton returnTo={this.props.location} tabIndex={1} key="1" block={true}>Continue with GitHub</GitHubAuthButton>
				<p className={styles.mid_text}>
					By continuing with GitHub, you agree to our <a href="/privacy" target="_blank">privacy policy</a> and <a href="/terms" target="_blank">terms</a>.
				</p>
			</div>
		);
	}
}

function AuthBrowserExtFlowComp(props: { location: any }): JSX.Element {
	return (
		<div className={styles.full_page}>
			<PageTitle title="Authorize Sourcegraph for Github" />
			<AuthExtForm {...props} />
		</div>
	);
}

export const AuthBrowserExtFlow = AuthBrowserExtFlowComp;
