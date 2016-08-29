// tslint:disable: typedef ordered-imports

import * as React from "react";
import {LocationStateToggleLink} from "sourcegraph/components/LocationStateToggleLink";
import Helmet from "react-helmet";
import {Container} from "sourcegraph/Container";
import {UserStore} from "sourcegraph/user/UserStore";
import * as styles from "sourcegraph/dashboard/styles/Dashboard.css";
import {locationForSearch} from "sourcegraph/search/routes";
import {GlobalSearchInput} from "sourcegraph/search/GlobalSearchInput";
import {Button, Logo} from "sourcegraph/components/index";
import {SearchSettings} from "sourcegraph/search/SearchSettings";
import * as invariant from "invariant";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import * as classNames from "classnames";
import {OnboardingContainer} from "sourcegraph/dashboard/OnboardingContainer";
import {Store} from "sourcegraph/Store";

type OnSelectQueryListener = (ev: React.MouseEvent<HTMLButtonElement>, query: string) => any;

interface Props {
	location?: any;
	route?: any;
}

type State = any;

export class DashboardContainer extends Container<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		siteConfig: React.PropTypes.object.isRequired,
		signedIn: React.PropTypes.bool.isRequired,
		router: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
	};

	_input: any;

	constructor(props: Props) {
		super(props);

		this.state = {
			isTyping: false,
			chromeExtensionInstalled: false,
		};

		this._handleInput = this._handleInput.bind(this);
		this._onSelectQuery = this._onSelectQuery.bind(this);
		this._installChromeExtensionClicked = this._installChromeExtensionClicked.bind(this);
	}

	stores(): Store<any>[] {
		return [UserStore];
	}

	componentDidMount() {
		super.componentDidMount();
		(this.context as any).router.setRouteLeaveHook(this.props.route, (route: any) => {
			if (Boolean(this._shouldStartOnboarding()) && route.pathname.includes("/info/")) {
				return false;
			}

			return true;
		});

		if (UserStore.authResponses["signup"]) {
			const newLoc = Object.assign({}, this.props.location, {query: {ob: "chrome"}});
			(this.context as any).router.replace(newLoc);
		}

		setTimeout(() => this.setState({
			chromeExtensionInstalled: document.getElementById("sourcegraph-app-bootstrap"),
		}), 10);
	}

	reconcileState(state, props, context) {
		Object.assign(state, props);

		const settings = UserStore.settings;
		state.langs = settings && settings.search ? settings.search.languages : null;
		state.scope = settings && settings.search ? settings.search.scope : null;
		state.signedIn = context && context.signedIn;
	}

	_shouldStartOnboarding(): string | null {
		if (this.props.location && (this.props.location.query["ob"] === "chrome" || this.props.location.query["ob"] === "github")) {
			return this.props.location.query["ob"];
		}

		return null;
	}

	onStateTransition(prevState: State, nextState: State): void {
		if (this._input && this._input.value && !prevState.isTyping) {
			this._goToSearch(this._input.value);
		}
	}

	_goToSearch(query: string) {
		(this.context as any).router.push(locationForSearch(this.props.location, query, this.state.langs, this.state.scope, true, true));
	}

	_handleInput(ev: React.FormEvent<HTMLInputElement>) {
		if (!(ev.currentTarget instanceof HTMLInputElement)) {
			return;
		}
		if (ev.currentTarget.value) {
			this._goToSearch(ev.currentTarget.value);
		}
	}

	_onSelectQuery(ev: React.MouseEvent<HTMLButtonElement>, query: string) {
		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_GLOBAL_SEARCH, AnalyticsConstants.ACTION_CLICK, "ExistingQueryClicked", {query: query, languages: this.state.langs, page_name: AnalyticsConstants.PAGE_DASHBOARD});

		invariant(this._input, "no input field");

		// Make it feel more realistic.
		const delay = (c: string) => 20 + (25 * Math.random()) + (c === " " ? 75 : 0);
		let simulateTyping;
		simulateTyping = (v: string, i: number = 0, then: Function) => {
			if (i >= v.length) {
				this.setState({isTyping: false});
				then();
				return;
			}
			const c = v.charAt(i);
			this._input.value += c;
			setTimeout(() => simulateTyping(v, i + 1, then), delay(c));
		};
		if (!this.state.isTyping) {
			this._input.focus();
			this._input.value = "";
			this.setState({isTyping: true}, () => simulateTyping(query, 0, () => setTimeout(() => this._goToSearch(query), 300)));
		}
	}

	_successHandler() {
		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_DASHBOARD, AnalyticsConstants.ACTION_SUCCESS, "ChromeExtensionInstalled", {page_name: AnalyticsConstants.PAGE_DASHBOARD});
		(this.context as any).eventLogger.setUserProperty("installed_chrome_extension", "true");
		this.setState({showChromeExtensionCTA: false, onboarding: null});
		setTimeout(() => document.dispatchEvent(new CustomEvent("sourcegraph:identify", (this.context as any).eventLogger.getAmplitudeIdentificationProps())), 10);
	}

	_failHandler(msg) {
		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_DASHBOARD, AnalyticsConstants.ACTION_ERROR, "ChromeExtensionInstallFailed", {page_name: AnalyticsConstants.PAGE_DASHBOARD});
		(this.context as any).eventLogger.setUserProperty("installed_chrome_extension", "false");
		this.setState({showChromeExtensionCTA: true, onboarding: null});
	}

	_installChromeExtensionClicked() {
		(this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_DASHBOARD, AnalyticsConstants.ACTION_CLICK, "ChromeExtensionCTAClicked", {page_name: AnalyticsConstants.PAGE_DASHBOARD});
		if (global.chrome) {
			global.chrome.webstore.install("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", this._successHandler.bind(this), this._failHandler.bind(this));
		}
	}

	_defaultAuthedDashboard(): JSX.Element | null {
		const langSelected = this.state.langs && this.state.langs.length > 0;
		return (
			<div>
				<Helmet title="Home" />
				<div className={styles.home_container}>
					<Logo type="logotype" className={styles.logo} />

					<h2 className={styles.description}>
						<strong>Instant&nbsp;usage&nbsp;examples and&nbsp;more&nbsp;as&nbsp;you&nbsp;code,</strong> automatically&nbsp;drawn&nbsp;from public and (your&nbsp;own)&nbsp;private&nbsp;code.
					</h2>

					<div className={styles.user_actions}>
						{!(this.context as any).signedIn && <LocationStateToggleLink href="/login" modalName="login" location={this.props.location}><Button className={styles.action_link} type="button" color="blue" outline={true}>Sign in</Button></LocationStateToggleLink>}
						{!this.state.chromeExtensionInstalled && <Button onClick={this._installChromeExtensionClicked} className={styles.action_link} type="button" color="blue" outline={true}>Install Chrome extension</Button>}
					</div>

					<GlobalSearchInput
						name="q"
						query={this.props.location.query.q || ""}
						autoFocus={true}
						domRef={e => this._input = e}
						className={styles.search_input}
						onChange={this._handleInput} />
					<div className={styles.search_actions}>
						<Button className={styles.search_button} type="button" color="blue">Find usage examples</Button>
					</div>

					{<TitledSection title="Top searches" className={styles.top_queries_panel}>
						{langSelected && <Queries langs={this.state.langs} onSelectQuery={this._onSelectQuery} />}
						{!langSelected && <p className={styles.notice}>Select a language below to get started.</p>}
					</TitledSection>}

					{<TitledSection title="Search options" className={styles.search_settings_panel}>
						<SearchSettings location={this.props.location} className={styles.search_settings} />
					</TitledSection>}
				</div>
			</div>
		);
	}

	render(): JSX.Element | null {
		let onboarding = this._shouldStartOnboarding();
		const conditionalDashboardRender = onboarding ? <OnboardingContainer currentStep={onboarding} location={this.props.location}/> : this._defaultAuthedDashboard();
		return (
			<div>
				{conditionalDashboardRender}
			</div>
		);
	}
}

const TitledSection = ({
	title,
	children,
	className,
}: {
	title: string,
	children?: any,
	className: string,
}) => (
	<div className={classNames(styles.titled_section, className)}>
		<div className={styles.section_title}>{title}</div>
		{children}
	</div>
);

const topQueries: {[key: string]: string[]} = {
	javascript: [
		"leftpad",
	],
	python: [
		"django render response",
		"argument parser",
		"os.path.relpath",
	],
	java: [
		"file open",
		"date time",
	],
	golang: [
		"new http request",
		"read file",
		"json encoder",
		"http get",
		"sql query",
		"indent json",
	],
};
function topQueriesFor(langs: string[]): string[] {
	let qs = [];
	for (let lang of langs) {
		if (topQueries[lang]) {
			qs.push.apply(qs, topQueries[lang]);
		}
	}
	return qs;
}
const Queries = ({
	langs,
	onSelectQuery,
}: {
	langs: string[],
	onSelectQuery: OnSelectQueryListener,
}) => (
	<div>{topQueriesFor(langs).map(q => <Button onClick={(ev) => onSelectQuery(ev, q)} key={q} className={styles.query} color="blue" outline={true} size="small">{q}</Button>)}</div>
);
