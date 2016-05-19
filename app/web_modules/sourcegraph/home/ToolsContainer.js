import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/Tools.css";
import {EventLocation} from "sourcegraph/util/EventLogger";
import {urlToGitHubOAuth, urlToPrivateGitHubOAuth} from "sourcegraph/util/urlTo";
import {Link} from "react-router";

class ToolsContainer extends React.Component {

	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
		githubToken: React.PropTypes.object,
		eventLogger: React.PropTypes.object.isRequired,
		router: React.PropTypes.object.isRequired,
	}

	constructor(props) {
		super(props);
		this.state = {
			showChromeExtensionCTA: !window.localStorage["installed_chrome_extension"],
		};
	}

	componentDidMount() {
		setTimeout(() => this.setState({
			showChromeExtensionCTA: !document.getElementById("sourcegraph-app-bootstrap") && !window.localStorage["installed_chrome_extension"],
		}), 1);
	}
	_successHandler() {
		this.context.eventLogger.logEvent("ChromeExtensionInstalled");
		this.context.eventLogger.setUserProperty("installed_chrome_extension", "true");
		this.setState({showChromeExtensionCTA: false});
		window.localStorage["installed_chrome_extension"] = true;
	}

	_failHandler() {
		this.context.eventLogger.logEvent("ChromeExtensionInstallFailed");
		this.context.eventLogger.setUserProperty("installed_chrome_extension", "false");
		this.setState({showChromeExtensionCTA: true});
		window.localStorage.removeItem("installed_chrome_extension");
	}

	_handleClick() {
		this.context.eventLogger.logEvent("ChromeExtensionCTAClicked");
		if (global.chrome) {
			global.chrome.webstore.install("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", this._successHandler.bind(this), this._failHandler.bind(this));
		}
	}

	_canLinkPrivateGithub() {
		return this.context.githubToken && (!this.context.githubToken.scope || !(this.context.githubToken.scope.includes("repo") && this.context.githubToken.scope.includes("read:org") && this.context.githubToken.scope.includes("user:email")));
	}

	_renderGithub() {
		if (!this.context.githubToken) {
			return (
				<div styleName="setup-container">
					<a href={urlToGitHubOAuth} onClick={() => this.context.eventLogger.logEventForPage("InitiateGitHubOAuth2Flow", EventLocation.Dashboard, {scopes: "", upgrade: true})}>
						<div styleName="info-box">
							<div styleName="text-box">
								<img styleName="img" src={`${this.context.siteConfig.assetsRoot}/img/Dashboard/github-octocat.svg`}></img>
								<div styleName="title">Connect with GitHub</div>
								<div styleName="subtitle">GitHub integration combined with code intelligence is going to keep you in the flow as a developer.</div>
							</div>
						</div>
					</a>
				</div>
			);
		} else if (this._canLinkPrivateGithub()) {
			return (
				<a styleName="setup-container" href={urlToPrivateGitHubOAuth} onClick={() => this.context.eventLogger.logEventForPage("InitiateGitHubOAuth2Flow", EventLocation.Dashboard, {scopes: "read:org,repo,user:email", upgrade: true})}>
					<div styleName="info-box">
						<div styleName="text-box">
							<img styleName="img" src={`${this.context.siteConfig.assetsRoot}/img/Dashboard/github-octocat.svg`}/>
							<div styleName="title">Connect your repositories</div>
							<div styleName="subtitle">GitHub integration combined with code intelligence is going to keep you in the flow as a developer.</div>
						</div>
					</div>
				</a>
			);
		}

		return (
			<div styleName="setup-container">
				<div styleName="completed-box">
					<div styleName="text-box">
						<img styleName="img" src={`${this.context.siteConfig.assetsRoot}/img/Dashboard/github-octocat.svg`}></img>
						<div styleName="title">Connected</div>
						<div styleName="subtitle">You successfully linked your GitHub repositories.</div>
					</div>
				</div>
			</div>
		);
	}

	_renderChromeExtension() {
		if (!global.chrome) {
			return (<div styleName="setup-container">
				<div styleName="info-box">
					<div styleName="text-box">
						<img styleName="img" src={`${this.context.siteConfig.assetsRoot}/img/Dashboard/GoogleChromeAsset.svg`}></img>
						<div styleName="title">Get the Sourcegraph Extension</div>
						<div styleName="subtitle">Hop on Chrome to install the extension and get inline tooltip documentation and better code search for GitHub.</div>
					</div>
				</div>
			</div>);
		}


		if (this.state.showChromeExtensionCTA) {
			return (<div styleName="setup-container" onClick={this._handleClick.bind(this)}>
				<div styleName="info-box">
					<div styleName="text-box">
						<img styleName="img" src={`${this.context.siteConfig.assetsRoot}/img/Dashboard/GoogleChromeAsset.svg`}></img>
						<div styleName="title">Get the Chrome Extension</div>
						<div styleName="subtitle">Get inline tooltip documentation and better code search for GitHub.</div>
					</div>
				</div>
			</div>);
		}

		return (
			<div styleName="setup-container">
				<div styleName="completed-box">
					<div styleName="text-box">
						<img styleName="img" src={`${this.context.siteConfig.assetsRoot}/img/Dashboard/GoogleChromeAsset.svg`}></img>
						<div styleName="title">Installed</div>
						<div styleName="subtitle">You are all set to jump to definitions and search code on GitHub.</div>
					</div>
				</div>
			</div>);
	}

	render() {
		return (
				<div styleName="container">
					<div styleName="anon-section">
						<div styleName="anon-title">Get the tools, stay in flow</div>
						<div styleName="anon-header-sub">Sourcegraph lives anywhere you work with code. Get it on your GitHub repositories, in your browser, or at Sourcegraph.com.</div>
					</div>
					<div styleName="list">
						{this._renderGithub()}
						{this._renderChromeExtension()}
						<div styleName="setup-container">
							<div styleName="info-box">
								<Link to="/">
									<div styleName="text-box">
										<img styleName="img" src={`${this.context.siteConfig.assetsRoot}/img/Dashboard/SearchAssetBlue.svg`}></img>
										<div styleName="title">Search the global code graph</div>
										<div styleName="subtitle">Enter a function or definition to see its usage across all repositories.</div>
									</div>
								</Link>
							</div>
						</div>
					</div>
			</div>
		);
	}
}

export default CSSModules(ToolsContainer, styles);
