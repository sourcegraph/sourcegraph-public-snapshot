import React from "react";

import Component from "sourcegraph/Component";
import repoLink from "sourcegraph/util/repoLink";
import TimeAgo from "sourcegraph/util/TimeAgo";

import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";

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
						<div styleName="list-section-header">{this.state.reposDisabled ? `Coming soon` : `Go Repositories`}</div>
						{this.state.repos.map((repo, i) =>
							<div key={i}>
								<div styleName={this.state.reposDisabled ? "list-item-disabled" : "list-item"} key={i}>
									<div>
										<span styleName="uri">
											{repoLink(repo.URI || `github.com/${repo.Owner}/${repo.Name}`, this._repoDisabled(repo))}
										</span>
										{this.state.reposDisabled &&
											<span styleName="disable-reason">{this._disabledReason(repo)}</span>
										}
									</div>
									<div>
										<p styleName="description">{repo.Description}</p>
										{this._repoTime(repo) && <p styleName="updated">Updated <TimeAgo time={this._repoTime(repo)} /></p>}
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


export default CSSModules(RepoList, styles);
