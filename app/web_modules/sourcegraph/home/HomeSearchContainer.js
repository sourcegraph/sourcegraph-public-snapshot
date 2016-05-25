import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/HomeSearch.css";
import base from "sourcegraph/components/styles/_base.css";
import GitHubAuthButton from "sourcegraph/user/GitHubAuthButton";
import {urlToGitHubOAuth, urlToPrivateGitHubOAuth} from "sourcegraph/util/urlTo";
import {Link} from "react-router";
import GlobalSearch from "sourcegraph/search/GlobalSearch";
import {EventLocation} from "sourcegraph/util/EventLogger";
import {Panel} from "sourcegraph/components";

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
			<div className={base.mt4} styleName="inline-block">
				{!this.context.githubToken &&
					<GitHubAuthButton href={urlToGitHubOAuth} onClick={() => this.context.eventLogger.logEventForPage("InitiateGitHubOAuth2Flow", EventLocation.Dashboard, {scopes: "", upgrade: true})}>Link GitHub account</GitHubAuthButton>
				}
				{this.context.githubToken && (!this.context.githubToken.scope || !(this.context.githubToken.scope.includes("repo") && this.context.githubToken.scope.includes("read:org") && this.context.githubToken.scope.includes("user:email"))) &&
					<GitHubAuthButton href={urlToPrivateGitHubOAuth} onClick={() => this.context.eventLogger.logEventForPage("InitiateGitHubOAuth2Flow", EventLocation.Dashboard, {scopes: "read:org,repo,user:email", upgrade: true})}>Use with private repositories</GitHubAuthButton>
				}
			</div>
		);
	}

	render() {
		return (
			<div styleName="bg">
				<div styleName="container-fixed" className={base.mt5}>
					<Panel className={`${base.mb4} ${base.pb4} ${base.ph4} ${base.pt3}`}>
						<GlobalSearch query={this.props.location.query.q || ""}/>
					</Panel>
					{!this.props.location.query.q && <div>
					{this.context.githubToken && <div styleName="tc" className={base.mv3}>
							<p styleName="cool-mid-gray">Try some common searches:</p>
							<Link to="github.com/golang/go@eb69476c66339ca494f98e65a78d315da99a9c79/-/info/GoPackage/net/http/-/Client/Get"><code>http.Get</code></Link>, <Link to="github.com/golang/go@eb69476c66339ca494f98e65a78d315da99a9c79/-/info/GoPackage/fmt/-/Sprintf"><code>Sprintf</code></Link>, <Link to="github.com/golang/go@b66b97e0a120880e37b03eba00c0c7679f0a70c1/-/info/GoPackage/image/-/Decode"><code>func Decode</code></Link>
						</div>}
						{!this.context.githubToken && <div styleName="tc">
							<div>Did you know you can search your repositories?</div>
							<div>Continue with GitHub to automatically index your repositories.</div>
							{this.renderCTAButtons()}
						</div>}
					</div>}
				</div>
			</div>
		);
	}
}

export default CSSModules(HomeSearchContainer, styles);
