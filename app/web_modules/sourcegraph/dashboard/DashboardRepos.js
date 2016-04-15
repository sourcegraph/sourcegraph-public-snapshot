import React from "react";
import Component from "sourcegraph/Component";
import RepoList from "sourcegraph/dashboard/RepoList";
import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";

class DashboardRepos extends Component {
	constructor(props) {
		super(props);
		this._filterInput = null;
		this._handleSearch = this._handleSearch.bind(this);
		this._showRepo = this._showRepo.bind(this);
		this._canMirror = this._canMirror.bind(this);
		this._repoSort = this._repoSort.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_handleSearch() {
		this.forceUpdate();
	}

	_qualifiedName(repo) {
		if (repo.URI) {
			let collection = [],
				parts = repo.URI.split("/");
			parts.forEach(part => {
				if (part !== "sourcegraph.com" && part !== "github.com") collection.push(part);
			});

			return collection.join("/");
		}

		return `${repo.Owner}/${repo.Name}`;
	}

	_showRepo(repo) {
		if (this._filterInput && this._filterInput.value &&
			this._qualifiedName(repo).indexOf(this._filterInput.value.trim()) === -1) {
			return false;
		}

		return true; // no filter; return all
	}

	_canMirror(repo) {
		return !Boolean(repo.GitHubID) || repo.Language === "Go";
	}

	_repoTime(repo) {
		return repo.UpdatedAt || repo.PushedAt || repo.CreatedAt;
	}

	_repoSort(a, b) {
		if (!this._canMirror(a) && this._canMirror(b)) return 1;
		if (this._canMirror(a) && !this._canMirror(b)) return -1;
		let ta = this._repoTime(a);
		let tb = this._repoTime(b);
		if (ta < tb) return 1;
		else if (ta === tb) return 0;
		return -1;
	}

	render() {
		const filteredRepos = this.state.repos.filter(this._showRepo).sort(this._repoSort);
		const filteredExampleRepos = this.state.exampleRepos.filter(this._showRepo);
		const enabledRepos = filteredExampleRepos.length ? filteredExampleRepos.concat(filteredRepos.filter(this._canMirror)) : filteredRepos.filter(this._canMirror);
		const disabledRepos = filteredRepos.filter(repo => !this._canMirror(repo));

		return (
			<div>
				<div styleName="list">
					{enabledRepos.length + disabledRepos.length === 0 &&
						<div styleName="list-section-header">No Matching Repositories</div>}
					<RepoList repos={enabledRepos} reposDisabled={false}/>
					<RepoList repos={disabledRepos} reposDisabled={true}/>
				</div>
			</div>
		);
	}
}

DashboardRepos.propTypes = {
	repos: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
	exampleRepos: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
};

export default CSSModules(DashboardRepos, styles);
