import React from "react";

import debounce from "lodash/function/debounce";

import Component from "sourcegraph/Component";
import context from "sourcegraph/context";

import {Input, Link} from "sourcegraph/components";

import NotificationWell from "sourcegraph/dashboard/NotificationWell";
import RepoList from "sourcegraph/dashboard/RepoList";

import {urlToGitHubOAuth} from "sourcegraph/util/urlTo";

import EventLogger from "sourcegraph/util/EventLogger";

import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";

class DashboardRepos extends Component {
	constructor(props) {
		super(props);
		this._filterInput = null;
		this._handleSearch = this._handleSearch.bind(this);
		this._handleSearch = debounce(this._handleSearch, 25);
		this._showRepo = this._showRepo.bind(this);
		this._canMirror = this._canMirror.bind(this);
		this._repoSort = this._repoSort.bind(this);
		this._showGitHubLinkWell = this._showGitHubLinkWell.bind(this);
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
		return context.currentUser && this.state.hasLinkedGitHub !== null && !this.state.hasLinkedGitHub;
	}

	render() {
		const filteredRepos = this.state.repos.filter(this._showRepo).sort(this._repoSort);
		const filteredExampleRepos = this.state.exampleRepos.filter(this._showRepo);
		const showExampleRepos = this._showCreateUserWell() || this._showGitHubLinkWell() || this.state.repos.length === 0;
		const enabledRepos = filteredRepos.filter(this._canMirror).concat(showExampleRepos ? filteredExampleRepos : []);
		const disabledRepos = filteredRepos.filter(repo => !this._canMirror(repo));

		return (
			<div>
				<NotificationWell visible={this._showCreateUserWell() || this._showGitHubLinkWell()}>
					{this._showCreateUserWell() &&
						[<span key="copy">Want Sourcegraph for your own code?</span>,
						<span key="cta" styleName="onboarding-cta">
							<Link styl="button" to="/join">
							Sign up
							</Link>
						</span>]
					}
					{this._showGitHubLinkWell() &&
						[<span key="copy">Almost there! Link your GitHub account so Sourcegraph can analyze your repositories.</span>,
						<span key="cta" styleName="onboarding-cta">
							<Link styl="button" href={urlToGitHubOAuth} onClick={() => EventLogger.logEvent("SubmitLinkGitHub")}>
							Import GitHub repos
							</Link>
						</span>]
					}
				</NotificationWell>
				<div styleName="header">
					<span styleName="repos-label">{" "}</span>
					<Input type="text"
						placeholder="Filter repositories..."
						ref={(c) => this._filterInput = c}
						onChange={this._handleSearch} />
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
	hasLinkedGitHub: React.PropTypes.bool,
};

export default CSSModules(DashboardRepos, styles);
