import React from "react";

import Component from "sourcegraph/Component";
import repoLink from "sourcegraph/util/repoLink";
import Styles from "./styles/Dashboard.css";
import TimeAgo from "sourcegraph/util/TimeAgo";

class RepoList extends Component {
	constructor(props) {
		super(props);
		this._repoDisabled = this._repoDisabled.bind(this);
		this._repoTime = this._repoTime.bind(this);
		this._disabledReason = this._disabledReason.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_repoDisabled(repo) {
		return !repo.URI && !(!repo.GitHubID || repo.Language === "Go");
	}

	_disabledReason(repo) {
		return `${repo.Language || ""} coming soon`;
	}

	_repoTime(repo) {
		return repo.UpdatedAt || repo.PushedAt || repo.CreatedAt;
	}

	render() {
		return (
			<div>
				{this.state.repos.length > 0 &&
					<div>
						<div className={Styles.repo_section_header}>{this.state.reposDisabled ? `Coming soon` : `Go Repositories`}</div>
						{this.state.repos.map((repo, i) =>
							<div key={i}>
								<div className={this.state.reposDisabled ? Styles.list_item_disabled : Styles.list_item} key={i}>
									<div>
										<span className={Styles.repo_title}>
											{repoLink(repo.URI || `github.com/${repo.Owner}/${repo.Name}`, this._repoDisabled(repo))}
										</span>
										{this.state.reposDisabled &&
											<span className={Styles.repo_disable_reason}>{this._disabledReason(repo)}</span>
										}
									</div>
									<div>
										<p className={Styles.repo_description}>{repo.Description}</p>
										{this._repoTime(repo) && <p className={Styles.repo_updated}>Updated <TimeAgo time={this._repoTime(repo)} /></p>}
									</div>
								</div>
							</div>
						)}
					</div>
				}
			</div>
		);
	}
}

RepoList.propTypes = {
	repos: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
	reposDisabled: React.PropTypes.bool.isRequired,
};


export default RepoList;
