import * as debounce from "lodash/debounce";
import * as React from "react";
import {Repo} from "sourcegraph/api/index";
import {context} from "sourcegraph/app/context";
import {Button, Heading, Input} from "sourcegraph/components";
import {GitHubAuthButton} from "sourcegraph/components/GitHubAuthButton";
import {GoogleAuthButton} from "sourcegraph/components/GoogleAuthButton";
import {RepoLink} from "sourcegraph/components/RepoLink";
import * as base from "sourcegraph/components/styles/_base.css";
import {whitespace} from "sourcegraph/components/utils/whitespace";
import {Location} from "sourcegraph/Location";
import * as styles from "sourcegraph/user/settings/styles/Repos.css";
import {privateGitHubOAuthScopes, privateGoogleOAuthScopes} from "sourcegraph/util/urlTo";

interface Props {
	repos: Repo[] | null;
	location?: Location;
}

type State = any;

export class Repos extends React.Component<Props, State> {
	_filterInput: any;

	constructor(props: Props) {
		super(props);
		this._filterInput = null;
		this._handleFilter = this._handleFilter.bind(this);
		this._handleFilter = debounce(this._handleFilter, 25);
		this._showRepo = this._showRepo.bind(this);
	}

	// _repoSort is a comparison function that sorts more recently
	// pushed repos first.
	_repoSort(a: Repo, b: Repo): number {
		if (a.PushedAt < b.PushedAt) {
			return 1;
		}
		if (a.PushedAt > b.PushedAt) {
			return -1;
		}
		return 0;
	}

	_handleFilter(): void {
		this.forceUpdate();
	}

	_showRepo(repo: Repo): boolean {
		if (this._filterInput && this._filterInput.value &&
			this._qualifiedName(repo).indexOf(this._filterInput.value.trim().toLowerCase()) === -1) {
			return false;
		}

		return true; // no filter; return all
	}

	_qualifiedName(repo: Repo): string {
		return (`${repo.Owner}/${repo.Name}`).toLowerCase();
	}

	_header(): JSX.Element {
		return (
			<header className={styles.header}>
				<Heading level={7} color="gray">Your repositories</Heading>
				{!context.hasPrivateGitHubToken() && <p>Private code indexed on Sourcegraph is only available to you and those with permissions to the underlying GitHub repository.</p>}
				<div className={styles.input_bar}>
					{!context.hasPrivateGitHubToken() && <GitHubAuthButton scopes={privateGitHubOAuthScopes} returnTo={this.props.location} className={styles.github_button}>Add private repositories</GitHubAuthButton>}
					{window.localStorage["google"] === "true" && !context.hasPrivateGoogleToken() && <GoogleAuthButton scopes={privateGoogleOAuthScopes} returnTo={this.props.location} className={styles.google_button}>Add GCP repositories</GoogleAuthButton>}
				</div>
			</header>);
	}

	_footer(): JSX.Element {
		return (<div>
			{this.props.location && this.props.location.query["onboarding"] &&
				<footer className={styles.footer}>
					<a className={styles.footer_link} href="/integrations?onboarding=t">
						<Button color="green" className={styles.footer_btn}>Continue</Button>
					</a>
				</footer>
			}
		</div>);
	}

	_repoList(repos: Repo[]): JSX.Element {
		return (
			<div style={{marginLeft: whitespace[4]}}>
				<div style={{marginTop: whitespace[4], marginBottom: whitespace[3]}}>
				{context.gitHubToken && repos.length === 0 ?
					<p className={styles.indicator}>Looks like you have no repositories.</p> :
					<div>
						<Input type="text"
							placeholder="Find a repository..."
							domRef={(e) => this._filterInput = e}
							spellCheck={false}
							onChange={this._handleFilter} />
					</div>}
				</div>
			<div className={styles.repos_list}>
				{repos.length > 0 && repos.map((repo, i) =>
					<div className={styles.row} key={i}>
						<div className={styles.info}>
							<RepoLink repo={repo.URI || `github.com/${repo.Owner}/${repo.Name}`} />
							{repo.Description && <p className={styles.description}>
								{repo.Description.length > 100 ? `${repo.Description.substring(0, 100)}...` : repo.Description}
							</p>}
						</div>
					</div>
				)}
			</div>
			{context.gitHubToken && this._filterInput && this._filterInput.value && repos.length === 0 &&
				<p className={styles.indicator}>No matching repositories</p>
			}
		</div>);
	}

	render(): JSX.Element | null {
		let filteredRepos;
		if (this.props.repos) {
			filteredRepos = this.props.repos.filter(this._showRepo).sort(this._repoSort);
		}

		return (
			<div className={base.pb6}>
				{this._header()}
				{!this.props.repos ?
					<p style={{marginTop: whitespace[4], marginBottom: whitespace[4], marginLeft: whitespace[4]}}>
						Loading...
					</p> : this._repoList(filteredRepos)}
				{this._footer()}
			</div>
		);
	}
}
