// @flow

import React from "react";
import TreeSearch from "sourcegraph/tree/TreeSearch";
import {urlToTree} from "sourcegraph/tree/routes";
import {treeParam} from "sourcegraph/tree";
import CSSModules from "react-css-modules";
import styles from "./styles/Tree.css";

class TreeMain extends React.Component {
	static propTypes = {
		location: React.PropTypes.object,
		repo: React.PropTypes.string,
		rev: React.PropTypes.string,
		commitID: React.PropTypes.string,
		route: React.PropTypes.object,
		routeParams: React.PropTypes.object.isRequired,
	};

	static contextTypes = {
		router: React.PropTypes.object.isRequired,
	};

	_onSelectPath(path: string) {
		this.context.router.push(urlToTree(this.props.repo, this.props.rev, path));
	}

	_onChangeQuery(query: string) {
		this.context.router.replace({...this.props.location, query: {q: query || undefined}}); // eslint-disable-line no-undefined
	}

	render() {
		if (!this.props.commitID) return null;

		return (
			<div styleName="tree-container">
				<TreeSearch
					repo={this.props.repo}
					rev={this.props.rev}
					commitID={this.props.commitID}
					path={treeParam(this.props.routeParams.splat)}
					query={this.props.location.query.q || ""}
					location={this.props.location}
					route={this.props.route}
					onChangeQuery={this._onChangeQuery.bind(this)}
					onSelectPath={this._onSelectPath.bind(this)}/>
			</div>
		);
	}
}

export default CSSModules(TreeMain, styles);
