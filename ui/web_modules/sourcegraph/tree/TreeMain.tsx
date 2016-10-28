import * as React from "react";
import Helmet from "react-helmet";
import {Route} from "react-router";
import {RouteParams} from "sourcegraph/app/routeParams";
import {Base, GridCol, Panel, RepoLink} from "sourcegraph/components";
import {colors} from "sourcegraph/components/utils";
import {trimRepo} from "sourcegraph/repo";
import {treeParam} from "sourcegraph/tree";
import {RepoNavContext} from "sourcegraph/tree/RepoNavContext";
import {TreeList} from "sourcegraph/tree/TreeList";

interface Props {
	location?: any;
	repo: string;
	rev: string;
	commitID?: string;
	route?: Route;
	routeParams: RouteParams;
};

export class TreeMain extends React.Component<Props, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	render(): JSX.Element | null {
		if (!this.props.commitID) {
			return null;
		}

		const path = treeParam(this.props.routeParams.splat);

		return (
			<div>
				<Panel style={{borderBottom: `1px solid ${colors.coolGray4(0.6)}`}}>
					<Base py={2} px={3}>
						<RepoLink repo={this.props.repo} rev={this.props.rev} style={{marginRight: 4}} />
						<RepoNavContext params={this.props.routeParams} />
					</Base>
				</Panel>
				{/* Refactor once new Panel and Grid code has been merged in */}
				<GridCol col={9} style={{marginRight: "auto", marginLeft: "auto", marginTop: 16, float: "none"}}>
					{path !== "/" && <Helmet title={`${path} · ${trimRepo(this.props.repo)}`} />}
					<TreeList
						repo={this.props.repo}
						rev={this.props.rev}
						commitID={this.props.commitID}
						path={path} />
				</GridCol>
			</div>
		);
	}
}
