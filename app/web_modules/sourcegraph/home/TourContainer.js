import React from "react";
import {Link} from "react-router";
import CSSModules from "react-css-modules";
import {urlToGitHubOAuth} from "sourcegraph/util/urlTo";
import styles from "./styles/Tour.css";
import base from "sourcegraph/components/styles/_base.css";
import {Heading, Hero, Panel, Stepper, ChecklistItem, Button, Emoji} from "sourcegraph/components";
import redirectForUnauthedUser from "sourcegraph/user/redirectForUnauthedUser";

class TourContainer extends React.Component {

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
			showSourcegraphLiveCTA: !window.localStorage["installed_sourcegraph_live"],
			totalSteps: 4,
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
		this.context.eventLogger.logEventForPage("ChromeExtensionInstalled", "DashboardTour");
		this.context.eventLogger.setUserProperty("installed_chrome_extension", "true");
		this.setState({showChromeExtensionCTA: false});
		window.localStorage["installed_chrome_extension"] = true;
	}

	_failHandler() {
		this.context.eventLogger.logEventForPage("ChromeExtensionInstallFailed", "DashboardTour");
		this.context.eventLogger.setUserProperty("installed_chrome_extension", "false");
		this.setState({showChromeExtensionCTA: true});
		window.localStorage.removeItem("installed_chrome_extension");
	}

	_installChromeExtensionClicked() {
		this.context.eventLogger.logEventForPage("ChromeExtensionCTAClicked", "DashboardTour");
		if (global.chrome) {
			global.chrome.webstore.install("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", this._successHandler.bind(this), this._failHandler.bind(this));
		}
	}

	_connectGitHubClicked() {
		this.context.eventLogger.logEventForPage("InitiateGitHubOAuth2Flow", "DashboardTour", {scopes: "", upgrade: true});
		window.open(urlToGitHubOAuth);
	}

	_installSourcegraphLiveClicked() {
		this.context.eventLogger.logEventForPage("SourcegraphLiveCTAClicked", "DashboardTour");
		window.localStorage["installed_sourcegraph_live"] = true;
		this.setState({showSourcegraphLiveCTA: false});
		window.location.assign("https://github.com/sourcegraph/sourcegraph-sublime");
	}

	_completedStepsCounter() {
		let steps = 1;
		steps += this.context.githubToken ? 1 : 0;
		steps += this.state.showSourcegraphLiveCTA ? 0 : 1;
		steps += this.state.showChromeExtensionCTA ? 0 : 1;
		return steps;
	}

	render() {
		const tourComplete = this.state.totalSteps === this._completedStepsCounter();
		return (
			<div styleName="bg" className={base.pv6}>
				<div styleName="container-fixed outer">
					<Panel hoverLevel="high" className={base.pb5}>
						<Hero pattern="objects-fade" color="transparent" styleName="relative">
							{tourComplete && <Emoji name="rocket" width="64" styleName="rocket"/>}
							<div styleName="header-container container-fixed inner" className={`${base.ph2} ${base.pt5} ${base.pb5}`}>
								<Heading level="3" align="left" className={base.mv2} styleName="heading-container">
									{tourComplete ? "Your code is ready for the future" : "Get started with Sourcegraph"}
								</Heading>
								<Stepper steps={[null, null, null, null]} stepsComplete={this._completedStepsCounter()} color="green" className={base.mt3} styleName="stepper-container" />
							</div>
						</Hero>
						<div className={`${base.pv4} ${base.ph4}`} styleName="container-fixed inner">
							<ChecklistItem complete={true} className={base.mb5}>
								<Heading level="4">Create an account</Heading>
								<p className={base.mt2} styleName="cool-mid-gray">
									Good job! Your Sourcegraph account is ready to go.
								</p>
							</ChecklistItem>
							<ChecklistItem complete={this.context.githubToken !== null} className={base.mb5} actionText="Connect" actionOnClick={this._connectGitHubClicked.bind(this)}>
								<Heading level="4">Connect with GitHub</Heading>
								<p className={base.mt2} styleName="cool-mid-gray">
									Connecting your account with GitHub puts Sourcegraph’s code intelligence to work on your private repositories.
								</p>
							</ChecklistItem>
							<ChecklistItem complete={!this.state.showChromeExtensionCTA} className={base.mb5} actionText="Install" actionOnClick={this._installChromeExtensionClicked.bind(this)}>
								<Heading level="4">Install a browser extension</Heading>
								<p className={base.mt2} styleName="cool-mid-gray">
									Browse GitHub like an IDE, with jump-to-definition links, semantic code search, and documentation tooltips.
								</p>
							</ChecklistItem>
							<ChecklistItem complete={!this.state.showSourcegraphLiveCTA} actionText="Install" actionOnClick={this._installSourcegraphLiveClicked.bind(this)}>
								<Heading level="4">Add Sourcegraph to your editor</Heading>
								<p className={base.mt2} styleName="cool-mid-gray">
									See global cross-references and live usage examples for Go code, as you type. For Sublime Text 3.
								</p>
							</ChecklistItem>
						</div>

						{tourComplete && <p styleName="center" className={`${base.pt4}`}>
							<Link to="/">
								<Button color="green" className={`${base.ph4}`}>Explore Sourcegraph</Button>
							</Link>
						</p>}
					</Panel>
				</div>
			</div>
		);
	}
}

export default redirectForUnauthedUser("/", CSSModules(TourContainer, styles, {allowMultiple: true}));
