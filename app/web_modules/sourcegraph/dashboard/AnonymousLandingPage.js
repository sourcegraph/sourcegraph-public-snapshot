import React from "react";
import {Link} from "react-router";
import Component from "sourcegraph/Component";
import Button from "sourcegraph/components/Button";
import Logo from "sourcegraph/components/Logo";
import CSSModules from "react-css-modules";
import styles from "./styles/Home.css";
import base from "sourcegraph/components/styles/_base.css";
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
					<h1 styleName="h1">
						<strong styleName="text-purple">Write better code</strong> by accessing examples and references from developers around the world
					</h1>
					<hr styleName="short-purple-line" />
					<h2 styleName="h2">Explore how you can use Sourcegraph to reference code from all over the world</h2>
				</div>
				<div styleName="container-lg">

					<div styleName="content-block">
						<div styleName="img-left">
						<Link to="/github.com/golang/go@0cc710dca63b79ed2dd6ce9375502e76e5fc0484/-/tree/src/testing?q=testing" onClick={(v) => v && eventLogger.logEvent("ClickedExplorePublicRepo")}>
							<img src={`${siteConfig.assetsRoot}/img/Homepage/screenshot-sourcegraph.png`} styleName="img" width="460" />
						</Link>
						</div>
						<div styleName="content-right">
							<div styleName="content">
								<Logo width="32px" className={base.mt4} />
								<h3 styleName="h3">Search public and (your) private code</h3>
								<p>Connect your GitHub account to Sourcegraph and we’ll analyze your repositories – letting you look up and reference code from any other public repository on the graph.</p>
							</div>
							<LocationStateToggleLink href="/join"
								modalName="signup" location={location}
								onToggle={(v) => v && eventLogger.logEvent("ViewSignupModal", {source: "homepage-maincontent"})}
								className={base.mr4}>
								<Button color="blue">Sign up and connect</Button>
							</LocationStateToggleLink>
							<Link to="/github.com/aws/aws-sdk-go/-/def/GoPackage/github.com/aws/aws-sdk-go/aws/credentials/-/NewStaticCredentials" onClick={(v) => v && eventLogger.logEvent("ClickedExplorePublicRepo")}>Explore public code example</Link>
						</div>
					</div>

					<div styleName="content-block">
						<div styleName="img-right">
							<a href="https://chrome.google.com/webstore/detail/sourcegraph-chrome-extens/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en" target="new">
								<img src={`${siteConfig.assetsRoot}/img/Homepage/screenshot-github.png`} styleName="img" width="460" />
							</a>
						</div>
						<div styleName="content-left">
							<div styleName="content">
								<img src={`${siteConfig.assetsRoot}/img/symbols/branch.svg`} className={base.mt3} />
								<h3 styleName="h3">On GitHub</h3>
								<p>Sourcegraph’s Chrome extension uses the global graph to show you how any library or function in GitHub is being used – across both public and your own private repos.</p>
							</div>
							<a href="https://chrome.google.com/webstore/detail/sourcegraph-chrome-extens/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en" target="new">
								<Button color="blue" onClick={(v) => v && eventLogger.logEvent("ClickedInstallChromeExt")}>
									Install the Chrome extension
								</Button>
							</a>
						</div>
					</div>

					<div styleName="box-white">
						<div styleName="responsive-container">
							<h1 styleName="h1">Growing language support</h1>
							<hr styleName="short-blue-line" />
							<p styleName="lead">Language support for Sourcegraph is powered by <a href="https://srclib.org/" target="new">srclib</a>, a hackable code analysis library.</p>

							<div styleName="language">
								Go
								<span styleName="label-blue">45,427 projects</span>
							</div>
							<h5 styleName="header-5">Top Go Projects</h5>

							<div styleName="row">
								<div styleName="featured-project">
									<Link to ="/github.com/golang/go">
										<img src={`${siteConfig.assetsRoot}/img/symbols/folder.svg`} className={`${base.mt1} ${base.mr2}`} width="16px" />
										<strong>golang/go</strong>
									</Link>
									<p styleName="project-desc">
										Used in 45,328 repositories
									</p>
								</div>
								<div styleName="featured-project">
									<Link to ="/github.com/gorilla/mux">
										<img src={`${siteConfig.assetsRoot}/img/symbols/folder.svg`} className={`${base.mt1} ${base.mr2}`} width="16px" />
										<strong>gorilla/mux</strong>
									</Link>
									<p styleName="project-desc">
										Used in 1,843 repositories
									</p>
								</div>
								<div styleName="featured-project">
									<Link to ="/github.com/aws/aws-sdk-go">
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

				<div styleName="box-purple-gradient">
					<div styleName="container">
						<div styleName="row">
							<div styleName="img-left-4">
								<img src={`${siteConfig.assetsRoot}/img/Homepage/how-ref.gif`} styleName="how-img" />
								<div styleName="row">
									<div styleName="question-mark">?</div>
									<div styleName="question">How are other developers using this function?</div>
								</div>
							</div>
							<div styleName="content-right-5">
								<h2 styleName="header-white">Stop coding alone</h2>
								<p styleName="text-white">Sourcegraph helps you write better code by giving you seamless access to references from codebases and developers all over the web.</p>
								<LocationStateToggleLink href="/join"
									modalName="signup" location={location}
									onToggle={(v) => v && eventLogger.logEvent("ViewSignupModal", {source: "homepage-bottom"})}>
									<Button color="white" className={base.mt3}>Join the graph</Button>
								</LocationStateToggleLink>
							</div>
						</div>
					</div>
				</div>

			</div>
		);
	}
}

export default CSSModules(AnonymousLandingPage, styles);
