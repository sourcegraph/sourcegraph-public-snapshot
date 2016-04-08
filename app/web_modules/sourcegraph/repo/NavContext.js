// @flow

import React from "react";

import RepoLink from "sourcegraph/components/RepoLink";

import CSSModules from "react-css-modules";
import styles from "./styles/Repo.css";

class NavContext extends React.Component {
	static propTypes = {
		repo: React.PropTypes.string.isRequired,
		rev: React.PropTypes.string,
		repoNavContext: React.PropTypes.element,
	};

	render() {
		return (
			<div style={{display: "inline-block"}}>
				<RepoLink repo={this.props.repo} />
				{this.props.repoNavContext}
			</div>
		);
	}
}

export default CSSModules(NavContext, styles);
