// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as styles from "sourcegraph/home/styles/Integrations.css";
import {Heading, Button} from "sourcegraph/components";
import {Component} from "sourcegraph/Component";
import {inBeta} from "sourcegraph/user";
import * as betautil from "sourcegraph/util/betautil.tsx";
import {context} from "sourcegraph/app/context";

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

let plugins = [
	new Tool("Sublime Text", "/img/Dashboard/sublime-text.svg", "https://sourcegraph.com/beta"),
	new Tool("IntelliJ", "/img/Dashboard/intellij.svg", "https://sourcegraph.com/beta"),
	new Tool("VS Code", "/img/Dashboard/vscode.svg", "https://sourcegraph.com/beta"),
	new Tool("Emacs", "/img/Dashboard/emacs.svg", "https://sourcegraph.com/beta"),
	new Tool("Vim", "/img/Dashboard/vim.svg", "https://sourcegraph.com/beta"),
];

const otherTools = [
	new Tool("Chrome", "/img/Dashboard/google-chrome.svg", "https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack"),
];

interface Props {
	location: any;
}

type State = any;

export class Integrations extends Component<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		siteConfig: React.PropTypes.object.isRequired,
		router: React.PropTypes.object.isRequired,
	};

	reconcileState(state: State, props: Props): void {
		Object.assign(state, props);
	}

	render(): JSX.Element | null {
		if (context.user && inBeta(context.user, betautil.DESKTOP)) {
			plugins[0]["url"] = "https://github.com/sourcegraph-beta/sourcegraph-sublime-beta#sourcegraph-for-sublime-text-";
			plugins[1]["url"] = "https://github.com/sourcegraph-beta/sourcegraph-intellij#sourcegraph-for-intellij-idea";
			plugins[2]["url"] = "https://github.com/sourcegraph-beta/sourcegraph-vscode#sourcegraph-for-visual-studio-code";
			plugins[3]["url"] = "https://github.com/sourcegraph-beta/sourcegraph-emacs#sourcegraph-for-emacs";
			plugins[4]["url"] = "https://github.com/sourcegraph-beta/sourcegraph-vim-beta#sourcegraph-for-vim-";
		}

		return (
			<div className={this.props.location.state && this.props.location.state.modal === "integrations" ? "" : styles.container}>
				<div className={styles.menu}>
					<Heading level={7} color="gray">Install an editor plugin</Heading>
					<p>Choose your editor to download the right plugin and get instructions on installation.</p>
					<div className={styles.tool_list}>
						{plugins.map((tool, i) => (
							<a key={i} href={tool.url} target="_blank" className={styles.tool}>
								<img className={styles.img} src={`${(this.context as any).siteConfig.assetsRoot}${tool.img}`}></img>
								<div className={styles.caption}>{tool.name}</div>
							</a>
						))}
					</div>
					<Heading level={7} color="gray">Other tools</Heading>
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
