import * as React from "react";
import TreeList from "sourcegraph/tree/TreeList";
import {urlToTree} from "sourcegraph/tree/routes";
import {treeParam} from "sourcegraph/tree";
import {trimRepo} from "sourcegraph/repo";
import CSSModules from "react-css-modules";
import styles from "./styles/Tree.css";
import Helmet from "react-helmet";

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

	state: {
		query: string;
	} = {
		query: "",
	};

	_onSelectPath(path: string) {
		this.context.router.push(urlToTree(this.props.repo, this.props.rev, path));
	}

	_onChangeQuery(query: string) {
		this.setState({query: query});
	}

	render() {
		if (!this.props.commitID) return null;

		const path = treeParam(this.props.routeParams.splat);

		return (
			<div styleName="tree_container">
				{/* Let RepoMain set title for the root path. */}
				{path !== "/" && <Helmet title={`${path} · ${trimRepo(this.props.repo)}`} />}
				<br />
				<TreeList
					repo={this.props.repo}
					rev={this.props.rev}
					commitID={this.props.commitID}
					path={path}
					query={this.state.query || ""}
					location={this.props.location}
					route={this.props.route}
					onChangeQuery={this._onChangeQuery.bind(this)}
					onSelectPath={this._onSelectPath.bind(this)}/>
			</div>
		);
	}
}

export default CSSModules(TreeMain, styles);
