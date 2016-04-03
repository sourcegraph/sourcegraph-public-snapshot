// @flow weak

import React from "react";
import TreeSearch from "sourcegraph/tree/TreeSearch";
import CSSModules from "react-css-modules";
import styles from "./styles/Repo.css";

import Header from "sourcegraph/components/Header";

class RepoMain extends React.Component {
	static propTypes = {
		location: React.PropTypes.object,
		repo: React.PropTypes.string,
		rev: React.PropTypes.string,
		repoObj: React.PropTypes.object,
		main: React.PropTypes.element,
		isCloning: React.PropTypes.bool,
		route: React.PropTypes.object,
	};

	render() {
		if (this.props.repoObj && this.props.repoObj.Error) {
			return (
				<Header
					title={`${this.props.repoObj.Error.Status}`}
					subtitle={`Repository "${this.props.repo}" is not available.`} />
			);
		}

		if (!this.props.repo || !this.props.rev) return null;

		if (this.props.isCloning) {
			return (
				<Header
					title="Sourcegraph is cloning this repository"
					subtitle="Refresh this page in a minute." />
			);
		}

		return (
			<div>
				{this.props.main}
				{this.props.route.disableTreeSearchOverlay ? null : <TreeSearch repo={this.props.repo} rev={this.props.rev} path="/" overlay={true} location={this.props.location} route={this.props.route} />}
			</div>
		);
	}
}

export default CSSModules(RepoMain, styles);
