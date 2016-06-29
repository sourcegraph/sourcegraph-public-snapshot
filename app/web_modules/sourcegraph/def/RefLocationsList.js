// @flow

import React from "react";
import Helmet from "react-helmet";
import {Link} from "react-router";
import {urlToDefInfo} from "sourcegraph/def/routes";
import styles from "sourcegraph/def/styles/Def.css";
import CSSModules from "react-css-modules";
import LocationStateToggleLink from "sourcegraph/components/LocationStateToggleLink";
import Button from "sourcegraph/components/Button";
import {urlToGitHubOAuth, privateGitHubOAuthScopes} from "sourcegraph/util/urlTo";
import {trimRepo} from "sourcegraph/repo";
import {defTitle, defTitleOK} from "sourcegraph/def/Formatter";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

class RefLocationsList extends React.Component {
	static propTypes = {
		def: React.PropTypes.object.isRequired,
		refLocations: React.PropTypes.object,
		showMax: React.PropTypes.number,

		// Current repo and path info, so that they can be highlighted.
		repo: React.PropTypes.string.isRequired,
		rev: React.PropTypes.string,
		path: React.PropTypes.string,

		location: React.PropTypes.object.isRequired,
	};

	static contextTypes = {
		signedIn: React.PropTypes.bool.isRequired,
		githubToken: React.PropTypes.object,
		eventLogger: React.PropTypes.object.isRequired,
	};

	render() {
		let def = this.props.def;
		let refLocs = this.props.refLocations;

		if (!refLocs) return null;

		const noGitHubPrivateReposScope = !this.context.githubToken || !this.context.githubToken.scope || !this.context.githubToken.scope.includes("repo");

		return (
			<div>
				{defTitleOK(def) && <Helmet title={`${defTitle(def)} · ${trimRepo(this.props.repo)}`} />}
				{refLocs.RepoRefs && refLocs.RepoRefs.map((repoRef, i) => (
					this.props.showMax && i >= this.props.showMax ? null : <div key={i} styleName="all-refs">
						<header styleName={this.props.repo === repoRef.Repo ? "active-group-header" : ""}>
							<span styleName="refs-count">{repoRef.Count}</span> <span>{repoRef.Repo}</span>
						</header>
					</div>
				))}
				{refLocs && refLocs.RepoRefs && refLocs.RepoRefs.length > 0 && this.props.showMax && (!refLocs.TotalRepos || refLocs.TotalRepos > this.props.showMax) &&
				<Link to={urlToDefInfo(def, this.props.rev)}>
					<Button styleName="view-all-button" color="blue">View all references</Button>
				</Link>}
				{/* Show a CTA for signup, but only if there are other external refs (so we don't
					annoyingly show it for every single internal ref. */}
				{(refLocs.RepoRefs && refLocs.RepoRefs.length > 1 && (!this.context.signedIn || noGitHubPrivateReposScope)) &&
					<p styleName="private-repos-cta">
						{!this.context.signedIn &&
							<LocationStateToggleLink styleName="cta-link"
								location={this.props.location}
								onClick={() => this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "Conversion_SignInFromRefList", {page_name: location.pathname})}
								href="/login"
								modalName="login">
								<strong>Sign in</strong> for results from your code
							</LocationStateToggleLink>
						}
						{this.context.signedIn && noGitHubPrivateReposScope &&
							<a styleName="cta-link"
								href={urlToGitHubOAuth(privateGitHubOAuthScopes, this.props.location)}
								onClick={() => this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "Conversion_AuthPrivateCodeFromRefList", {page_name: location.pathname})}>
								<strong>Authorize</strong> to see results from your private code
							</a>
						}
					</p>
				}
			</div>
		);
	}
}

export default CSSModules(RefLocationsList, styles, {allowMultiple: true});
