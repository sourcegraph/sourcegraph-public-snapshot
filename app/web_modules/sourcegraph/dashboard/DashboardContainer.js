// @flow

import React from "react";
import {Link} from "react-router";
import LocationStateToggleLink from "sourcegraph/components/LocationStateToggleLink";
import Helmet from "react-helmet";
import CSSModules from "react-css-modules";
import Container from "sourcegraph/Container";
import UserStore from "sourcegraph/user/UserStore";
import styles from "./styles/Dashboard.css";
import {locationForSearch} from "sourcegraph/search/routes";
import GlobalSearchInput from "sourcegraph/search/GlobalSearchInput";
import {Button, Logo} from "sourcegraph/components";
import SearchSettings from "sourcegraph/search/SearchSettings";
import type {LanguageID} from "sourcegraph/Language";
import invariant from "invariant";

type OnSelectQueryListener = (ev: Event, query: string) => mixed;

class DashboardContainer extends Container {
	static propTypes = {
		location: React.PropTypes.object.isRequired,
	};

	static contextTypes = {
		signedIn: React.PropTypes.bool.isRequired,
		router: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this._handleInput = this._handleInput.bind(this);
		this._onSelectQuery = this._onSelectQuery.bind(this);
	}

	stores() { return [UserStore]; }

	reconcileState(state, props) {
		Object.assign(state, props);

		const settings = UserStore.settings.get();
		state.langs = settings && settings.search ? settings.search.languages : null;
	}

	onStateTransition(prevState, nextState) {
		if (this._input && this._input.value) this._goToSearch(this._input.value);
	}

	_goToSearch(query: string) {
		this.context.router.push(locationForSearch(this.props.location, query, true, true));
	}

	_handleInput: Function;
	_handleInput(ev: KeyboardEvent) {
		if (!(ev.currentTarget instanceof HTMLInputElement)) return;
		if (ev.currentTarget.value) {
			this._goToSearch(ev.currentTarget.value);
		}
	}

	_onSelectQuery: OnSelectQueryListener;
	_onSelectQuery(ev: Event, query: string) {
		invariant(this._input, "no input field");

		// Make it feel more realistic.
		const delay = (c: string) => 20 + (25 * Math.random()) + (c === " " ? 75 : 0);
		const simulateTyping = (v: string, i: number = 0, then: Function) => {
			if (i >= v.length) {
				then();
				return;
			}
			const c = v.charAt(i);
			this._input.value += c;
			setTimeout(() => simulateTyping(v, i + 1, then), delay(c));
		};

		this._input.focus();
		this._input.value = "";
		simulateTyping(query, 0, () => setTimeout(() => this._goToSearch(query), 300));
	}

	render() {
		const langSelected = this.state.langs && this.state.langs.length > 0;
		return (
			<div>
				<Helmet title="Home" />
				<div styleName="home-container">
					<Logo type="logotype" styleName="logo" />

					<h2 styleName="description">
						<strong>Instant&nbsp;usage&nbsp;examples&nbsp;and&nbsp;more&nbsp;as&nbsp;you&nbsp;code,</strong> automatically&nbsp;drawn&nbsp;from public&nbsp;and&nbsp;(your&nbsp;own)&nbsp;private&nbsp;code.
					</h2>

					<div styleName="user-actions">
						{!this.context.signedIn && <LocationStateToggleLink href="/login" modalName="login" location={this.props.location}><Button styleName="action-link" type="button" color="blue" outline={true}>Sign in</Button></LocationStateToggleLink>}
						<Link to="/tools/browser"><Button styleName="action-link" type="button" color="blue" outline={true}>Install Chrome extension</Button></Link>
						<Link to="/tools/editor"><Button styleName="action-link" type="button" color="blue" outline={true}>Install editor plugin</Button></Link>
					</div>

					<GlobalSearchInput
						name="q"
						border={true}
						query={this.props.location.query.q || ""}
						autoFocus={true}
						domRef={e => this._input = e}
						styleName="search-input"
						onChange={this._handleInput} />
					<div styleName="search-actions">
						<Button styleName="search-button" type="button" color="blue">Find usage examples</Button>
					</div>

					{<TitledSection title="Top searches" className={styles["top-queries-panel"]}>
						{langSelected && <Queries langs={this.state.langs} onSelectQuery={this._onSelectQuery} />}
						{!langSelected && <p styleName="notice">Select a language below to get started.</p>}
					</TitledSection>}

					{<TitledSection title="Search options" className={styles["search-settings-panel"]}>
						<SearchSettings location={this.props.location} styleName="search-settings" />
					</TitledSection>}
				</div>
			</div>
		);
	}
}

export default CSSModules(DashboardContainer, styles, {allowMultiple: true});

const TitledSection = CSSModules(({
	title,
	children,
	className,
}: {
	title: string,
	children?: React$Element | Array<React$Element>,
	className: string,
}) => (
	<div styleName="titled-section" className={className}>
		<div styleName="section-title">{title}</div>
		{children}
	</div>
), styles);

const topQueries: {[key: LanguageID]: string[]} = {
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
function topQueriesFor(langs: Array<LanguageID>): string[] {
	let qs = [];
	for (let lang of langs) {
		if (topQueries[lang]) {
			qs.push.apply(qs, topQueries[lang]);
		}
	}
	return qs;
}
const Queries = CSSModules(({
	langs,
	onSelectQuery,
}: {
	langs: LanguageID[],
	onSelectQuery: OnSelectQueryListener,
}) => (
	<div>{topQueriesFor(langs).map(q => <Button onClick={(ev) => onSelectQuery(ev, q)} key={q} styleName="query" color="blue" outline={true} size="small">{q}</Button>)}</div>
), styles);
