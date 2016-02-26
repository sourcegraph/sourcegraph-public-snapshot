import React from "react";
import update from "react/lib/update";
import classNames from "classnames";

import Component from "sourcegraph/Component";
import Dispatcher from "sourcegraph/Dispatcher";
import moment from "moment";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";

class DashboardRepos extends Component {
	constructor(props) {
		super(props);
		this.state = {
			searchQuery: "",
			filter: null,
		};
		this._handleSearch = this._handleSearch.bind(this);
		this._selectFilter = this._selectFilter.bind(this);
		this._showRepo = this._showRepo.bind(this);
		this._canMirror = this._canMirror.bind(this);
		this._disabledReason = this._disabledReason.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_handleSearch(e) {
		this.setState(update(this.state, {
			searchQuery: {$set: e.target.value},
		}));
	}

	_selectFilter(filterValue) {
		this.setState(update(this.state, {
			filter: {$set: filterValue},
		}));
	}

	_showRepo(repo) {
		const isPrivate = Boolean(repo.Private);
		if (this.state.searchQuery && repo.URI.indexOf(this.state.searchQuery) === -1) {
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
		if (this.state.onWaitlist) {
			if (repo.Private) return false;
		}
		return repo.Language === "Go" || repo.Language === "Java";
	}

	_disabledReason(repo) {
		if (this.state.onWaitlist && repo.Private) return "private repositories coming soon";
		return `${repo.Language || ""} coming soon`;
	}

	render() {
		const toggles = [null, "private", "public"].map((filterValue, i) =>
			<button key={i}
				className={`btn btn-block toggle ${this.state.filter === filterValue ? "btn-primary" : "btn-default"}`}
				onClick={() => this._selectFilter(filterValue)}>
				<span className="toggle-label">{filterValue ? filterValue : "all"}</span>
			</button>
		);

		const repoSort = (a, b) => {
			if (!this._canMirror(a) && this._canMirror(b)) return 1;
			if (this._canMirror(a) && !this._canMirror(b)) return -1;
			if (moment(a.UpdatedAt).isBefore(moment(b.UpdatedAt))) return 1;
			return -1;
		};

		const clickHandler = (repo) => {
			if (repo.ExistsLocally) return _ => window.location.href = `/${repo.URI}`;
			if (!this._canMirror(repo)) return _ => null;
			return _ => {
				Dispatcher.dispatch(new DashboardActions.WantAddMirrorRepo({
					URI: repo.URI,
					Private: Boolean(repo.Private),
				}));
			};
		};

		const repoRowClass = (repo) => classNames("list-group-item", {
			"hover-pointer": this._canMirror(repo) || repo.ExistsLocally,
			"disabled": !repo.ExistsLocally && (this.state.allowGitHubMirrors && !this._canMirror(repo)),
		});

		const emptyStateLabel = this.state.allowGitHubMirrors ? "Link your GitHub account to add repositories." : "No repositories.";

		const filteredRepos = this.state.repos.filter(this._showRepo);

		return (
			<div className="repos-list">
				<nav>
					{this.state.allowGitHubMirrors && <div className="toggles">
						<div className="btn-group">{toggles}</div>
					</div>}
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
					{this.state.repos.length === 0 ? <div className="well">{emptyStateLabel}</div> : <div className="list-group">
						{filteredRepos.length === 0 ? <div className="well">No matching repositories.</div> : filteredRepos.sort(repoSort).map((repo, i) => (
							<div className={repoRowClass(repo)} key={i}
								onClick={clickHandler(repo)}>
								<div className="repo-header">
									<h4>
										<i className={`repo-attr-icon icon-${repo.Private ? "private" : "public"}`}></i>
										{repo.URI}
									</h4>
									{this.state.allowGitHubMirrors && !this._canMirror(repo) &&
										<span className="disabled-reason">{this._disabledReason(repo)}</span>
									}
								</div>
								<div className="repo-body">
									<p className="description">{repo.Description}</p>
									<p className="updated">{`Updated ${moment(repo.UpdatedAt).fromNow()}`}</p>
								</div>
							</div>
						))}
					</div>}
				</div>
			</div>
		);
	}
}

DashboardRepos.propTypes = {
	repos: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
	onWaitlist: React.PropTypes.bool.isRequired,
	allowGitHubMirrors: React.PropTypes.bool.isRequired,
};

export default DashboardRepos;
