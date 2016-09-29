// tslint:disable: typedef ordered-imports

import * as React from "react";
import {TreeList} from "sourcegraph/tree/TreeList";
import {treeParam} from "sourcegraph/tree";
import {trimRepo} from "sourcegraph/repo";
import Helmet from "react-helmet";
import {RepoLink} from "sourcegraph/components/RepoLink";
import {RepoNavContext} from "sourcegraph/tree/RepoNavContext";
import {Base, GridCol, Panel} from "sourcegraph/components";
import {colors} from "sourcegraph/components/utils";

interface Props {
	location?: any;
	repo: string;
	rev: string;
	commitID?: string;
	repoNavContext: any;
	route?: any;
	routeParams: any;
};

type State = any;

export class TreeMain extends React.Component<Props, State> {
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
				{RepoNavContext &&
					<Panel style={{borderBottom: `1px solid ${colors.coolGray4(0.6)}`}}>
						<Base py={2} px={3}>
							<RepoLink repo={this.props.repo} rev={this.props.rev} style={{marginRight: 4}} />
							<RepoNavContext params={this.props.routeParams} />
						</Base>
					</Panel>
				}
				{/* Refactor once new Panel and Grid code has been merged in */}
				<GridCol col={9} style={{marginRight: "auto", marginLeft: "auto", marginTop: 16, float: "none"}}>
					{path !== "/" && <Helmet title={`${path} · ${trimRepo(this.props.repo)}`} />}
					<TreeList
						repo={this.props.repo}
						rev={this.props.rev}
						commitID={this.props.commitID}
						path={path}
						location={this.props.location}
						route={this.props.route}/>
				</GridCol>
			</div>
		);
	}
}
