import React from "react";
import DashboardModal from "./DashboardModal";
import styles from "./styles/DashboardModal.css";
import CSSModules from "react-css-modules";
import EventLogger, {EventLocation} from "sourcegraph/util/EventLogger";
import {Button} from "sourcegraph/components";
import {GitHubIcon} from "sourcegraph/components/Icons";
import {urlToGitHubOAuth} from "sourcegraph/util/urlTo";
import ChromeExtensionCTA from "./LiteChromeExtensionCTA";
import {Link} from "react-router";
import Component from "sourcegraph/Component";

class OnboardingModals extends Component {
	static propTypes = {
		canShowChromeExtensionCTA: React.PropTypes.bool,
		onboardingFlow: React.PropTypes.string.isRequired,
	}

	static contextTypes = {
		user: React.PropTypes.object,
		githubToken: React.PropTypes.object,
		siteConfig: React.PropTypes.object.isRequired,
	}

	constructor(props) {
		super(props);
		this.state = {
			canShowChromeExtensionCTA: props.canShowChromeExtensionCTA,
			onboardingFlowState: props.onboardingFlow,
			gitHubAuthed: false,
		};

		EventLogger.logEventForPage("OnboardingModalsViewed", EventLocation.Onboarding);
	}

	reconcileState(state, props, context) {
		Object.assign(state, props);
		state.gitHubAuthed = context.githubToken;

		if (state.onboardingFlowState) {
			localStorage["onboard-state"] = state.onboardingFlowState;
		} else {
			localStorage.removeItem("onboard-state");
		}

		state.modals = {
			"new-user": {
				header: "Start using Sourcegraph",
				onClick: this.clickHandler.bind(this),
				subheader: "Welcome!",
				body: "Let’s quickly walk through how Sourcegraph's code intelligence can help you while writing code. Do you want to use it on your private code, or on open-source code?",
				hasNext: true,
				primaryCTA: this._onboardSelectionCTA.bind(this),
			},
			"open-source-1": {
				header: "Open source it is!",
				onClick: this.clickHandler.bind(this, "open-source-2"),
				subheader: "Stop drowning in browser tabs",
				body: "Sourcegraph will save you time with live usage examples for code, drawn from hundreds of thousands of open source repositories.",
				hasNext: true,
				img: this._onboardCheckmarkImages.bind(this),
				primaryCTA: this._bigCTA.bind(this, "Get the tools", "ContinueCTAClicked", this.state.canShowChromeExtensionCTA ? "open-source-2" : "open-source-3"),
			},
			"open-source-2": {
				header: "Reading code on GitHub? No problem",
				subheader: this.state.canShowChromeExtensionCTA ? "Get configured" : "Looks like you already have the Chrome Extension!",
				body: this.state.canShowChromeExtensionCTA ? "The Sourcegraph Chrome Extension gives you all the benefits of Sourcegraph when you're on GitHub. It even works on public and private code!" : "Make sure you've enabled the extension so you can get the benefits of Sourcegraph when you're on GitHub viewing public or private code.",
				hasNext: true,
				onClick: this.clickHandler.bind(this, "open-source-3"),
				primaryCTA: this.state.canShowChromeExtensionCTA ? this._chromeExtensionCTA.bind(this, "open-source-3") : null,
				secondaryCTA: this._skipCTA.bind(this, this.state.canShowChromeExtensionCTA ? "Skip" : "Continue", "open-source-3", this.state.ChromeExtensionCTA ? "SkipCTAClicked" : "ContinueCTAClicked"),
			},
			"open-source-3": {
				header: "Almost Done!",
				subheader: "Sourcegraph pairs perfectly with GitHub",
				body: this.state.gitHubAuthed ? "Since you've connected GitHub, you can use Sourcegraph on your own repositories for live usage examples and intelligent code browsing." : "GitHub integration combined with code intelligence is going to keep you in the flow as a developer. Continue with GitHub so you can give Code Intelligence a test drive on your own code.",
				hasNext: true,
				onClick: this.clickHandler.bind(this),
				primaryCTA: this.state.gitHubAuthed ? this._skipCTA.bind(this, "Continue", "open-source-4", "ContinueCTAClicked") : this._gitHubCTA.bind(this, "open-source-4"),
				secondaryCTA: !this.state.gitHubAuthed ? this._skipCTA.bind(this, "Skip", "open-source-4", "SkipCTAClicked") : null,
			},
			"open-source-4": {
				header: "Navigate the Graph!",
				subheader: "You're all set to experience code intelligence anywhere!",
				body: "Check out usage and examples across all repositories.",
				hasNext: false,
				onClick: this.clickHandler.bind(this),
				primaryCTA: this._goLangCTALink.bind(this),
				secondaryCTA: this._skipCTA.bind(this, "Later", null, "ContinueCTAClicked"),
			},
			"private-repo-1": {
				header: this.state.gitHubAuthed ? "You're connected!" : "Got GitHub?",
				onClick: this.clickHandler.bind(this, "private-repo-2"),
				subheader: this.state.gitHubAuthed ? "Sourcegraph pairs perfectly with GitHub" : "We built the tools you need to stay in developer flow",
				body: this.state.gitHubAuthed ? "You're building the perfect experience! Since you've connected GitHub, you'll be able to use Sourcegraph on your own repositories for live usage examples and intelligent code browsing!" : "Before we can get you started with Code Intelligence on your own repositories, you'll need to continue with GitHub.",
				hasNext: true,
				primaryCTA: this.state.gitHubAuthed ? this._skipCTA.bind(this, "Let's get started", this.state.canShowChromeExtensionCTA ? "private-repo-2" : "private-repo-3", "ContinueCTAClicked") : this._gitHubCTA.bind(this, "private-repo-2"),
				secondaryCTA: !this.state.gitHubAuthed ? this._skipCTA.bind(this, "Skip", this.state.canShowChromeExtensionCTA ? "private-repo-2" : "private-repo-3", "SkipCTAClicked") : null,
			},
			"private-repo-2": {
				header: "Get configured",
				subheader: this.state.canShowChromeExtensionCTA ? "Get configured" : "Looks like you already have the Chrome extension!",
				body: this.state.canShowChromeExtensionCTA ? "The Sourcegraph Chrome extension gives you all the benefits of Sourcegraph when you're on GitHub, on both public and private code." : "Make sure you've enabled the extension so you can get the benefits of Sourcegraph when you're on GitHub viewing public or private code.",
				hasNext: true,
				onClick: this.clickHandler.bind(this, "private-repo-3"),
				primaryCTA: this.state.canShowChromeExtensionCTA ? this._chromeExtensionCTA.bind(this, "private-repo-3") : null,
				secondaryCTA: this._skipCTA.bind(this, this.state.canShowChromeExtensionCTA ? "Later" : "Continue", "private-repo-3", this.state.canShowChromeExtensionCTA ? "SkipCTAClicked" : "ContinueCTAClicked"),
			},
			"private-repo-3": {
				header: "Great work!",
				subheader: "You're all set to use Sourcegraph anywhere!",
				body: "Start off by navigating through your own repositories.",
				hasNext: false,
				onClick: this.clickHandler.bind(this),
				primaryCTA: this._bigCTA.bind(this, "Take me to my dashboard", "OnboardingModalsDashboardCTAClicked"),
			},
		};
	}

