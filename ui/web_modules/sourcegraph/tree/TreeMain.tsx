// tslint:disable

import * as React from "react";
import TreeList from "sourcegraph/tree/TreeList";
import {urlToTree} from "sourcegraph/tree/routes";
import {treeParam} from "sourcegraph/tree/index";
import {trimRepo} from "sourcegraph/repo/index";
import * as styles from "./styles/Tree.css";
import Helmet from "react-helmet";

class TreeMain extends React.Component<any, any> {
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

	render(): JSX.Element | null {
		if (!this.props.commitID) return null;

		const path = treeParam(this.props.routeParams.splat);

		return (
			<div className={styles.tree_container}>
				{/* Let RepoMain set title for the root path. */}
				{path !== "/" && <Helmet title={`${path} · ${trimRepo(this.props.repo)}`} />}
				<br />
				<TreeList
					repo={this.props.repo}
					rev={this.props.rev}
					commitID={this.props.commitID}
					path={path}
					location={this.props.location}
					route={this.props.route}/>
			</div>
		);
	}
}

export default TreeMain;
