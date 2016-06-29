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
import * as RepoActions from "sourcegraph/repo/RepoActions";
import type {Settings} from "sourcegraph/user";
import {allLangs, langName} from "sourcegraph/Language";
import type {LanguageID} from "sourcegraph/Language";
import {privateGitHubOAuthScopes} from "sourcegraph/util/urlTo";
import {withUserContext} from "sourcegraph/app/user";
import LocationStateToggleLink from "sourcegraph/components/LocationStateToggleLink";
import {LocationStateModal, dismissModal} from "sourcegraph/components/Modal";
import InterestForm from "sourcegraph/home/InterestForm";

class SearchSettings extends Container {
	static propTypes = {
		location: React.PropTypes.object.isRequired,
		repo: React.PropTypes.string,
		className: React.PropTypes.string,
		innerClassName: React.PropTypes.string,
		showAlerts: React.PropTypes.bool.isRequired,
		githubToken: React.PropTypes.object,
	};

	static contextTypes = {
		router: React.PropTypes.object.isRequired,
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
				Dispatcher.Backends.dispatch(new RepoActions.WantRemoteRepos({deps: true, private: false}));
			} else if (scope.private) {
				Dispatcher.Backends.dispatch(new RepoActions.WantRemoteRepos({deps: true, private: true}));
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
	}

	_setScope(scope: any, currState: any) {
		if (!currState) currState = this.state;

		const newSettings = {
			...currState.settings,
			search: {
				...currState.settings.search,
				scope: {
					...(currState.settings.search && currState.settings.search.scope),
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
								key={lang}
								color={!langs.includes(lang) ? "default" : "blue"}
								size="small"
								styleName="choice-button"
								onClick={() => this._toggleLang(lang)}
								outline={!langs.includes(lang)}>
									{langName(lang)}
							</Button>
						))}
				</div>
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
						{this.state.repo && <Button
							color={!scope.repo ? "blue" : "default"}
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
							<GitHubAuthButton color="green" size="small" outline={true} styleName="choice-button" returnTo={this.props.location}>Your public projects + deps</GitHubAuthButton>}
						{this.props.githubToken &&
							<Button
								color={!scope.public ? "default" : "blue"}
								size="small"
								styleName="choice-button"
								onClick={() => this._setScope({public: !scope.public})}
								outline={!scope.public}>Your public projects + deps</Button>
						}
						{(!this.state.signedIn || !this._hasPrivateGitHubToken()) &&
							<GitHubAuthButton scopes={privateGitHubOAuthScopes} color="green" size="small" outline={true} styleName="choice-button" returnTo={this.props.location}>Your private projects + deps</GitHubAuthButton>}
						{this._hasPrivateGitHubToken() &&
							<Button
								color={!scope.private ? "default" : "blue"}
								size="small"
								styleName="choice-button"
								onClick={() => this._setScope({private: !scope.private})}
								outline={!scope.private}>Your private projects + deps</Button>
						}
					</div>
				</div>
			</div>
		);
	}

	render() {
		const langChosen = this.state.settings && this.state.settings.search && this.state.settings.search.languages && this.state.settings.search.languages.length > 0;
		const scope = this._scope();

		return (
			<div styleName="groups" className={this.props.className}>
				<div styleName="groups-inner" className={this.props.innerClassName}>
					<div styleName="row">
						{this._renderLanguages()}
					</div>
					{!langChosen && this.state.showAlerts && <div styleName="row">
						<div styleName="group">
							<Alert>Select a language to search.</Alert>
						</div>
					</div>}
					{this._renderScope()}
					{this.state.signedIn && this.state.showAlerts && !scope.public && !scope.private && (!scope.repo || !this.state.repo) && !scope.popular &&
						<div styleName="row">
							<div styleName="group">
								<Alert>Select repositories to include.</Alert>
							</div>
						</div>
					}
				</div>
				{this.props.location.state && this.props.location.state.modal === "beta" && this.state.betaLanguage &&
					<LocationStateModal modalName="beta" location={this.props.location}>
						<div styleName="modal">
							<h2 styleName="modalTitle">Join the Sourcegraph beta list</h2>
							<h3 styleName="modalTitle">We don't support {langName(this.state.betaLanguage)} yet, but will soon</h3>
							<InterestForm
								rowClass={styles.modalRow}
								language={this.state.betaLanguage}
								onSubmit={dismissModal("beta", this.props.location, this.context.router)} />
						</div>
					</LocationStateModal>
				}
			</div>
		);
	}
}
export default withUserContext(CSSModules(SearchSettings, styles));

const Alert = CSSModules(({children}: {children: React$Element | Array<React$Element>}) => (
	<span styleName="alert">
		{children}
	</span>
), styles);

