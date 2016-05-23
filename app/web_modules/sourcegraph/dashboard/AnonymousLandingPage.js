import React from "react";
import {Link} from "react-router";
import Component from "sourcegraph/Component";
import CSSModules from "react-css-modules";
import styles from "./styles/Home.css";
import base from "sourcegraph/components/styles/_base.css";
import GitHubAuthButton from "sourcegraph/user/GitHubAuthButton";
import LocationStateToggleLink from "sourcegraph/components/LocationStateToggleLink";


class AnonymousLandingPage extends Component {
	static contextTypes = {
		signedIn: React.PropTypes.bool.isRequired,
		siteConfig: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
	}

	render() {
		const {siteConfig, eventLogger} = this.context;
		return (
			<div styleName="home">
				<div styleName="container-with-globe">
					<div styleName="row">
						<div styleName="hero">
							<h1 styleName="h1">
								<strong styleName="text-purple">Fast,&nbsp;semantic&nbsp;code&nbsp;search &amp; cross&#x2011;reference&nbsp;engine</strong>
							</h1>
							<hr styleName="short-purple-line" />
							<div styleName="hero-body">
								<p style={{display: "none"}}>Search, browse code like an IDE, and see live usage examples.</p>
								<p>Search for a function/type/package &amp; see how other developers use it, across all public and (your) private code.</p>
								<LocationStateToggleLink href="/join"
									modalName="signup" location={location}
									onToggle={(v) => v && eventLogger.logEvent("ViewSignupModal", {source: "homepage-maincontent"})}
									className={base.mr4}>
									<GitHubAuthButton style={{display: "inline-block"}}><strong>Continue with GitHub</strong></GitHubAuthButton>
								</LocationStateToggleLink>
								<Link to="/github.com/aws/aws-sdk-go/-/info/GoPackage/github.com/aws/aws-sdk-go/aws/credentials/-/NewStaticCredentials" onClick={(v) => v && eventLogger.logEvent("ClickedExplorePublicRepo")}>Try it on a popular codebase</Link>
							</div>
						</div>
					</div>
				</div>
				<div styleName="box-demo">
					<div styleName="demo-container">
						<div styleName="demo-animation">
							<img src={`${siteConfig.assetsRoot}/img/Homepage/how-ref.gif`} styleName="how-img" />
						</div>
					</div>
				</div>
				<div styleName="container-lg">
					<div styleName="box-white">
						<div styleName="language-container">
							<h1 styleName="language-header">Language support</h1>
							<hr styleName="short-blue-line" />
							<p styleName="lead">Powered by <a href="https://srclib.org/" target="new">srclib</a>, a hackable code analysis library.</p>

							<div styleName="language">
								Go
								<span styleName="label-blue">75,311 projects</span>
							</div>
							<h5 styleName="header-5">Top Go Projects</h5>

							<div styleName="row">
								<div styleName="featured-project">
									<Link to="/github.com/golang/go">
										<img src={`${siteConfig.assetsRoot}/img/symbols/folder.svg`} className={`${base.mt1} ${base.mr2}`} width="16px" />
										<strong>golang/go</strong>
									</Link>
									<p styleName="project-desc">
										Used in 45,328 repositories
									</p>
								</div>
								<div styleName="featured-project">
									<Link to="/github.com/gorilla/mux">
										<img src={`${siteConfig.assetsRoot}/img/symbols/folder.svg`} className={`${base.mt1} ${base.mr2}`} width="16px" />
										<strong>gorilla/mux</strong>
									</Link>
									<p styleName="project-desc">
										Used in 1,843 repositories
									</p>
								</div>
								<div styleName="featured-project">
									<Link to="/github.com/aws/aws-sdk-go">
										<img src={`${siteConfig.assetsRoot}/img/symbols/folder.svg`} className={`${base.mt1} ${base.mr2}`} width="16px" />
										<strong>aws-sdk-go</strong>
									</Link>
									<p styleName="project-desc">
										Used in 171 repositories
									</p>
								</div>

							</div>

							<h5 styleName="header-5">Coming soon</h5>

							<div styleName="row">
								<div styleName="language-2">
									C#
								</div>
								<div styleName="language-3">
									Java
								</div>
								<div styleName="language-5">
									JavaScript
								</div>
								<div styleName="language-2">
									Python
								</div>
							</div>
						</div>
					</div>

				</div>

				<div styleName="box-bottom">
					<div styleName="bottom-container">
						<h2 styleName="bottom-header">We&nbsp;built&nbsp;Sourcegraph&nbsp;to keep&nbsp;you&nbsp;in&nbsp;flow while&nbsp;coding.</h2>
						<p styleName="bottom-text">Start saving time and sharpening your skills. Join tons of other developers who use Sourcegraph, around the world and in large, well-known companies.</p>
						<GitHubAuthButton color="purple" outline="true" style={{display: "inline-block", marginTop: "15px", fontSize: "1.6rem"}}><strong>Continue with GitHub</strong></GitHubAuthButton>
						<a target="_blank"
							styleName="bottom-link"
							href="https://chrome.google.com/webstore/detail/sourcegraph-chrome-extens/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en"
							onClick={(v) => v && eventLogger.logEvent("ClickedInstallChromeExt")}>
							Just install the Chrome extension
						</a>
					</div>
				</div>

			</div>
		);
	}
}

export default CSSModules(AnonymousLandingPage, styles);
