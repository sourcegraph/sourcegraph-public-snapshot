// @flow

import React from "react";
import Container from "sourcegraph/Container";
import CSSModules from "react-css-modules";
import styles from "./styles/SearchSettings.css";
import {Button} from "sourcegraph/components";
import GitHubAuthButton from "sourcegraph/components/GitHubAuthButton";
import Dispatcher from "sourcegraph/Dispatcher";
import UserStore from "sourcegraph/user/UserStore";
import * as UserActions from "sourcegraph/user/UserActions";
import type {Settings} from "sourcegraph/user";
import {allLangs, langName} from "sourcegraph/Language";
import type {LanguageID} from "sourcegraph/Language";

class SearchSettings extends Container {
	static propTypes = {
		className: React.PropTypes.string,
		showAlerts: React.PropTypes.bool.isRequired,
	};

	state: {
		settings: Settings;
	};

	stores() { return [UserStore]; }

	reconcileState(state, props) {
		Object.assign(state, props);

		state.settings = UserStore.settings.get();

		// Use this instead of context signedIn because of the issues surrounding
		// propagating context through components that use shouldComponentUpdate.
		// We're already observing UserStore, so this doesn't add any extra overhead.
		state.signedIn = Boolean(UserStore.activeAuthInfo());
	}

	_langs() {
		return this.state.settings && this.state.settings.search && this.state.settings.search.languages ? Array.from(this.state.settings.search.languages) : [];
	}

	_scope() {
		return this.state.settings && this.state.settings.search && this.state.settings.search.scope ? this.state.settings.search.scope : {popular: false, public: false, private: false, starred: false, team: false};
	}

	_toggleLang(lang: LanguageID) {
		const langs = this._langs();
		const enabled = langs.includes(lang);

		if (enabled) langs.splice(langs.indexOf(lang), 1);
		else {
			langs.push(lang);
			langs.sort();
		}

		const newSettings = {
			...this.state.settings,
			search: {
				...this.state.settings.search,
				languages: langs,
			},
		};

		Dispatcher.Stores.dispatch(new UserActions.UpdateSettings(newSettings));
	}

	_setScope(scope: any) {
		const newSettings = {
			...this.state.settings,
			search: {
				...this.state.settings.search,
				scope: {
					...(this.state.settings.search && this.state.settings.search.scope),
					...scope,
				},
			},
		};

		Dispatcher.Stores.dispatch(new UserActions.UpdateSettings(newSettings));
	}

	_renderLanguages() {
		const langs = this._langs();
		return (
			<div styleName="group">
				<span styleName="label">Languages:</span>
				<div>
					{allLangs.map(lang => (
						<Button
							key={lang}
							color="default"
							size="small"
							styleName="choice-button"
							onClick={() => this._toggleLang(lang)}
							outline={!langs.includes(lang)}>{langName(lang)}</Button>
					))}
				</div>
				{!this.state.signedIn && <GitHubAuthButton color="green" size="small" outline={true} styleName="choice-button">Auto-detect your languages</GitHubAuthButton>}
			</div>
		);
	}

	_renderScope() {
		const scope = this._scope();
		return (
			<div styleName="row">
				<div styleName="group">
					<span styleName="label">Include:</span>
					<div>
						<Button
							color="default"
							size="small"
							styleName="choice-button"
							onClick={() => this._setScope({popular: !scope.popular})}
							outline={!scope.popular}>Popular libraries</Button>
						{!this.state.signedIn && <GitHubAuthButton color="green" size="small" outline={true} styleName="choice-button">Libraries you use</GitHubAuthButton>}
						{!this.state.signedIn && <GitHubAuthButton color="green" size="small" outline={true} styleName="choice-button">Your public projects</GitHubAuthButton>}
						{!this.state.signedIn && <GitHubAuthButton color="green" size="small" outline={true} styleName="choice-button">Private</GitHubAuthButton>}
						{!this.state.signedIn && <GitHubAuthButton color="green" size="small" outline={true} styleName="choice-button">Team</GitHubAuthButton>}
						{!this.state.signedIn && <GitHubAuthButton color="green" size="small" outline={true} styleName="choice-button">Starred</GitHubAuthButton>}
					</div>
				</div>
			</div>
		);
	}

	render() {
		const langChosen = this.state.settings && this.state.settings.search && this.state.settings.search.languages && this.state.settings.search.languages.length > 0;

		return (
			<div styleName="groups"><div styleName="groups-inner" className={this.props.className}>
				<div styleName="row">
					{this._renderLanguages()}
				</div>
				{!(this.state.signedIn || langChosen) && this.state.showAlerts && <div styleName="row">
					<div styleName="group">
						<Alert>Select a language to search.</Alert>
					</div>
				</div>}
				{this._renderScope()}
			</div></div>
		);
	}
}
export default CSSModules(SearchSettings, styles);

const Alert = CSSModules(({children}: {children: React$Element | Array<React$Element>}) => (
	<span styleName="alert">
		{children}
	</span>
), styles);

