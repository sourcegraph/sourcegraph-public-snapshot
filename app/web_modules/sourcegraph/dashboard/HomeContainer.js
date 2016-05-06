import React from "react";

import Component from "sourcegraph/Component";

import CSSModules from "react-css-modules";
import styles from "./styles/Home.css";

class HomeContainer extends Component {
	static contextTypes = {
		signedIn: React.PropTypes.bool.isRequired,
	};

	constructor(props) {
		super(props);
	}

	render() {
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
						<div styleName="img-left"><img src="https://placekitten.com/g/460/300" styleName="img" /></div>
						<div styleName="content-right">
							<div styleName="content">
								<h3 styleName="h3">Search public and (your) private code</h3>
								<p>Connect your GitHub account to Sourcegraph and we’ll analyze your repositories – letting you look up and reference code from any other public repository on the graph.</p>
							</div>
							<a href="#" styleName="btn-main">Sign up and connect</a>
							<a href="#" styleName="link-secondary">Explore public repos</a>
						</div>
					</div>

					<div styleName="content-block">
						<div styleName="img-right"><img src="https://placekitten.com/g/460/300" styleName="img" /></div>
						<div styleName="content-left">
							<div styleName="content">
								<h3 styleName="h3">Search public and (your) private code</h3>
								<p>Connect your GitHub account to Sourcegraph and we’ll analyze your repositories – letting you look up and reference code from any other public repository on the graph.</p>
							</div>
							<a href="#" styleName="btn-main">Sign up and connect</a>
							<a href="#" styleName="link-secondary">Explore public repos</a>
						</div>
					</div>

					<div styleName="content-block">
						<div styleName="img-left"><img src="https://placekitten.com/g/460/300" styleName="img" /></div>
						<div styleName="content-right">
							<div styleName="content">
								<h3 styleName="h3">Search public and (your) private code</h3>
								<p>Connect your GitHub account to Sourcegraph and we’ll analyze your repositories – letting you look up and reference code from any other public repository on the graph.</p>
							</div>
							<a href="#" styleName="btn-main">Sign up and connect</a>
							<a href="#" styleName="link-secondary">Explore public repos</a>
						</div>
					</div>

					<div styleName="box-white">
						<h1 styleName="h1">Growing language support</h1>
						<hr styleName="short-blue-line" />
						<p styleName="lead">Language support for Sourcegraph is powered by srclib, an hackable code analysis library.</p>
					</div>

					<div styleName="box-purple-gradient">
						<h3>Stop coding alone</h3>
						<p>Sourcegraph helps you write better code by giving you seamless access to references from codebases and developers from around the world.</p>
					</div>

				</div>

			</div>
		);
	}
}

export default CSSModules(HomeContainer, styles);
