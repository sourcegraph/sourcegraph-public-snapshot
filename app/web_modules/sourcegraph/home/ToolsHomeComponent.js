import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/Tools.css";
import base from "sourcegraph/components/styles/_base.css";
import {Heading, Hero, Panel, Button} from "sourcegraph/components";
import Component from "sourcegraph/Component";
import {urlToGitHubOAuth} from "sourcegraph/util/urlTo";
import ToolComponent from "./ToolComponent";

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

	constructor(props) {
		super(props);
		this.state = {
			showChromeExtensionCTA: !window.localStorage["installed_chrome_extension"],
		};

		this.supportedTools = {
			browser: {
				hero: {
					title: "Sourcegraph for your browser",
					subtitle: "Jump-to-definition, code search, and documentation.",
					paragraph: "This extension enhances code pages on GitHub by making every identifier a jump-to-definition link. Hovering over identifiers displays a tooltip with documentation and type information. The new keyboard shortcut Shift-T allows you to search for functions, types, and other definitions. After installing it, view some code on GitHub and hover your mouse over identifiers.",
					img: "/img/Dashboard/GoogleChromeAsset.svg",
				},
				primaryButton: this._browserCTA(),
			},
			editor: {
				hero: {
					title: "Sourcegraph for your editor",
					subtitle: "Live usage examples for Go code, as you type.",
					paragraph: "When Sourcegraph is installed in your editor, you get real usage examples from across GitHub, immediate access to source code, and documentation about libraries, objects, and functions as you type. No more repetitive Google searches, and no more interruptions. We currently only support Go, but are quickly building support for additional languages and editors.",
					img: "/img/Dashboard/SourcegraphLiveAsset.svg",
				},
				primaryButton: this._sfyeCTA(),
				gif: "/img/Dashboard/LiveEditorExampleGif.gif",
			},
		};
	}

	componentDidMount() {
		this.timeout = setTimeout(() => this.setState({
			showChromeExtensionCTA: !document.getElementById("sourcegraph-app-bootstrap") && !window.localStorage["installed_chrome_extension"],
		}), 1);
	}

	reconcileState(state, props, context) {
		Object.assign(state, props);
		state.githubToken = context.githubToken;

		this.optionalSelectedToolComponent = this.supportedTools[state.location.query.tool] ? <ToolComponent supportedTool={this.supportedTools[state.location.query.tool]} location={state.location}/> : null;
	}

	_toolClicked(toolType) {
		this.context.eventLogger.logEventForPage("ToolCTAClicked", "DashboardTools", {toolType: toolType});
		this.context.router.replace({...this.props.location, query: {tool: toolType}});
	}

	_browserCTA() {
		if (global.chrome) {
			if (!this.state.showChromeExtensionCTA) {
				return (
					<Button disabled={true}>Installed</Button>
				);
			}

			return (
				<Button color="purple" disabled={false} onClick={this._installChromeExtensionClicked.bind(this)}>Install</Button>
			);
		}

		return (
			<Button color="purple" onClick={this._browserLearnMoreCTAClicked.bind(this)}>View on the chrome web store</Button>
		);
	}

	_browserLearnMoreCTAClicked() {
		this.context.eventLogger.logEventForPage("ChromeExtensionStoreCTAClicked", "DashboardTools");
		window.location.href("https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en");
	}

	_successHandler() {
		this.context.eventLogger.logEventForPage("ChromeExtensionInstalled", "DashboardTools");
		this.context.eventLogger.setUserProperty("installed_chrome_extension", "true");
		this.setState({showChromeExtensionCTA: false});
		window.localStorage["installed_chrome_extension"] = true;
	}

	_failHandler() {
		this.context.eventLogger.logEventForPage("ChromeExtensionInstallFailed", "DashboardTools");
		this.context.eventLogger.setUserProperty("installed_chrome_extension", "false");
		this.setState({showChromeExtensionCTA: true});
		window.localStorage.removeItem("installed_chrome_extension");
	}

	_installChromeExtensionClicked() {
		this.context.eventLogger.logEventForPage("ChromeExtensionCTAClicked", "DashboardTools");
		if (global.chrome) {
			global.chrome.webstore.install("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", this._successHandler.bind(this), this._failHandler.bind(this));
		}
	}

	_chromeCTAClicked() {
		if (global.chrome && this.state.showChromeExtensionCTA) {
			this._installChromeExtensionClicked();
		}

		this._toolClicked("browser");
	}

	_installEditorForSublimeCTAClicked() {
		this.context.eventLogger.logEventForPage("SourcegraphLiveCTAClicked", "DashboardTools", {editorType: "Sublime"});
		window.location.assign("https://packagecontrol.io/packages/Sourcegraph");
	}

	_installEditorForEmacsCTAClicked() {
		this.context.eventLogger.logEventForPage("SourcegraphLiveCTAClicked", "DashboardTools", {editorType: "Emacs"});
		window.location.assign("https://github.com/sourcegraph/sourcegraph/tree/master/client/emacs");
	}

	_sfyeCTA() {
		return (
			<div>
				<span className={base.ph1}>
					<Button color="purple" onClick={this._installEditorForSublimeCTAClicked.bind(this)}>Sublime</Button>
				</span>
				<span className={base.ph1}>
					<Button color="purple" onClick={this._installEditorForEmacsCTAClicked.bind(this)}>Emacs</Button>
				</span>
			</div>
		);
	}

	_connectGitHubClicked() {
		this.context.eventLogger.logEventForPage("InitiateGitHubOAuth2Flow", "DashboardTools", {scopes: "", upgrade: true});
		window.open(urlToGitHubOAuth);
	}

	render() {
		return (
			<div styleName="container">
			{this.optionalSelectedToolComponent}
				<Hero color="purple" pattern="objects">
					<div styleName="container-fixed">
						<Heading level="1" color="white" underline="white">Get Sourcegraph everywhere you code</Heading>
						<p style={{maxWidth: "560px"}} className={base.center}>
							Add Sourcegraph's instant coding assistance to your workflow.
						</p>
					</div>
				</Hero>
				{<div styleName="panel-container">
					{!this.context.signedIn && <div styleName="panel-item">
						<Panel hoverLevel="high">
							<img styleName="img" src={`${this.context.siteConfig.assetsRoot}/img/Dashboard/github-octocat.svg`}></img>
							<Heading align="center" level="4" className={base.ph4}>Sourcegraph for your repositories</Heading>
							<p styleName="cool-mid-gray" className={base.ph4}>
								Start searching, browsing, and cross-referencing your code.
							</p>
							<div styleName="button-container">
								<Button onClick={this._connectGitHubClicked.bind(this)} color="purple">Connect</Button>
							</div>
						</Panel>
					</div>}
					<div styleName="panel-item">
						<Panel hoverLevel="high">
							<img styleName="img" src={`${this.context.siteConfig.assetsRoot}/img/Dashboard/GoogleChromeAsset.svg`}></img>
							<Heading align="center" level="4" className={base.ph4}>Sourcegraph for your browser</Heading>
							<p styleName="cool-mid-gray" className={base.ph4}>
								Jump-to-definition, code search, and documentation.
							</p>
							<div styleName="button-container">
								<Button onClick={this._chromeCTAClicked.bind(this)} color="purple">{!this.state.showChromeExtensionCTA ? "Learn more" : "Install"}</Button>
							</div>
						</Panel>
					</div>
					<div styleName="panel-item">
						<Panel hoverLevel="high">
							<img styleName="img" src={`${this.context.siteConfig.assetsRoot}/img/Dashboard/SourcegraphLiveAsset.svg`}></img>
							<Heading align="center" level="4" className={base.ph4}>Sourcegraph for your editor</Heading>
							<p styleName="cool-mid-gray" className={base.ph4}>
								Live usage examples for Go code, as you type.
							</p>
							<div styleName="button-container">
								<Button onClick={this._toolClicked.bind(this, "editor")} color="purple">
									Learn more
								</Button>
							</div>
						</Panel>
					</div>
				</div>}
			</div>
		);
	}
}

export default CSSModules(ToolsHomeComponent, styles);
