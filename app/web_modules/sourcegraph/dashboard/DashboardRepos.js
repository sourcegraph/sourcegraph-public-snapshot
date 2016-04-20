import React from "react";
import RepoLink from "sourcegraph/components/RepoLink";

import {qualifiedNameAndType} from "sourcegraph/def/Formatter";
import context from "sourcegraph/app/context";

import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";

class DashboardRepos extends React.Component {
	_repoSort(a, b) {
		const name = (repo) => `${repo.Owner}/${repo.Name}`;
		const nameA = name(a);
		const nameB = name(b);
		if (nameA < nameB) return -1;
		else if (nameA > nameB) return 1;
		return 0;
	}

	render() {
		let repos = this.props.exampleRepos.concat(this.props.repos.sort(this._repoSort));

		return (repos.length > 0 &&
			<div styleName="list">
				{context.currentUser && <div styleName="list-section-header">Repositories</div>}
				{repos.map((repo, i) =>
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
		);
	}
}

DashboardRepos.propTypes = {
	repos: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
	exampleRepos: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
};

export default CSSModules(DashboardRepos, styles);
