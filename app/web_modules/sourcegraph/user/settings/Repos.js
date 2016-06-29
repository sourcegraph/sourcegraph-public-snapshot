import React from "react";
import RepoLink from "sourcegraph/components/RepoLink";
import CSSModules from "react-css-modules";
import styles from "./styles/Repos.css";
import base from "sourcegraph/components/styles/_base.css";
import {Input, Table} from "sourcegraph/components";
import debounce from "lodash/function/debounce";
import GitHubAuthButton from "sourcegraph/components/GitHubAuthButton";
import {privateGitHubOAuthScopes} from "sourcegraph/util/urlTo";

class Repos extends React.Component {
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
		return this.context && this.context.githubToken;
	}

	_hasPrivateGitHubToken() {
		return this.context.githubToken && (!this.context.githubToken.scope || !(this.context.githubToken.scope.includes("repo") && this.context.githubToken.scope.includes("read:org") && this.context.githubToken.scope.includes("user:email")));
	}

	render() {
		let repos = (this.props.repos || []).filter(this._showRepo).sort(this._repoSort);

		return (
			<div className={base.pb4}>
				<div>
					<Table styleName="repos">
						<thead>
							<tr>
								<td>
									<Input type="text"
										placeholder="Find a repository..."
										domRef={(e) => this._filterInput = e}
										spellCheck={false}
										styleName="filter-input"
										onChange={this._handleFilter} />
								</td>
								<td>
									{this._hasGithubToken && !this._hasPrivateGitHubToken && <GitHubAuthButton scopes={privateGitHubOAuthScopes} returnTo={this.props.location} styleName="github-button">Connect your private repositories</GitHubAuthButton>}
									{!this._hasGithubToken() && <GitHubAuthButton returnTo={this.props.location} styleName="github-button">Connect with GitHub</GitHubAuthButton>}
								</td>
							</tr>
						</thead>
						<tbody>
							{repos.length > 0 && repos.map((repo, i) =>
								<tr styleName="row" key={i}>
									<td styleName="cell" colSpan="2">
										<RepoLink styleName="repo-link" repo={repo.URI || `github.com/${repo.Owner}/${repo.Name}`} />
										{repo.Description && <p styleName="description">{repo.Description}</p>}
									</td>
								</tr>
							)}
						</tbody>
					</Table>
					{this._hasGithubToken() && repos.length === 0 && (!this._filterInput || !this._filterInput.value) &&
						<p styleName="indicator">Loading...</p>
					}

					{this._hasGithubToken() && this._filterInput && this._filterInput.value && repos.length === 0 &&
						<p styleName="indicator">No matching repositories</p>
					}
				</div>
			</div>
		);
	}
}

export default CSSModules(Repos, styles, {allowMultiple: true});
