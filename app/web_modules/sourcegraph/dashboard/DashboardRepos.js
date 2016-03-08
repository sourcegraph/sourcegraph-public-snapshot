import React from "react";
import update from "react/lib/update";
import classNames from "classnames";

import Component from "sourcegraph/Component";
import moment from "moment";
import repoLink from "sourcegraph/util/repoLink";

class DashboardRepos extends Component {
	constructor(props) {
		super(props);
		this.state = {
			searchQuery: "",
			filter: null,
		};
		this._handleSearch = this._handleSearch.bind(this);
		this._selectFilter = this._selectFilter.bind(this);
		this._qualifiedName = this._qualifiedName.bind(this);
		this._showRepo = this._showRepo.bind(this);
		this._canMirror = this._canMirror.bind(this);
		this._disabledReason = this._disabledReason.bind(this);
		this._repoSort = this._repoSort.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_handleSearch(e) {
		if (typeof e.target === "undefined") return; // TODO(autotest): support DOM events
		this.setState(update(this.state, {
			searchQuery: {$set: e.target.value},
		}));
	}

	_selectFilter(filterValue) {
		this.setState(update(this.state, {
			filter: {$set: filterValue},
		}));
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
		const isPrivate = Boolean(repo.Private);
		if (this.state.searchQuery && this._qualifiedName(repo).indexOf(this.state.searchQuery) === -1) {
			return false;
		}
		if (this.state.filter) {
			if (this.state.filter === "private" && !isPrivate) {
				return false;
			}
			if (this.state.filter === "public" && isPrivate) {
				return false;
			}
		}
		return true; // no filter; return all
	}

	_canMirror(repo) {
		return !repo.GitHubID || repo.Language === "Go";
	}

	_disabledReason(repo) {
		return `${repo.Language || ""} coming soon`;
	}

	_repoTime(repo) {
		return repo.UpdatedAt || repo.PushedAt || repo.CreatedAt;
	}

	_repoSort(a, b) {
		if (!this._canMirror(a) && this._canMirror(b)) return 1;
		if (this._canMirror(a) && !this._canMirror(b)) return -1;
		if (moment(this._repoTime(a)).isBefore(moment(this._repoTime(b)))) return 1;
		if (moment(this._repoTime(a)).isAfter(moment(this._repoTime(b)))) return -1;
		return -1;
	}

	render() {
		const toggles = [null, "private", "public"].map((filterValue, i) =>
			<button key={i}
				className={`btn btn-block toggle ${this.state.filter === filterValue ? "btn-primary" : "btn-default"}`}
				onClick={() => this._selectFilter(filterValue)}>
				<span className="toggle-label">{filterValue ? filterValue : "all"}</span>
			</button>
		);

		const repoDisabled = (repo) => !repo.URI && !this._canMirror(repo);

		const repoRowClass = (repo) => classNames("list-group-item", {
			"repo-disabled": repoDisabled(repo),
		});

		const filteredRepos = this.state.repos.filter(this._showRepo);

		return (
			<div className="repos-list">
				<nav>
					<div className="toggles">
						<div className="btn-group">{toggles}</div>
					</div>
					<div className="search-bar">
						<div className="input-group">
							<input className="form-control search-input"
								placeholder="Find a repository..."
								value={this.state.searchQuery}
								onChange={this._handleSearch}
								type="text" />
							<span className="input-group-addon search-addon"><i className="fa fa-search search-icon"></i></span>
						</div>
					</div>
				</nav>
				<div className="repos">
					{this.state.repos.length === 0 ?
						<div className="well">{"Link your GitHub account to add repositories."}</div> :
						<div className="list-group">
							{filteredRepos.length === 0 ? <div className="well">No matching repositories.</div> : filteredRepos.sort(this._repoSort).map((repo, i) => (
								<div className={repoRowClass(repo)} key={i}>
									<div className="repo-header">
										<h4>
											<i className={`sg-icon repo-attr-icon sg-icon-${repo.Private ? "private" : "public"}`}></i>
											{repoLink(repo.URI || `github.com/${repo.Owner}/${repo.Name}`, repoDisabled(repo))}
										</h4>
										{!this._canMirror(repo) &&
											<span className="disabled-reason">{this._disabledReason(repo)}</span>
										}
									</div>
									<div className="repo-body">
										<p className="description">{repo.Description}</p>
										<p className="updated">{`Updated ${moment(this._repoTime(repo)).fromNow()}`}</p>
									</div>
								</div>
							))}
						</div>
					}
				</div>
			</div>
		);
	}
}

DashboardRepos.propTypes = {
	repos: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
	linkGitHub: React.PropTypes.bool.isRequired,
};

export default DashboardRepos;