	clickHandler(nextPath, actionName) {
		if (!nextPath) {
			this.setState({
				onboardingFlowState: null,
			});

			EventLogger.logEventForPage("OnboardingSequenceCompleted", EventLocation.Onboarding, {CurrentModal: this.state.onboardingFlowState, CTA: "Dismiss", GitHubAuthed: this.state.gitHubAuthed ? "true" : "false"});
			return;
		}

		if (actionName) {
			EventLogger.logEventForPage(actionName, EventLocation.Onboarding, {CurrentModal: this.state.onboardingFlowState, GitHubAuthed: this.state.gitHubAuthed ? "true" : "false"});
		} else {
			EventLogger.logEventForPage("OnboardingModalCTAClicked", EventLocation.Onboarding, {CurrentModal: this.state.onboardingFlowState, GitHubAuthed: this.state.gitHubAuthed ? "true" : "false"});
		}

		this.setState({
			onboardingFlowState: nextPath,
		});
	}

	_onboardCheckmarkImages() {
		return (
				<img styleName="feature-list" src={`${this.context.siteConfig.assetsRoot}/img/Dashboard/SourcegraphFeatureList.svg`}></img>
		);
	}

	_onboardSelectionCTA() {
		return (
			<div>
				<div styleName="option" onClick={this.clickHandler.bind(this, "private-repo-1", "PersonalRepoPathClicked")}>
					<p styleName="title">Personal Projects</p>
					<p styleName="subtitle">I want to see Code Intelligence on my personal repositories</p>
				</div>
				<div styleName="option" onClick={this.clickHandler.bind(this, "open-source-1", "OpenSourcePathClicked")}>
					<p styleName="title">Open Source</p>
					<p styleName="subtitle">Sourcegraph has 100% coverage on all Go Repositories</p>
				</div>
			</div>
		);
	}

