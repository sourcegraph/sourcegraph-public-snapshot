import React from "react";
import Component from "sourcegraph/Component";
import RepoList from "sourcegraph/dashboard/RepoList";
import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";

class DashboardRepos extends Component {
	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	_repoSort(a, b) {
		const name = (repo) => `${repo.Owner}/${repo.Name}`;
		const nameA = name(a);
		const nameB = name(b);
		if (nameA < nameB) return -1;
		else if (nameA > nameB) return 1;
		return 0;
	}

	render() {
		return (
			<div styleName="list">
				<RepoList repos={this.state.exampleRepos.concat(this.state.repos.sort(this._repoSort))} />
			</div>
		);
	}
}

DashboardRepos.propTypes = {
	repos: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
	exampleRepos: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
};

export default CSSModules(DashboardRepos, styles);
