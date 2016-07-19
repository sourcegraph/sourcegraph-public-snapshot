import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/Tools.css";
import base from "sourcegraph/components/styles/_base.css";
import {Heading, Button} from "sourcegraph/components";
import Component from "sourcegraph/Component";

class Tool {
	constructor(name, img, url) {
		this.name = name;
		this.img = img;
		this.url = url;
	}
}

const plugins = [
	new Tool("Sublime Text", "/img/Dashboard/SourcegraphSublime.svg", "https://github.com/sourcegraph/sourcegraph-sublime"),
];

const otherTools = [
	new Tool("Chrome", "/img/Dashboard/GoogleChromeAsset.svg", "https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack"),
];

class ToolsHomeComponent extends Component {
	static propTypes = {
		location: React.PropTypes.object.isRequired,
	};

	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
		router: React.PropTypes.object.isRequired,
		signedIn: React.PropTypes.bool.isRequired,
		githubToken: React.PropTypes.object,
	};

	reconcileState(state, props, context) {
		Object.assign(state, props);
	}

	render() {
		return (
			<div styleName="container">
				<div styleName="menu">
					<Heading level="7" color="cool-mid-gray">Install an editor plugin</Heading>
					<p>Choose your editor to download the right plugin and get instructions on installation.</p>
					<div styleName="tool-list">
						{plugins.map((tool, i) => (
							<a key={i} href={tool.url} styleName="tool">
								<img styleName="img" src={`${this.context.siteConfig.assetsRoot}${tool.img}`}></img>
								<div styleName="caption">{tool.name}</div>
							</a>
						))}
					</div>
					<Heading level="7" color="cool-mid-gray" className={base.pb3}>Other tools</Heading>
					<div styleName="tool-list">
						{otherTools.map((tool, i) => (
							<a key={i} href={tool.url} styleName="tool">
								<img styleName="img" src={`${this.context.siteConfig.assetsRoot}${tool.img}`}></img>
								<div styleName="caption">{tool.name}</div>
							</a>
						))}
					</div>
				</div>
				<footer styleName="footer">
					<Button color="green" styleName="footer-btn">Continue</Button>
				</footer>
			</div>
		);
	}
}

export default CSSModules(ToolsHomeComponent, styles, {allowMultiple: true});
