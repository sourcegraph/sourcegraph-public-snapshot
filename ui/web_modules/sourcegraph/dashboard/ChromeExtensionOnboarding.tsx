import * as classNames from "classnames";
import * as React from "react";
import {context} from "sourcegraph/app/context";
import {Button, Heading, Panel} from "sourcegraph/components";
import {PageTitle} from "sourcegraph/components/PageTitle";
import * as base from "sourcegraph/components/styles/_base.css";
import * as colors from "sourcegraph/components/styles/_colors.css";
import * as typography from "sourcegraph/components/styles/_typography.css";
import {GitHubLogo} from "sourcegraph/components/symbols";
import {colors as jsColors, whitespace} from "sourcegraph/components/utils";
import {EditorDemo} from "sourcegraph/dashboard/EditorDemo";
import * as styles from "sourcegraph/dashboard/styles/Dashboard.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";

interface Props {
	location?: any;
	completeStep?: any;
}

type State = any;

export class ChromeExtensionOnboarding extends React.Component<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	constructor(props: Props) {
		super(props);
		this._installChromeExtensionClicked = this._installChromeExtensionClicked.bind(this);
	}

	_successHandler(): void {
		AnalyticsConstants.Events.ChromeExtension_Installed.logEvent({page_name: "ChromeExtensionOnboarding"});
		EventLogger.setUserProperty("installed_chrome_extension", "true");
		// Syncs the our site analytics tracking with the chrome extension tracker.
		EventLogger.updateTrackerWithIdentificationProps();
		this._continueOnboarding();
	}

	_failHandler(): void {
		AnalyticsConstants.Events.ChromeExtensionInstall_Failed.logEvent({page_name: "ChromeExtensionOnboarding"});
		EventLogger.setUserProperty("installed_chrome_extension", "false");
	}

	_installChromeExtensionClicked(): void {
		AnalyticsConstants.Events.ChromeExtensionCTA_Clicked.logEvent({page_name: "ChromeExtensionOnboarding"});

		if (!!global.chrome) {
			AnalyticsConstants.Events.ChromeExtensionInstall_Started.logEvent({page_name: "ChromeExtensionOnboarding"});
			global.chrome.webstore.install("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", this._successHandler.bind(this), this._failHandler.bind(this));
		} else {
			AnalyticsConstants.Events.ChromeExtensionStore_Redirected.logEvent({page_name: "ChromeExtensionOnboarding"});
			window.open("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", "_newtab");
		}
	}

	_skipClicked(): void {
		AnalyticsConstants.Events.ChromeExtensionSkipCTA_Clicked.logEvent({page_name: "ChromeExtensionOnboarding"});
		this._continueOnboarding();
	}

	_continueOnboarding(): void {
		this.props.completeStep();
	}

	_exampleProps(): JSX.Element | null {
		return <EditorDemo repo="github.com/gorilla/mux" rev="master" path="mux.go" startLine={211} />;
	}

	_isBrowserSupported(): boolean {
		const isChrome = navigator.userAgent.includes("Chrome");
		const isMobile = /Android|iPhone|iPad|iPod|Mobile/i.test(navigator.userAgent);

		if (isMobile || !isChrome) {
			AnalyticsConstants.Events.ChromeExtensionUnsupportedBrowser_Failed.logEvent({page_name: "ChromeExtensionOnboarding"});
			return false;
		}

		return true;
	}

	_isChromeExtensionInstalled(): boolean {
		// Check for a invisible element injected by the extension on https://(www.)?sourcegraph.com/*
		return document.getElementById("sourcegraph-app-bootstrap") !== null;
	}

	render(): JSX.Element | null {
		if (!this._isBrowserSupported() || this._isChromeExtensionInstalled()) {
			this._continueOnboarding();
			return null;
		}

		return (
			<div>
				<PageTitle title="Home" />
				<div className={styles.onboarding_container}>
					<Panel className={classNames(base.pb3, base.ph4, base.ba, base.br2, colors.b__cool_pale_gray)}>
						<Heading style={{paddingTop: whitespace[4]}} align="center" level={3}>
							Browse code smarter on Sourcegraph
						</Heading>
						<div className={styles.user_actions} style={{ maxWidth: "360px" }}>
							<p className={classNames(typography.tc, base.mt3, base.mb2, typography.f6, colors.cool_gray_8)} >
								Hover over the code snippet below to view function definitions and documentation
							</p>
						</div>
						{this._exampleProps()}
						<div className={classNames(styles.user_actions)}>
						<div className={classNames(styles.inline_actions, base.pt3)} style={{verticalAlign: "top"}}>
							<GitHubLogo width={70} color={jsColors.coolGray3()} className={classNames(base.hidden_s)} style={{marginRight: "-20px", verticalAlign: "top"}}/>
							<img width={70} className={classNames(base.hidden_s)} src={`${context.assetsRoot}/img/sourcegraph-mark.svg`}></img>
						</div>
						<div className={classNames(styles.inline_actions, base.pt2, base.pl3)} style={{maxWidth: "340px"}}>
							<Heading align="left" level={6}>
								Want code intelligence while browsing GitHub?
							</Heading>
								<p className={classNames(typography.tl, base.mt3, typography.f6)}>
									Browse GitHub with instant documentation, jump to definition, and intelligent code search with the Sourcegraph for GitHub browser extension.
							</p>
							</div>
							<p>
								<Button onClick={this._installChromeExtensionClicked} className={styles.action_link} type="button" color="blue">Install Sourcegraph for GitHub</Button>
							</p>
							<p>
								<a onClick={this._skipClicked.bind(this)}>Skip</a>
							</p>
						</div>
					</Panel>
				</div>
			</div>
		);
	}
}
