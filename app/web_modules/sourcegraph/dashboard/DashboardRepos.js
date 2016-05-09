import React from "react";
import RepoLink from "sourcegraph/components/RepoLink";

import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";

class DashboardRepos extends React.Component {
	static contextTypes = {
		signedIn: React.PropTypes.bool.isRequired,
		githubToken: React.PropTypes.object,
	};

	// _repoSort is a comparison function that sorts more recently
	// pushed repos first.
	_repoSort(a, b) {
		if (a.PushedAt < b.PushedAt) return 1;
		else if (a.PushedAt > b.PushedAt) return -1;
		return 0;
	}

	render() {
		let repos = this.props.repos.sort(this._repoSort);

		return (
			<div styleName="list">
				{this.context.signedIn && <div styleName="list-section-header">Repositories</div>}
				{this.context.githubToken && repos.length === 0 && <div styleName="list-item-loading">Loading...</div>}
				{!this.context.githubToken && <div styleName="list-item-loading">Link your GitHub account above to see your repositories here.</div>}
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
