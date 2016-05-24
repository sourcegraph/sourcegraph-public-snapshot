import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/Tools.css";
import base from "sourcegraph/components/styles/_base.css";
import {Heading, Hero, Panel, Button} from "sourcegraph/components";
import redirectForUnauthedUser from "sourcegraph/user/redirectForUnauthedUser";

class ToolsContainer extends React.Component {

	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
		router: React.PropTypes.object.isRequired,
	}

	constructor(props) {
		super(props);
		this.state = {
			showChromeExtensionCTA: !window.localStorage["installed_chrome_extension"],
			showSourcegraphLiveCTA: !window.localStorage["installed_sourcegraph_live"],
		};
	}

	componentDidMount() {
		this.timeout = setTimeout(() => this.setState({
			showChromeExtensionCTA: !document.getElementById("sourcegraph-app-bootstrap") && !window.localStorage["installed_chrome_extension"],
		}), 1);
	}

	componentWillUnmount() {
		clearTimeout(this.timeout);
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

	_installSourcegraphLiveClicked() {
		this.context.eventLogger.logEventForPage("SourcegraphLiveCTAClicked", "DashboardTools");
		window.localStorage["installed_sourcegraph_live"] = true;
		this.setState({showSourcegraphLiveCTA: false});
		window.location.assign("https://github.com/sourcegraph/sourcegraph-sublime");
	}

	render() {
		return (
			<div styleName="container">
				<Hero color="purple" pattern="objects">
					<div styleName="container-fixed">
						<Heading level="1" color="white" underline="white">Get code intelligence in every part of your workflow</Heading>
						<p style={{maxWidth: "560px"}} className={base.center}>
							Add Sourcegraph to everywhere you write code.
						</p>
					</div>
				</Hero>
				<div styleName="panel-container">
					<div styleName="panel-item">
						<Panel hoverLevel="high">
							<img styleName="img" src={`${this.context.siteConfig.assetsRoot}/img/Dashboard/GoogleChromeAsset.svg`}></img>
							<Heading align="center" level="4" className={base.ph4}>Sourcegraph for Chrome</Heading>
							<p style={{color: "rgba(119, 147, 174, 1)"}} className={base.ph4}>
								Smart search and instant documentation on GitHub.
							</p>
							<div styleName="button-container">
								<Button onClick={this._installChromeExtensionClicked.bind(this)} outline={this.state.showChromeExtensionCTA} color="purple">{this.state.showChromeExtensionCTA ? "Install" : "Installed"}</Button>
							</div>
						</Panel>
					</div>
					<div styleName="panel-item">
						<Panel hoverLevel="high">
							<img styleName="img" src={`${this.context.siteConfig.assetsRoot}/img/Dashboard/SublimeTextLogo.svg`}></img>
							<Heading align="center" level="4" className={base.ph4}>Sourcegraph for Sublime Text</Heading>
							<p style={{color: "rgba(119, 147, 174, 1)"}} className={base.ph4}>
								View examples instantly as you write code.
							</p>
							<div styleName="button-container">
								<Button onClick={this._installSourcegraphLiveClicked.bind(this)} outline={this.state.showSourcegraphLiveCTA} color="purple">{this.state.showSourcegraphLiveCTA ? "Install" : "Installed"}</Button>
							</div>
						</Panel>
					</div>
				</div>
			</div>
		);
	}
}

export default redirectForUnauthedUser("/", CSSModules(ToolsContainer, styles));
