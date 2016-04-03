// @flow

import React from "react";
import {Link} from "react-router";
import Style from "./styles/Repo.css";
import urlTo from "sourcegraph/util/urlTo";

export default class NavContext extends React.Component {
	static propTypes = {
		repo: React.PropTypes.string.isRequired,
		rev: React.PropTypes.string,
		repoNavContext: React.PropTypes.element,
	};

	render() {
		return (
			<div className={Style.navContext}>
				<Link to={urlTo("repo", {splat: this.props.repo})} className={Style.repoName}>{this.props.repo}</Link>
				{this.props.repoNavContext}
			</div>
		);
	}
}
