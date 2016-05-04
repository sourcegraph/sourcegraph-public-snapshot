import React from "react";
import RepoLink from "sourcegraph/components/RepoLink";

import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";

class DashboardRepos extends React.Component {
	static contextTypes = {
		signedIn: React.PropTypes.bool.isRequired,
	};

	_repoSort(a, b) {
		const name = (repo) => `${repo.Owner}/${repo.Name}`;
		const nameA = name(a);
		const nameB = name(b);
		if (nameA < nameB) return -1;
		else if (nameA > nameB) return 1;
		return 0;
	}

	render() {
		let repos = this.props.repos.sort(this._repoSort);

		return (
			<div styleName="list">
				{this.context.signedIn && <div styleName="list-section-header">Repositories</div>}
				{repos.length === 0 && <div styleName="list-item-loading">Loading...</div>}
				{repos.length > 0 && repos.map((repo, i) =>
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
						</div>
					</div>
				)}
			</div>
		);
	}
}

DashboardRepos.propTypes = {
	repos: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
};

export default CSSModules(DashboardRepos, styles);
