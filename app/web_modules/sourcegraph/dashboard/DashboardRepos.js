import React from "react";
import RepoLink from "sourcegraph/components/RepoLink";
import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";
import {Input} from "sourcegraph/components";
import debounce from "lodash/function/debounce";
import GitHubAuthButton from "sourcegraph/user/GitHubAuthButton";

class DashboardRepos extends React.Component {
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

	render() {
		let repos = this.props.repos.filter(this._showRepo).sort(this._repoSort);

		return (
			<div styleName="list">
				{this.context.signedIn &&
					<div styleName="header">
						<div styleName="list-section-header">Repositories</div>
						<div styleName="filter"><Input type="text"
							placeholder="Filter repositories..."
							domRef={(e) => this._filterInput = e}
							onChange={this._handleFilter} />
						</div>
					</div>}
				{this._hasGithubToken() && repos.length === 0 && (!this._filterInput || !this._filterInput.value) && <div styleName="list-item-loading">Loading...</div>}
				{this._hasGithubToken() && this._filterInput && this._filterInput.value && repos.length === 0 && <div styleName="list-item-loading">No matching repositories</div>}
				{!this._hasGithubToken() && <div styleName="list-item-loading">Link your GitHub account to see your repositories here.
					<div styleName="cta" style={{justifyContent: "flex-start"}}>
						<GitHubAuthButton>Link GitHub account</GitHubAuthButton>
					</div>
				</div>}
				{repos.length > 0 && repos.map((repo, i) =>
					<div key={i}>
						<div styleName="list-item">
							<div styleName="uri-container">
								<div styleName="uri">
									<RepoLink repo={repo.URI || `github.com/${repo.Owner}/${repo.Name}`} />
								</div>
							</div>

							{repo.Description && <div>
								<p styleName="description">{repo.Description}</p>
							</div>}
						</div>
					</div>
				)}
			</div>
		);
	}
}

DashboardRepos.propTypes = {
	repos: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
};

export default CSSModules(DashboardRepos, styles);
