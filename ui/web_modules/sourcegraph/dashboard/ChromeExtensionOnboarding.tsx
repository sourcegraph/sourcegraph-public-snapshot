// tslint:disable: typedef ordered-imports

import * as React from "react";
import Helmet from "react-helmet";
import * as styles from "sourcegraph/dashboard/styles/Dashboard.css";
import {Button, Heading, Panel} from "sourcegraph/components/index";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import * as classNames from "classnames";
import * as typography from "sourcegraph/components/styles/_typography.css";
import * as base from "sourcegraph/components/styles/_base.css";
import * as colors from "sourcegraph/components/styles/_colors.css";
import {OnboardingExampleRefsContainer} from "sourcegraph/def/OnboardingExampleRefsContainer";
import {GitHubLogo} from "sourcegraph/components/symbols/index";

interface Props {
	location?: any;
	completeStep?: any;
}

type State = any;

export class ChromeExtensionOnboarding extends React.Component<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		siteConfig: React.PropTypes.object.isRequired,
		signedIn: React.PropTypes.bool.isRequired,
		router: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
	};

	constructor(props: Props) {
		super(props);
		this._installChromeExtensionClicked = this._installChromeExtensionClicked.bind(this);
	}

	_successHandler() {
		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_SUCCESS, "ChromeExtensionInstalled", {page_name: "ChromeExtensionOnboarding"});
		(this.context as any).eventLogger.setUserProperty("installed_chrome_extension", "true");
		setTimeout(() => document.dispatchEvent(new CustomEvent("sourcegraph:identify", (this.context as any).eventLogger.getAmplitudeIdentificationProps())), 10);
		this.props.completeStep();
	}

	_failHandler(msg) {
		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_ERROR, "ChromeExtensionInstallFailed", {page_name: "ChromeExtensionOnboarding"});
		(this.context as any).eventLogger.setUserProperty("installed_chrome_extension", "false");
		this.props.completeStep();
	}

	_installChromeExtensionClicked() {
		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLICK, "ChromeExtensionCTAClicked", {page_name: "ChromeExtensionOnboarding"});
		if (!!global.chrome) {
			(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLICK, "ChromeExtensionInstallStarted", {page_name: "ChromeExtensionOnboarding"});
			global.chrome.webstore.install("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", this._successHandler.bind(this), this._failHandler.bind(this));
		} else {
			(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLICK, "ChromeExtensionStoreRedirect", {page_name: "ChromeExtensionOnboarding"});
			window.location.assign("https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack");
		}
	}

	_skipClicked() {
		this.props.completeStep();
	}

	_exampleProps(): JSX.Element | null {
		const example =	{
			"location": {
			"pathname": "/github.com/gorilla/mux/-/info/GoPackage/github.com/gorilla/mux/-/NewRouter",
				"search": "",
				"hash": "",
				"state": "",
				"action": "PUSH",
				"key": "pahd2c",
				"query": {},
				"$searchBase": {
					"search": "",
					"searchBase": "",
				},
			},
			"repo": "github.com/gorilla/mux",
			"rev": "",
			"def": "GoPackage/github.com/gorilla/mux/-/NewRouter",
			"defObj": {
				"Repo": "github.com/gorilla/mux",
				"CommitID": "cf79e51a62d8219d52060dfc1b4e810414ba2d15",
				"UnitType": "GoPackage",
				"Unit": "github.com/gorilla/mux",
				"Path": "NewRouter",
				"Name": "NewRouter",
				"Kind": "func",
				"File": "mux.go",
				"DefStart": 281,
				"DefEnd": 383,
				"Exported": true,
				"Data": {
					"Exported": true,
					"PkgScope": true,
					"PkgName": "mux",
					"TypeString": "func() *github.com/gorilla/mux.Router",
					"UnderlyingTypeString": "func() *github.com/gorilla/mux.Router",
					"Kind": "func",
					"PackageImportPath": "github.com/gorilla/mux",
				},
				"Docs": [
					{
						"Format": "text/plain",
						"Data": "NewRouter returns a new router instance.\n",
					},
				],
				"TreePath": "./NewRouter",
				"DocHTML": {
					"__html": "<p>\nNewRouter returns a new router instance.\n</p>\n",
				},
				"FmtStrings": {
					"Name": {
						"Unqualified": "NewRouter",
						"ScopeQualified": "NewRouter",
						"DepQualified": "mux.NewRouter",
						"RepositoryWideQualified": "mux.NewRouter",
						"LanguageWideQualified": "github.com/gorilla/mux.NewRouter",
					},
					"Type": {
						"Unqualified": "() *Router",
						"ScopeQualified": "() *Router",
						"DepQualified": "() *mux.Router",
						"RepositoryWideQualified": "() *mux.Router",
						"LanguageWideQualified": "() *muxRouter",
					},
					"Language": "Go",
					"DefKeyword": "func",
					"Kind": "func",
				},
			},
			"examples": {
				"RepoRefs": [
					{
						"Repo": "github.com/10gen/evg-json",
						"Files": [
							{
								"Path": "json.go",
							},
						],
					},
				],
			},
		};

		let refLocs = example.examples;

		return (
			<div>
				{refLocs && refLocs.RepoRefs && refLocs.RepoRefs.map((repoRefs, i) => <OnboardingExampleRefsContainer
					location={example.location}
					key={i}
					refIndex={i}
					repo={example.repo}
					rev={example.rev}
					def={example.def}
					defObj={example.defObj}
					repoRefs={repoRefs}
					initNumSnippets={1}
					rangeLimit={1}
					fileCollapseThreshold={5} />)}
			</div>
		);
	}

	render(): JSX.Element | null {
		return (
			<div>
				<Helmet title="Home" />
				<div className={styles.onboarding_container}>
					<Panel className={classNames(base.pb3, base.ph4, base.ba, base.br2, colors.b__cool_pale_gray)}>
						<Heading className={classNames(base.pt4)} align="center" level="">
							Browse code smarter on Sourcegraph
						</Heading>
						<div className={styles.user_actions} style={{maxWidth: "360px"}}>
							<p className={classNames(typography.tc, base.mt3, base.mb2, typography.f6, colors.cool_gray_8)} >
								Hover over the code snippet below to view function definitions and documentation
							</p>
						</div>
						{this._exampleProps()}
						<div className={classNames(styles.user_actions)}>
						<div className={classNames(styles.inline_actions, base.pt3)} style={{verticalAlign: "top"}}>
							<GitHubLogo width={70} className={classNames(base.hidden_s)} style={{marginRight: "-20px"}}/>
							<img width={70} className={classNames(base.hidden_s)} src={`${(this.context as any).siteConfig.assetsRoot}/img/sourcegraph-mark.svg`}></img>
						</div>
						<div className={classNames(styles.inline_actions, base.pt2, base.pl3)} style={{maxWidth: "340px"}}>
							<Heading align="left" level="6">
								Want code intelligence while browsing GitHub?
							</Heading>
							<p className={classNames(typography.tl, base.mt3, typography.f6)}>
								Browse GitHub with instant documentation, jump to definition, and intelligent code search with the Sourcegraph for GitHub browser extension.
							</p>
						</div>
							<p>
								<a href="#install-chrome" onClick={this._installChromeExtensionClicked}><Button className={styles.action_link} type="button" color="blue">Install Sourcegraph for GitHub</Button></a>
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
