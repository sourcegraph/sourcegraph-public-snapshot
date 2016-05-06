import React from "react";

import Component from "sourcegraph/Component";

import CSSModules from "react-css-modules";
import styles from "./styles/Home.css";

class HomeContainer extends Component {
	static contextTypes = {
		signedIn: React.PropTypes.bool.isRequired,
		siteConfig: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
	}

	render() {
		const {siteConfig} = this.context;
		return (
			<div styleName="home">
				<div styleName="container">
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
							<a href="#" styleName="btn-main">Sign up and connect</a>
							<a href="#">Explore public repos</a>
						</div>
					</div>

					<div styleName="content-block">
						<div styleName="img-right"><img src={`${siteConfig.assetsRoot}/img/Homepage/screenshot-github.png`} styleName="img" width="460" /></div>
						<div styleName="content-left">
							<div styleName="content">
								<h3 styleName="h3">On GitHub</h3>
								<p>Sourcegraph’s Chrome extension uses the global graph to show you how any library or function in GitHub is being used – across both public and your own private repos.</p>
							</div>
							<a href="https://chrome.google.com/webstore/detail/sourcegraph-chrome-extens/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en" styleName="btn-main">Install the Chrome extension</a>
						</div>
					</div>

					<div styleName="content-block">
						<div styleName="img-left"><img src={`${siteConfig.assetsRoot}/img/Homepage/screenshot-godocs.png`} styleName="img" width="460" /></div>
						<div styleName="content-right">
							<div styleName="content">
								<h3 styleName="h3">In Go documentation</h3>
								<p>Our GoDoc integration shows global graph usage and references for Go’s standard library. Browse docs and examples as seamlessly as looking up code in your editor.</p>
							</div>
							<a href="https://godoc.org/" styleName="btn-main">View on GoDoc.org</a>
						</div>
					</div>

					<div styleName="box-white">
						<div styleName="responsive-container">
							<h1 styleName="h1">Growing language support</h1>
							<hr styleName="short-blue-line" />
							<p styleName="lead">Language support for Sourcegraph is powered by <a href="https://srclib.org/">srclib</a>, an hackable code analysis library.</p>

							<div styleName="language">Go</div>
							<h5 styleName="header-5">Top Go Projects</h5>

							<div styleName="row">
								<div styleName="featured-project">
									<a href="https://sourcegraph.com/github.com/golang/go">golang/go</a>
									<p styleName="project-desc">
										Used by 21,453 developers <br />
										Used in 5,398 projects
									</p>
								</div>
								<div styleName="featured-project">
									<a href="https://sourcegraph.com/github.com/aws/aws-sdk-go">aws-sdk-go</a>
									<p styleName="project-desc">
										Used by 21,453 developers <br />
										Used in 5,398 projects
									</p>
								</div>
								<div styleName="featured-project">
									<a href="https://sourcegraph.com/github.com/kubernetes/kubernetes">kubernetes/kubernetes</a>
									<p styleName="project-desc">
										Used by 21,453 developers <br />
										Used in 5,398 projects
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
									Ruby
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
								<div>
									<div styleName="question-mark">?</div>
									<div styleName="question">How are other developers using this function?</div>
								</div>
							</div>
							<div styleName="content-right-5">
								<h2 styleName="header-white">Stop coding alone</h2>
								<p styleName="text-white">Sourcegraph helps you write better code by giving you seamless access to references from codebases and developers from around the world.</p>
								<a href="https://chrome.google.com/webstore/detail/sourcegraph-chrome-extens/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en" styleName="btn-white">Join the graph</a>
							</div>
						</div>
					</div>
				</div>

			</div>
		);
	}
}

export default CSSModules(HomeContainer, styles);
