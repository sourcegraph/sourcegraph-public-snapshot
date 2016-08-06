// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";
import * as styles from "./styles/Repos.css";
import * as base from "sourcegraph/components/styles/_base.css";
import {Input, Heading, Button, ToggleSwitch} from "sourcegraph/components/index";
import RepoLink from "sourcegraph/components/RepoLink";
import Dispatcher from "sourcegraph/Dispatcher";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import debounce from "lodash.debounce";
import GitHubAuthButton from "sourcegraph/components/GitHubAuthButton";
import {privateGitHubOAuthScopes} from "sourcegraph/util/urlTo";

class Repos extends React.Component<any, any> {
	_filterInput: any;
	
	static propTypes = {
		repos: React.PropTypes.arrayOf(React.PropTypes.object),
		location: React.PropTypes.object.isRequired,
	};

	static contextTypes = {
		signedIn: React.PropTypes.bool.isRequired,
		githubToken: React.PropTypes.object,
		eventLogger: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this._filterInput = null;
		this._handleFilter = this._handleFilter.bind(this);
		this._handleFilter = debounce(this._handleFilter, 25);
		this._showRepo = this._showRepo.bind(this);
	}

	// _repoSort is a comparison function that sorts more recently
	// pushed repos first.
	_repoSort(a, b) {
		if (a.PushedAt < b.PushedAt) return 1;
		else if (a.PushedAt > b.PushedAt) return -1;
		return 0;
	}

	_handleFilter() {
		this.forceUpdate();
	}

	_showRepo(repo) {
		if (this._filterInput && this._filterInput.value &&
			this._qualifiedName(repo).indexOf(this._filterInput.value.trim().toLowerCase()) === -1) {
			return false;
		}

		return true; // no filter; return all
	}

	_qualifiedName(repo) {
		return (`${repo.Owner}/${repo.Name}`).toLowerCase();
	}

	_hasGithubToken() {
		return this.context && (this.context as any).githubToken;
	}

	_hasPrivateGitHubToken() {
		return (this.context as any).githubToken && (this.context as any).githubToken.scope && (this.context as any).githubToken.scope.includes("repo") && (this.context as any).githubToken.scope.includes("read:org");
	}

	_toggleRepo(remoteRepo: any) {
		Dispatcher.Backends.dispatch(new RepoActions.WantCreateRepo(remoteRepo.URI, remoteRepo, true));
	}

	render(): JSX.Element | null {
		let repos = (this.props.repos || []).filter(this._showRepo).sort(this._repoSort);

		return (
			<div className={base.pb6}>
				<header className={styles.header}>
					<Heading level="7" color="cool_mid_gray">Your repositories</Heading>
					<p>To get jump-to-definition, search, and code examples, enable indexing on your repositories using the toggle. Private code indexed on Sourcegraph is only available to you and those with permissions to the underlying GitHub repository.</p>
					<div className={styles.input_bar}>
						{!this._hasGithubToken() && <GitHubAuthButton returnTo={this.props.location} className={styles.github_button}>Add public repositories</GitHubAuthButton>}
						{!this._hasPrivateGitHubToken() && <GitHubAuthButton scopes={privateGitHubOAuthScopes} returnTo={this.props.location} className={styles.github_button}>Add private repositories</GitHubAuthButton>}
					</div>
				</header>
				<div className={styles.settings}>
					{this._hasGithubToken() &&
					<div className={styles.list_heading}>
						<Input type="text"
							placeholder="Find a repository..."
							domRef={(e) => this._filterInput = e}
							spellCheck={false}
							className={styles.filter_input}
							onChange={this._handleFilter} />
						<span className={styles.list_label}>Enable Indexing</span>
					</div>}
					<div className={styles.repos_list}>
						{repos.length > 0 && repos.map((repo, i) =>
							<div className={styles.row} key={i}>
								<div className={styles.info}>
									{repo.ID ?
										<RepoLink repo={repo.URI || `github.com/${repo.Owner}/${repo.Name}`} /> :
										(repo.URI && repo.URI.replace("github.com/", "").replace("/", " / ", 1)) || `${repo.Owner} / ${repo.Name}`
									}
									{repo.Description && <p className={styles.description}>
										{repo.Description.length > 100 ? `${repo.Description.substring(0, 100)}...` : repo.Description}
									</p>}
								</div>
								<div className={styles.toggle}>
									<ToggleSwitch defaultChecked={Boolean(repo.ID)} onChange={(checked) => {
										this._toggleRepo(repo);
									}}/>
								</div>
							</div>
						)}
					</div>
					{this._hasGithubToken() && repos.length === 0 && (!this._filterInput || !this._filterInput.value) &&
						<p className={styles.indicator}>Loading...</p>
					}

					{this._hasGithubToken() && this._filterInput && this._filterInput.value && repos.length === 0 &&
						<p className={styles.indicator}>No matching repositories</p>
					}
				</div>
				{this.props.location.query.onboarding &&
					<footer className={styles.footer}>
						<a className={styles.footer_link} href="/integrations?onboarding=t">
							<Button color="green" className={styles.footer_btn}>Continue</Button>
						</a>
					</footer>
				}
			</div>
		);
	}
}

export default CSSModules(Repos, styles, {allowMultiple: true});
