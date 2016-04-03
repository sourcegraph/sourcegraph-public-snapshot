import React from "react";

import Component from "sourcegraph/Component";
import context from "sourcegraph/context";

import {Input, Button} from "sourcegraph/components";

import NotificationWell from "sourcegraph/dashboard/NotificationWell";
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
		this._showGitHubLinkWell = this._showGitHubLinkWell.bind(this);
		this._showNoGoRepoWell = this._showNoGoRepoWell.bind(this);
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
		if (this._filterInput && this._filterInput.getValue() && this._qualifiedName(repo).indexOf(this._filterInput.getValue()) === -1) {
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

	_showCreateUserWell() {
		return !context.currentUser;
	}

	_showGitHubLinkWell() {
		return context.currentUser && !this.state.hasLinkedGitHub;
	}

	_showNoGoRepoWell() {
		return this.state.hasLinkedGitHub && this.state.repos.filter(this._canMirror).length === 0;
	}

	render() {
		const filteredRepos = this.state.repos.filter(this._showRepo).sort(this._repoSort);
		const showExampleRepos = this._showCreateUserWell() || this._showGitHubLinkWell() || this._showNoGoRepoWell();
		const enabledRepos = filteredRepos.filter(this._canMirror).concat(showExampleRepos ? this.state.exampleRepos : []);
		const disabledRepos = filteredRepos.filter(repo => !this._canMirror(repo));

		return (
			<div>
				<NotificationWell visible={showExampleRepos}>
					{this._showCreateUserWell() &&
						[<span key="copy">Want Sourcegraph for your own code?</span>,
						<span key="cta" styleName="onboarding-cta"><Button outline={true} small={true}>
							<a styleName="cta-link" href="/join">
							Sign up
							</a>
						</Button></span>]
					}
					{this._showGitHubLinkWell() &&
						[<span key="copy">Almost there! Link your GitHub account so Sourcegraph can analyze your repositories.</span>,
						<span key="cta" styleName="onboarding-cta"><Button outline={true} small={true}>
							<a styleName="cta-link" href={this.state.linkGitHubURL}>
							Import GitHub repos
							</a>
						</Button></span>]
					}
					{this._showNoGoRepoWell() &&
						<span>It looks like you do not have any Go repositories. Support for other languages is coming soon!</span>
					}
				</NotificationWell>
				<div styleName="header">
					<span styleName="repos-label">{" "}</span>
					{this.state.hasLinkedGitHub &&
						<Input type="text"
							placeholder="Filter repositories..."
							ref={(c) => this._filterInput = c}
							onChange={this._handleSearch} />
					}
				</div>
				<div styleName="list">
					{enabledRepos.length + disabledRepos.length === 0 &&
						<div styleName="list-section-header">No Matching Repositories</div>
					}
					<RepoList repos={enabledRepos} reposDisabled={false} />
					<RepoList repos={disabledRepos} reposDisabled={true} />
				</div>
			</div>
		);
	}
}

DashboardRepos.propTypes = {
	repos: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
	exampleRepos: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
	hasLinkedGitHub: React.PropTypes.bool.isRequired,
	linkGitHubURL: React.PropTypes.string.isRequired,
};

export default CSSModules(DashboardRepos, styles);
