// tslint:disable: typedef ordered-imports

import * as React from "react";
import Helmet from "react-helmet";
import {Link} from "react-router";
import {urlToDefInfo} from "sourcegraph/def/routes";
import * as styles from "sourcegraph/def/styles/Def.css";
import {LocationStateToggleLink} from "sourcegraph/components/LocationStateToggleLink";
import {Button} from "sourcegraph/components/Button";
import {urlToGitHubOAuth, privateGitHubOAuthScopes} from "sourcegraph/util/urlTo";
import {trimRepo} from "sourcegraph/repo/index";
import {defTitle, defTitleOK} from "sourcegraph/def/Formatter";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

interface Props {
	def: any;
	refLocations?: any;
	showMax?: number;

	// Current repo and path info, so that they can be highlighted.
	repo: string;
	rev?: string;
	path?: string;

	location: any;
}

export class RefLocationsList extends React.Component<Props, any> {
	static contextTypes = {
		signedIn: React.PropTypes.bool.isRequired,
		githubToken: React.PropTypes.object,
		eventLogger: React.PropTypes.object.isRequired,
	};

	render(): JSX.Element | null {
		let def = this.props.def;
		let refLocs = this.props.refLocations;

		if (!refLocs) {
			return null;
		}

		const context = this.context as any;
		const noGitHubPrivateReposScope = !context.githubToken || !context.githubToken.scope || !context.githubToken.scope.includes("repo");

		return (
			<div>
				{defTitleOK(def) && <Helmet title={`${defTitle(def)} · ${trimRepo(this.props.repo)}`} />}
				{refLocs.RepoRefs && refLocs.RepoRefs.map((repoRef, i) => (
					this.props.showMax && i >= this.props.showMax ? null : <div key={i} className={styles.all_refs}>
						<header className={this.props.repo === repoRef.Repo ? styles.b : ""}>
							<span className={styles.refs_count}>{repoRef.Count}</span> <span>{repoRef.Repo}</span>
						</header>
					</div>
				))}
				{refLocs && refLocs.RepoRefs && refLocs.RepoRefs.length > 0 && this.props.showMax && (!refLocs.TotalRepos || refLocs.TotalRepos > this.props.showMax) &&
				<Link to={urlToDefInfo(def, this.props.rev)}>
					<Button className={styles.view_all_button} color="blue">View all references</Button>
				</Link>}
				{/* Show a CTA for signup, but only if there are other external refs (so we don't
					annoyingly show it for every single internal ref. */}
				{(refLocs.RepoRefs && refLocs.RepoRefs.length > 1 && (!context.signedIn || noGitHubPrivateReposScope)) &&
					<p className={styles.private_repos_cta}>
						{!context.signedIn &&
							<LocationStateToggleLink className={styles.cta_link}
								location={this.props.location}
								onClick={() => context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "Conversion_SignInFromRefList", {page_name: location.pathname})}
								href="/login"
								modalName="login">
								<strong>Sign in</strong> for results from your code
							</LocationStateToggleLink>
						}
						{context.signedIn && noGitHubPrivateReposScope &&
							<a className={styles.cta_link}
								href={urlToGitHubOAuth(privateGitHubOAuthScopes, this.props.location)}
								onClick={() => context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "Conversion_AuthPrivateCodeFromRefList", {page_name: location.pathname})}>
								<strong>Authorize</strong> to see results from your private code
							</a>
						}
					</p>
				}
			</div>
		);
	}
}
