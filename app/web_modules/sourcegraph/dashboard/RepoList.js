import React from "react";

import Component from "sourcegraph/Component";
import RepoLink from "sourcegraph/components/RepoLink";

import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";

import {qualifiedNameAndType} from "sourcegraph/def/Formatter";
import context from "sourcegraph/app/context";

class RepoList extends Component {
	constructor(props) {
		super(props);
		this._repoDisabled = this._repoDisabled.bind(this);
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

	render() {
		return (
			<div>
				{this.state.repos.length > 0 &&
					<div>
						{context.currentUser && <div styleName="list-section-header">{this.state.reposDisabled ? `Coming soon` : `Go Repositories`}</div>}
						{this.state.repos.map((repo, i) =>
							<div key={i}>
								<div styleName={this.state.reposDisabled ? "list-item-disabled" : "list-item"}>
									<div styleName="uri-container">
										<div styleName="uri">
											<RepoLink repo={repo.URI || `github.com/${repo.Owner}/${repo.Name}`} disabledLink={this._repoDisabled(repo)} />
										</div>

										{this.state.reposDisabled &&
											<div styleName="disable-reason">{this._disabledReason(repo)}</div>}
									</div>

									{repo.Description && <div>
										<p styleName="description">{repo.Description}</p>
									</div>}

									{repo.Examples && repo.Examples.map((functions, j) =>
										<div styleName="function-example-container" key={j}>
											<span styleName="function" key={functions.Functions.Path}>
												<a href={functions.Functions.Path}><code>{qualifiedNameAndType(functions.Functions)}</code></a>
											</span>
											{functions.Functions.FunctionCallCount &&
												<span styleName="function-call-count" key={functions.Functions.FunctionCallCount}>
													<a href={functions.Functions.Path}><code>{functions.Functions.FunctionCallCount} references</code></a>
												</span>}
										</div>
									)}
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
