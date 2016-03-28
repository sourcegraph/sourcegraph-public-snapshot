import React from "react";
import update from "react/lib/update";

import Component from "sourcegraph/Component";
import moment from "moment";
import repoLink from "sourcegraph/util/repoLink";
import Styles from "./styles/Dashboard.css";
import context from "sourcegraph/context";
import NotificationWell from "sourcegraph/dashboard/NotificationWell";

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
		const repoDisabled = (repo) => !repo.URI && !this._canMirror(repo);

		const repoRowClass = (repo, index, size) => {
			if (!repoDisabled(repo)) {
				if (index === (size-1)) {
					return Styles.list_item_bottom;
				}
				return Styles.list_item;
			}
			if (index === (size-1)) {
				return Styles.list_item_disabled_bottom;
			}
			return Styles.list_item_disabled;
		};

		const filteredRepos = this.state.repos.filter(this._showRepo);
		const showCreateUserWell = !context.currentUser || !context.currentUser.Login;
		const showGitHubLinkWell = Boolean(context.currentUser) && Boolean(context.currentUser.Login) && this.state.linkGitHub;
		let seenInvalid = false;
		// <i className={`sg-icon repo-attr-icon sg-icon-${repo.Private ? "private" : "public"}`}></i>
		return (
			<div className="repos-list">
				<nav>{!this.state.linkGitHub &&
					<div className="search-bar">
						<div className="input-group">
							<input className="form-control search-input"
								placeholder="Filter repositories..."
								value={this.state.searchQuery}
								onChange={this._handleSearch}
								type="text" />
							<span className="input-group-addon search-addon"><i className="fa fa-search search-icon"></i></span>
						</div>
					</div>}
					<NotificationWell initVisible={showCreateUserWell}>
						<span>We've set up Sourcegraph on some of the top Go repositories for you to check out. When you are ready, you can</span>
					<a className={Styles.wrap_link} href={this.state.signup}>set up Sourcegraph for your own repositories!</a>
					</NotificationWell>
					<NotificationWell initVisible={showGitHubLinkWell}>
						<span>Almost there! You'll need to </span>
						<a href={this.state.linkGitHubURL}>link your GitHub account</a>
						<span> to use Sourcegraph on your personal repositories</span>
					</NotificationWell>
					<span className={Styles.section_header}>{"Repositories"}</span>
				</nav>
				<div>
					{this.state.repos.length > 0 &&
						<div className="list-group">
							<div className={Styles.repo_list_header}>Go Repositories</div>
							{filteredRepos.length === 0 ? <div className="well">No matching repositories.</div> : filteredRepos.sort(this._repoSort).map((repo, i) => (
								<div key={i}>
									{!this._canMirror(repo) && seenInvalid++ === 0 &&
										<div className={Styles.repo_list_header_temp}>Coming Soon</div>
									}
									<div className={repoRowClass(repo, i, this.state.repos.length)} key={i}>
										<div className="repo-header">
											<span className={Styles.repo_title}>
												<span className={Styles.repo_title}>{repoLink(repo.URI || `github.com/${repo.Owner}/${repo.Name}`, repoDisabled(repo))}</span>
											</span>
											{!this._canMirror(repo) &&
												<span className={Styles.repo_disable_reason}>{this._disabledReason(repo)}</span>
											}
										</div>
										<div className="repo-body">
											<p className={Styles.repo_description}>{repo.Description}</p>
											<p className={Styles.repo_updated}>{`Updated ${moment(this._repoTime(repo)).fromNow()}`}</p>
										</div>
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
	linkGitHubURL: React.PropTypes.string.isRequired,
	signup: React.PropTypes.string.isRequired,
};

export default DashboardRepos;
