// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";
import * as styles from "./styles/Integrations.css";
import * as base from "sourcegraph/components/styles/_base.css";
import {Heading, Button} from "sourcegraph/components/index";
import Component from "sourcegraph/Component";

class Tool {
	name: string;
	img: string;
	url: string;

	constructor(name, img, url) {
		this.name = name;
		this.img = img;
		this.url = url;
	}
}

const plugins = [
	new Tool("Sublime Text", "/img/Dashboard/sublime-text.svg", "https://github.com/sourcegraph-beta/sourcegraph-sublime-beta#sourcegraph-for-sublime-text-"),
	new Tool("IntelliJ", "/img/Dashboard/intellij.svg", "https://github.com/sourcegraph/sourcegraph-intellij#sourcegraph-for-intellij-idea"),
	new Tool("VS Code", "/img/Dashboard/vscode.svg", "https://github.com/sourcegraph-beta/sourcegraph-vscode#sourcegraph-for-visual-studio-code"),

	new Tool("Emacs", "/img/Dashboard/emacs.svg", "https://github.com/sourcegraph/sourcegraph-emacs#sourcegraph-for-emacs"),
	new Tool("Vim", "/img/Dashboard/vim.svg", "https://github.com/sourcegraph-beta/sourcegraph-vim-beta#sourcegraph-for-vim-"),
];

const otherTools = [
	new Tool("Chrome", "/img/Dashboard/google-chrome.svg", "https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack"),
];

class Integrations extends Component<any, any> {
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

	render(): JSX.Element | null {
		return (
			<div styleName={this.props.location.state && this.props.location.state.modal === "integrations" ? "" : "container"}>
				<div className={styles.menu}>
					<Heading level="7" color="cool_mid_gray">Install an editor plugin</Heading>
					<p>Choose your editor to download the right plugin and get instructions on installation.</p>
					<div className={styles.tool_list}>
						{plugins.map((tool, i) => (
							<a key={i} href={tool.url} target="_blank" className={styles.tool}>
								<img className={styles.img} src={`${(this.context as any).siteConfig.assetsRoot}${tool.img}`}></img>
								<div className={styles.caption}>{tool.name}</div>
							</a>
						))}
					</div>
					<Heading level="7" color="cool_mid_gray" className={base.pb3}>Other tools</Heading>
					<div className={styles.tool_list}>
						{otherTools.map((tool, i) => (
							<a key={i} href={tool.url} target="_blank" className={styles.tool}>
								<img className={styles.img} src={`${(this.context as any).siteConfig.assetsRoot}${tool.img}`}></img>
								<div className={styles.caption}>{tool.name}</div>
							</a>
						))}
					</div>
				</div>
				{this.props.location.query.onboarding &&
					<footer className={styles.footer}>
						<a className={styles.footer_link} href="/desktop/home">
							<Button color="green" className={styles.footer_btn}>Continue</Button>
						</a>
					</footer>
				}
			</div>
		);
	}
}

export default CSSModules(Integrations, styles, {allowMultiple: true});
