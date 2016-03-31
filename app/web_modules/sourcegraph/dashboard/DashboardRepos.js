import React from "react";
import update from "react/lib/update";

import Component from "sourcegraph/Component";
import Styles from "./styles/Dashboard.css";
import context from "sourcegraph/context";
import NotificationWell from "sourcegraph/dashboard/NotificationWell";
import RepoList from "sourcegraph/dashboard/RepoList";

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
		this._repoSort = this._repoSort.bind(this);
		this._showCreateUserWell = this._showCreateUserWell.bind(this);
		this._showGitHubLinkWell = this._showGitHubLinkWell.bind(this);
		this._showNoGoRepoWell = this._showNoGoRepoWell.bind(this);
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

	_showCreateUserWell() {
		return !context.currentUser || !context.currentUser.Login;
	}

	_showGitHubLinkWell() {
		return Boolean(context.currentUser) && Boolean(context.currentUser.Login) && this.state.linkGitHub;
	}

	_showNoGoRepoWell() {
		return this.state.repos.filter(this._canMirror).length === 0;
	}

	render() {
		console.log(this._showNoGoRepoWell());
		const filteredRepos = this.state.repos.filter(this._showRepo).sort(this._repoSort);

		return (
			<div className={Styles.repos_list}>
				<nav>
					<NotificationWell initVisible={this._showCreateUserWell() || this._showGitHubLinkWell() || this._showNoGoRepoWell()}>
						{this._showCreateUserWell() &&
							<span>
								We've set up Sourcegraph on some of the top Go repositories for you to check out.
								When you are ready, you can
								<a className={Styles.onboarding_cta_link} href={this.state.signup}>
									set up Sourcegraph for your own repositories!
								</a>
							</span>
						}
						{this._showGitHubLinkWell() &&
							<span>
								Almost there! You'll need to
								<a className={Styles.onboarding_cta_link} href={this.state.linkGitHubURL}>
									link your GitHub account
								</a>
								to use Sourcegraph on your personal repositories.
							</span>
						}
						{this._showNoGoRepoWell() &&
							<span>
								It looks like you do not have any Go repositories. Support for other languages is coming soon!
							</span>
						}
					</NotificationWell>
					<div className={Styles.repos_header}>
						<span className={Styles.repos_label}>Repositories</span>
						{!this.state.linkGitHub &&
							<input
								className={Styles.search_input}
								placeholder="Filter repositories..."
								value={this.state.searchQuery}
								onChange={this._handleSearch}
								type="text" />
						}
					</div>
				</nav>
				<div className={Styles.list_item_group}>
					{filteredRepos.length === 0 &&
						<div className={Styles.repo_section_header}>No Matching Repositories</div>
					}
					<RepoList repos={filteredRepos.filter(this._canMirror)}
						reposDisabled={false} />
					<RepoList repos={filteredRepos.filter(repo => !this._canMirror(repo))}
						reposDisabled={true} />
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
