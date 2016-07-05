// @flow

import React from "react";
import Container from "sourcegraph/Container";
import CSSModules from "react-css-modules";
import styles from "./styles/SearchSettings.css";
import base from "sourcegraph/components/styles/_base.css";
import {Button} from "sourcegraph/components";
import GitHubAuthButton from "sourcegraph/components/GitHubAuthButton";
import Dispatcher from "sourcegraph/Dispatcher";
import UserStore from "sourcegraph/user/UserStore";
import * as UserActions from "sourcegraph/user/UserActions";
import * as RepoActions_typed from "sourcegraph/repo/RepoActions_typed";
import type {Settings} from "sourcegraph/user";
import {allLangs, langName} from "sourcegraph/Language";
import type {LanguageID} from "sourcegraph/Language";
import {privateGitHubOAuthScopes} from "sourcegraph/util/urlTo";
import {withUserContext} from "sourcegraph/app/user";
import LocationStateToggleLink from "sourcegraph/components/LocationStateToggleLink";
import {LocationStateModal, dismissModal} from "sourcegraph/components/Modal";
import InterestForm from "sourcegraph/home/InterestForm";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

class SearchSettings extends Container {
	static propTypes = {
		location: React.PropTypes.object.isRequired,
		repo: React.PropTypes.string,
		className: React.PropTypes.string,
		innerClassName: React.PropTypes.string,
		githubToken: React.PropTypes.object,
	};

	static contextTypes = {
		router: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
	};

	state: {
		settings: Settings;
		betaLanguage: ?LanguageID;
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

	onStateTransition(prevState, nextState) {
		if (prevState.settings !== nextState.settings && nextState.settings && nextState.settings.search && nextState.settings.search.scope) {
			const scope = nextState.settings.search.scope;
			if (scope.public) {
				Dispatcher.Backends.dispatch(new RepoActions_typed.WantRepos("Private=false"));
			}
			if (scope.private) {
				Dispatcher.Backends.dispatch(new RepoActions_typed.WantRepos("Private=true"));
			}
		}
	}

	_langs(): Array<LanguageID> {
		return this.state.settings && this.state.settings.search && this.state.settings.search.languages ? Array.from(this.state.settings.search.languages) : ["golang"];
	}

	_scope() {
		return this.state.settings && this.state.settings.search && this.state.settings.search.scope ? this.state.settings.search.scope : {};
	}

	_hasPrivateGitHubToken() {
		return this.props.githubToken && this.props.githubToken.scope && this.props.githubToken.scope.includes("repo") && this.props.githubToken.scope.includes("read:org") && this.props.githubToken.scope.includes("user:email");
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

		this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_GLOBAL_SEARCH, AnalyticsConstants.ACTION_TOGGLE, "SearchLanguageToggled", {language: lang, enabled: !enabled, languages: langs});
	}

	_setScope(scope: Object) {
		this.context.eventLogger.logEventForCategory(
			AnalyticsConstants.CATEGORY_GLOBAL_SEARCH, AnalyticsConstants.ACTION_CLICK,
			"SearchScopeChanged",
			{
				old: this.state.settings.search && this.state.settings.search.scope,
				update: scope,
			},
		);

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
				<span styleName="label" className={base.pr3}>Languages:</span>
				<div>
					{allLangs.map(lang => (
						lang === "python" || lang === "javascript" ?
							<LocationStateToggleLink key={lang} href="/beta" modalName="beta" location={this.props.location}>
								<Button
									color="default"
									size="small"
									styleName="choice-button"
									onClick={() => this.setState({betaLanguage: lang})}
									outline={true}>
										{langName(lang)}
								</Button>
							</LocationStateToggleLink> :
							<Button
								id={`e2etest-search-lang-select-${lang}`}
								key={lang}
								color={!langs.includes(lang) ? "default" : "blue"}
								size="small"
								styleName="choice-button"
								onClick={() => this._toggleLang(lang)}
								outline={!langs.includes(lang)}>
									{langName(lang)}
							</Button>
						))}
					<LocationStateToggleLink href="/beta" modalName="beta" location={this.props.location}>
						<Button
							color="default"
							size="small"
							styleName="choice-button"
							onClick={() => this.setState({betaLanguage: "more"})}
							outline={true}>
								More...
						</Button>
					</LocationStateToggleLink>
				</div>
			</div>
		);
	}

	_renderScope() {
		const scope = this._scope();
		return (
			<div styleName="row">
				<div styleName="group">
					<span styleName="label" className={base.pr3}>Include:</span>
					<div>
						{this.state.repo && <Button
							color={scope.repo ? "blue" : "default"}
							size="small"
							styleName="choice-button"
							onClick={() => this._setScope({repo: !scope.repo})}
							outline={!scope.repo}>{this.state.repo}</Button>}
						<Button
							color={this.state.githubToken && !scope.popular ? "default" : "blue"}
							size="small"
							styleName="choice-button"
							onClick={() => {
								if (this.props.githubToken) this._setScope({popular: !scope.popular});
							}}
							outline={this.state.githubToken && !scope.popular}>Popular libraries</Button>
						{(!this.state.signedIn || !this.props.githubToken) &&
							<GitHubAuthButton color="green" size="small" outline={true} styleName="choice-button" returnTo={this.props.location}>My public projects</GitHubAuthButton>}
						{this.props.githubToken &&
							<Button
								color={!scope.public ? "default" : "blue"}
								size="small"
								styleName="choice-button"
								onClick={() => this._setScope({public: !scope.public})}
								outline={!scope.public}>My public projects</Button>
						}
						{(!this.state.signedIn || !this._hasPrivateGitHubToken()) &&
							<GitHubAuthButton scopes={privateGitHubOAuthScopes} color="green" size="small" outline={true} styleName="choice-button" returnTo={this.props.location}>My private projects</GitHubAuthButton>}
						{this._hasPrivateGitHubToken() &&
							<Button
								color={!scope.private ? "default" : "blue"}
								size="small"
								styleName="choice-button"
								onClick={() => this._setScope({private: !scope.private})}
								outline={!scope.private}>My private projects</Button>
						}
					</div>
				</div>
			</div>
		);
	}

	render() {
		return (
			<div styleName="groups" className={this.props.className}>
				<div styleName="groups-inner" className={this.props.innerClassName}>
					<div styleName="row">
						{this._renderLanguages()}
					</div>
					{this._renderScope()}
				</div>
				{this.props.location.state && this.props.location.state.modal === "beta" && this.state.betaLanguage &&
					<LocationStateModal modalName="beta" location={this.props.location}>
						<div styleName="modal">
							<h2 styleName="modalTitle">Participate in the Sourcegraph beta for {`${this.state.betaLanguage === "more" ? "your preferred language" : `${langName(this.state.betaLanguage)}`}`}</h2>
							<InterestForm
								rowClass={styles.modalRow}
								language={this.state.betaLanguage === "more" ? null : this.state.betaLanguage}
								onSubmit={dismissModal("beta", this.props.location, this.context.router)} />
						</div>
					</LocationStateModal>
				}
			</div>
		);
	}
}

export default withUserContext(CSSModules(SearchSettings, styles));
