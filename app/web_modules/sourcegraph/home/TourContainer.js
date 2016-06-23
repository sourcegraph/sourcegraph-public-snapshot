import React from "react";
import {Link} from "react-router";
import CSSModules from "react-css-modules";
import {urlToGitHubOAuth} from "sourcegraph/util/urlTo";
import styles from "./styles/Tour.css";
import base from "sourcegraph/components/styles/_base.css";
import {Heading, Hero, Panel, Stepper, ChecklistItem, Button, Emoji} from "sourcegraph/components";
import redirectForUnauthedUser from "sourcegraph/user/redirectForUnauthedUser";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

class TourContainer extends React.Component {
	static propTypes = {
		location: React.PropTypes.object.isRequired,
	}

	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
		githubToken: React.PropTypes.object,
		eventLogger: React.PropTypes.object.isRequired,
		router: React.PropTypes.object.isRequired,
	}

	constructor(props) {
		super(props);
		this.state = {
			showBrowserExtensionCTA: global.chrome ? !document.getElementById("sourcegraph-app-bootstrap") : !window.localStorage["explored_browser_extensions"],
			showSourcegraphLiveCTA: !window.localStorage["installed_sourcegraph_live"],
			totalSteps: 4,
		};
	}

	componentDidMount() {
		if (global.chrome) {
			setTimeout(() => this.setState({
				showBrowserExtensionCTA: !document.getElementById("sourcegraph-app-bootstrap"),
			}), 1);
		}
	}

	_successHandler() {
		this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_TOOLS, AnalyticsConstants.ACTION_SUCCESS, "ChromeExtensionInstalled", {page_name: AnalyticsConstants.PAGE_TOUR});
		this.context.eventLogger.setUserProperty("installed_chrome_extension", "true");
		this.setState({showBrowserExtensionCTA: false});
	}

	_failHandler() {
		this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_TOOLS, AnalyticsConstants.ACTION_ERROR, "ChromeExtensionInstallFailed", {page_name: AnalyticsConstants.PAGE_TOUR});
		this.context.eventLogger.setUserProperty("installed_chrome_extension", "false");
		this.setState({showBrowserExtensionCTA: true});
	}

	_browserExtensionCTAClicked() {
		this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_TOOLS, AnalyticsConstants.ACTION_CLICK, "ChromeExtensionCTAClicked", {page_name: AnalyticsConstants.PAGE_TOUR});
		if (typeof chrome !== "undefined" && global.chrome && global.chrome.webstore) {
			global.chrome.webstore.install("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", this._successHandler.bind(this), this._failHandler.bind(this));
		} else {
			window.localStorage["explored_browser_extensions"] = "true";
			this.setState({showBrowserExtensionCTA: false});
			window.location.assign("/tools/browser");
		}
	}

	_connectGitHubClicked() {
		this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "InitiateGitHubOAuth2Flow", {page_name: AnalyticsConstants.PAGE_TOUR, scopes: "", upgrade: true});
		window.open(urlToGitHubOAuth(null, this.props.location));
	}

	_installSourcegraphLiveClicked() {
		this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_TOOLS, AnalyticsConstants.ACTION_CLICK, "SourcegraphLiveCTAClicked", {page_name: AnalyticsConstants.PAGE_TOUR});
		window.localStorage["installed_sourcegraph_live"] = true;
		this.setState({showSourcegraphLiveCTA: false});
		window.location.assign("/tools/editor");
	}

	_completedStepsCounter() {
		let steps = 1;
		steps += this.context.githubToken ? 1 : 0;
		steps += this.state.showSourcegraphLiveCTA ? 0 : 1;
		steps += this.state.showBrowserExtensionCTA ? 0 : 1;
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
									See your public contributions and grant access to use Sourcegraph on your private repositories.
								</p>
							</ChecklistItem>
							<ChecklistItem complete={!this.state.showBrowserExtensionCTA} className={base.mb5} actionText="Install" actionOnClick={this._browserExtensionCTAClicked.bind(this)}>
								<Heading level="4">Install a browser extension</Heading>
								<p className={base.mt2} styleName="cool-mid-gray">
									Browse GitHub like an IDE, with jump-to-definition links, semantic code search, and instant documentation.
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
