import React from "react";
import {Link} from "react-router";
import Component from "sourcegraph/Component";
import CSSModules from "react-css-modules";
import styles from "./styles/Home.css";
import LocationStateToggleLink from "sourcegraph/components/LocationStateToggleLink";


class HomeContainer extends Component {
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
						<div styleName="img-left"><img src={`${siteConfig.assetsRoot}/img/Homepage/screenshot-sourcegraph.png`} styleName="img" width="460" /></div>
						<div styleName="content-right">
							<div styleName="content">
								<h3 styleName="h3">Search public and (your) private code</h3>
								<p>Connect your GitHub account to Sourcegraph and we’ll analyze your repositories – letting you look up and reference code from any other public repository on the graph.</p>
							</div>
							<LocationStateToggleLink href="/join"
								modalName="signup" location={location}
								onToggle={(v) => v && eventLogger.logEvent("ViewSignupModal", {source: "homepage-maincontent"})}
								styleName="btn-main">
								Sign up and connect
							</LocationStateToggleLink>
							<Link to="/github.com/aws/aws-sdk-go/-/def/GoPackage/github.com/aws/aws-sdk-go/aws/credentials/-/NewStaticCredentials" onClick={(v) => v && eventLogger.logEvent("ClickedExplorePublicRepo")}>Explore public code example</Link>
						</div>
					</div>

					<div styleName="content-block">
						<div styleName="img-right"><img src={`${siteConfig.assetsRoot}/img/Homepage/screenshot-github.png`} styleName="img" width="460" /></div>
						<div styleName="content-left">
							<div styleName="content">
								<h3 styleName="h3">On GitHub</h3>
								<p>Sourcegraph’s Chrome extension uses the global graph to show you how any library or function in GitHub is being used – across both public and your own private repos.</p>
							</div>
							<a href="https://chrome.google.com/webstore/detail/sourcegraph-chrome-extens/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en" styleName="btn-main" target="new" onClick={(v) => v && eventLogger.logEvent("ClickedInstallChromeExt")}>Install the Chrome extension</a>
						</div>
					</div>

					<div styleName="box-white">
						<div styleName="responsive-container">
							<h1 styleName="h1">Growing language support</h1>
							<hr styleName="short-blue-line" />
							<p styleName="lead">Language support for Sourcegraph is powered by <a href="https://srclib.org/" target="new">srclib</a>, a hackable code analysis library.</p>

							<div styleName="language">
								Go
								<span styleName="label-blue">55,123 projects</span>
							</div>
							<h5 styleName="header-5">Top Go Projects</h5>

							<div styleName="row">
								<div styleName="featured-project">
									<Link to ="/github.com/golang/go">golang/go</Link>
									<p styleName="project-desc">
										An open source programming language that makes it easy to build simple, reliable, and efficient software.
									</p>
								</div>
								<div styleName="featured-project">
									<Link to ="/github.com/gorilla/mux">gorilla/mux</Link>
									<p styleName="project-desc">
										A powerful URL router and dispatcher.
									</p>
								</div>
								<div styleName="featured-project">
									<Link to ="/github.com/aws/aws-sdk-go">aws-sdk-go</Link>
									<p styleName="project-desc">
										The official AWS SDK for the Go programming language.
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
									onToggle={(v) => v && eventLogger.logEvent("ViewSignupModal", {source: "homepage-bottom"})}
									styleName="btn-white">
									Join the graph
								</LocationStateToggleLink>
							</div>
						</div>
					</div>
				</div>

			</div>
		);
	}
}

export default CSSModules(HomeContainer, styles);
