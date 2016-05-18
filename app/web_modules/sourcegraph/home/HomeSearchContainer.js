import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/HomeSearch.css";
import GitHubAuthButton from "sourcegraph/user/GitHubAuthButton";
import {urlToGitHubOAuth, urlToPrivateGitHubOAuth} from "sourcegraph/util/urlTo";
import {Link} from "react-router";
import GlobalSearch from "sourcegraph/search/GlobalSearch";
import {EventLocation} from "sourcegraph/util/EventLogger";

class HomeSearchContainer extends React.Component {

	static propTypes = {
		location: React.PropTypes.object.isRequired,
	}

	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
		user: React.PropTypes.object,
		signedIn: React.PropTypes.bool.isRequired,
		githubToken: React.PropTypes.object,
		eventLogger: React.PropTypes.object.isRequired,
		router: React.PropTypes.object.isRequired,
	};

	renderCTAButtons() {
		return (
			<div>
				{!this.context.githubToken && <div styleName="cta">
					<GitHubAuthButton href={urlToGitHubOAuth} onClick={() => this.context.eventLogger.logEventForPage("InitiateGitHubOAuth2Flow", EventLocation.Dashboard, {scopes: "", upgrade: true})}>Link GitHub account</GitHubAuthButton>
				</div>}
				{this.context.githubToken && (!this.context.githubToken.scope || !(this.context.githubToken.scope.includes("repo") && this.context.githubToken.scope.includes("read:org") && this.context.githubToken.scope.includes("user:email"))) && <div styleName="cta">
					<GitHubAuthButton href={urlToPrivateGitHubOAuth} onClick={() => this.context.eventLogger.logEventForPage("InitiateGitHubOAuth2Flow", EventLocation.Dashboard, {scopes: "read:org,repo,user:email", upgrade: true})}>Use with private repositories</GitHubAuthButton>
				</div>}
			</div>
		);
	}

	render() {
		return (<div styleName="anon-section">
					<div styleName="anon-title">Global code search
						<div styleName="anon-title-left">
							<GlobalSearch query={this.props.location.query.q || ""}/>
						</div>
					</div>
					{!this.props.location.query.q && <div styleName="container">
					{this.context.githubToken && <div styleName="suggestion-subheader">Top queries
							<div styleName="suggestion"><Link to="github.com/golang/go@eb69476c66339ca494f98e65a78d315da99a9c79/-/info/GoPackage/net/http/-/Client/Get"><code>http.Get</code></Link></div>
							<div styleName="suggestion"><Link to="github.com/golang/go@eb69476c66339ca494f98e65a78d315da99a9c79/-/info/GoPackage/fmt/-/Sprintf"><code>Sprintf</code></Link></div>
							<div styleName="suggestion"><Link to="github.com/golang/go@b66b97e0a120880e37b03eba00c0c7679f0a70c1/-/info/GoPackage/image/-/Decode"><code>func Decode</code></Link></div>
						</div>}
						{!this.context.githubToken && <div>
							<div styleName="suggestion-header">Did you know you can search your repositories?
								<div styleName="suggestion-subheader">
									Continue with GitHub to automatically index your repositories.
								</div>
							</div>
							{this.renderCTAButtons()}
						</div>}
					</div>}
			</div>);
	}
}

export default CSSModules(HomeSearchContainer, styles);
