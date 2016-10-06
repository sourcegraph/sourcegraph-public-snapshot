// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as base from "sourcegraph/components/styles/_base.css";
import * as classNames from "classnames";
import * as colors from "sourcegraph/components/styles/_colors.css";
import * as styles from "sourcegraph/dashboard/styles/Dashboard.css";
import * as typography from "sourcegraph/components/styles/_typography.css";
import Helmet from "react-helmet";
import {Heading, Panel, Loader} from "sourcegraph/components";
import {GitHubAuthButton} from "sourcegraph/components/GitHubAuthButton";
import {privateGitHubOAuthScopes} from "sourcegraph/util/urlTo";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";
import {context} from "sourcegraph/app/context";

interface Props {
	location?: any;
	privateCodeAuthed?: any;
	repos: any[];
	completeStep?: any;
}

type State = any;

export class GitHubPrivateAuthOnboarding extends React.Component<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);

		this.state = {
			showAll: false,
			isLoading: this.props.location.search.includes("CompletedGitHubOAuth2Flow"),
		};
	}

	_skipClicked() {
		EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLICK, "SkipGitHubPrivateAuth", {page_name: "GitHubPrivateCodeOnboarding"});
		this.props.completeStep();
	}

	render(): JSX.Element | null {
		if (this.props.privateCodeAuthed) {
			this.props.completeStep();
			return null;
		}

		return (
			<div>
				<Helmet title="Home" />
				<div className={styles.onboarding_container}>
					<Panel className={classNames(base.pb3, base.ph4, base.ba, base.br2, colors.b__cool_pale_gray)}>
						<Heading pt={4} align="center" level={3}>
							Browse your private code with Sourcegraph
						</Heading>
						<div className={styles.user_actions} style={{maxWidth: "380px"}}>
							<p className={classNames(typography.tc, base.mt3, base.mb2, typography.f6, colors.cool_gray_8)} >
								Enable Sourcegraph on any private GitHub repositories for a better coding experience
							</p>
							<div className={classNames(base.pv5)}>
								<img width={332} style={{marginBottom: "-95px"}} src={`${context.assetsRoot}/img/Dashboard/OnboardingRepos.png`}></img>
								{this.state.isLoading && <div><Loader/></div>}
								{!this.state.isLoading && <GitHubAuthButton pageName={"GitHubPrivateCodeOnboarding"} scopes={privateGitHubOAuthScopes} returnTo={this.props.location} className={styles.github_button}>Add private repositories</GitHubAuthButton>}
							</div>
							<p>
								<a onClick={this._skipClicked.bind(this)}>Skip</a>
							</p>
						</div>
					</Panel>
				</div>
			</div>
		);
	}
}