	_gitHubCTA(nextPath) {
		return (
			<div styleName="cta">
				<a href={urlToGitHubOAuth} onClick={() => {
					EventLogger.logEventForPage("OnboardingModalGithubCTAClicked", EventLocation.Onboarding);
				}}>
					<Button outline={true} color="warning"><GitHubIcon style={{marginRight: "10px", fontSize: "16px"}} />Continue with GitHub</Button>
				</a>
			</div>
		);
	}

	_chromeExtensionCTA(nextPath) {
		return (<div styleName="cta">
					<ChromeExtensionCTA onSuccess={() => {
						this.setState({showChromeExtensionCTA: false});
						this.clickHandler.bind(this, nextPath, "OnboardingModalChromeExtensionInstallSuccess");
					}}/>
				</div>);
	}

	_bigCTA(copy, actionName, nextPath) {
		return (
			<div styleName="cta">
				<Button onClick={this.clickHandler.bind(this, nextPath, actionName)} color="primary" size="large" unspaced={true} lowercase={true}>{copy}</Button>
			</div>
		);
	}

	_goLangCTALink() {
		return (
			<div styleName="cta">
				<Link to="github.com/golang/go/-/def/GoPackage/net/http/-/NewRequest/-/info" onClick={() => {
					EventLogger.logEventForPage("OnboardingSequenceCompleted", EventLocation.Onboarding, {CurrentModal: this.state.onboardingFlowState, CTA: "Dismiss", GitHubAuthed: this.state.gitHubAuthed});
					EventLogger.logEventForPage("GoHTTPDefRefsCTAClicked", EventLocation.Onboarding);
				}}>
				<Button color="primary" size="large" unspaced={true} lowercase={true}>See usage examples for http.NewRequest &raquo;</Button>
				</Link>
			</div>
		);
	}

	/* MARK: SECONDARY CTAs */
	_skipCTA(copy, nextPath, actionName) {
		return (
			<div styleName="cta">
				<p styleName="skip" onClick={this.clickHandler.bind(this, nextPath, actionName)}>{copy}</p>
			</div>
		);
	}

	render() {
		let modal = this.state.modals[this.state.onboardingFlowState];
		return (
			<div styleName="container">
				{modal && <DashboardModal img={modal.img ? modal.img : null} primaryCTA={modal.primaryCTA ? modal.primaryCTA : null} secondaryCTA={modal.secondaryCTA ? modal.secondaryCTA : null} onClick={modal.onClick} header={modal.header} subheader={modal.subheader} body={modal.body} hasNext={modal.hasNext}/>}
			</div>
		);
	}
}

export default CSSModules(OnboardingModals, styles);
