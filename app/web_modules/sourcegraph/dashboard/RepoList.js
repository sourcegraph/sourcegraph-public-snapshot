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
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		return (
			<div>
				{this.state.repos.length > 0 &&
					<div>
						{context.currentUser && <div styleName="list-section-header">Repositories</div>}
						{this.state.repos.map((repo, i) =>
							<div key={i}>
								<div styleName="list-item">
									<div styleName="uri-container">
										<div styleName="uri">
											<RepoLink repo={repo.URI || `github.com/${repo.Owner}/${repo.Name}`} />
										</div>
									</div>

									{repo.Description && <div>
										<p styleName="description">{repo.Description}</p>
									</div>}

									{repo.Examples && repo.Examples.map((functions, j) =>
										<div styleName="function-example-container" key={j}>
											<span styleName="function" key={functions.Functions.Path}>
												<a href={functions.Functions.Path}>
													<code>{qualifiedNameAndType(functions.Functions)}</code>
												</a>
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
};


export default CSSModules(RepoList, styles);
