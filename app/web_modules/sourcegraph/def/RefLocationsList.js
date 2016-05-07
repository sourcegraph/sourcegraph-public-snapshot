// @flow

import React from "react";
import {Link} from "react-router";
import {urlToDefInfo} from "sourcegraph/def/routes";
import styles from "sourcegraph/def/styles/Def.css";
import CSSModules from "react-css-modules";
import LocationStateToggleLink from "sourcegraph/components/LocationStateToggleLink";
import {urlToPrivateGitHubOAuth} from "sourcegraph/util/urlTo";

class RefLocationsList extends React.Component {
	static propTypes = {
		def: React.PropTypes.object.isRequired,
		refLocations: React.PropTypes.object,

		// Current repo and path info, so that they can be highlighted.
		repo: React.PropTypes.string.isRequired,
		rev: React.PropTypes.string.isRequired,
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
				{refLocs.RepoRefs && refLocs.RepoRefs.map((repoRef, i) => (
					<div key={i} styleName="all-refs">
						<header styleName={this.props.repo === repoRef.Repo ? "active-group-header" : ""}>
							<span styleName="refs-count">{repoRef.Count}</span> <Link to={urlToDefInfo(def, this.props.rev)}>{repoRef.Repo}</Link>
						</header>
					</div>
				))}
				{/* Show a CTA for signup, but only if there are other external refs (so we don't
					annoyingly show it for every single internal ref. */}
				{(refLocs.RepoRefs && refLocs.RepoRefs.length > 1 && (!this.context.signedIn || noGitHubPrivateReposScope)) &&
					<p styleName="private-repos-cta">
						{!this.context.signedIn &&
							<LocationStateToggleLink styleName="cta-link"
								location={this.props.location}
								onClick={() => this.context.eventLogger.logEvent("Conversion_SignInFromRefList")}
								href="/login"
								modalName="login">
								<strong>Sign in</strong> for results from your code
							</LocationStateToggleLink>
						}
						{this.context.signedIn && noGitHubPrivateReposScope &&
							<a styleName="cta-link"
								href={urlToPrivateGitHubOAuth}
								onClick={() => this.context.eventLogger.logEvent("Conversion_AuthPrivateCodeFromRefList")}>
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
