import * as cloneDeep from "lodash/cloneDeep";
import * as React from "react";
import Helmet from "react-helmet";
import * as Relay from "react-relay";
import {InjectedRouter, Route} from "react-router";
import {RouteParams} from "sourcegraph/app/routeParams";
import {GridCol, Panel, RepoLink} from "sourcegraph/components";
import {colors} from "sourcegraph/components/utils";
import {whitespace} from "sourcegraph/components/utils/index";
import {Location} from "sourcegraph/Location";
import {repoParam, repoPath, repoRev, trimRepo} from "sourcegraph/repo";
import {RepoMain} from "sourcegraph/repo/RepoMain";
import {treeParam} from "sourcegraph/tree";
import {RepoNavContext} from "sourcegraph/tree/RepoNavContext";
import {TreeList} from "sourcegraph/tree/TreeList";

interface Props {
	repo: string;
	rev: string;
	path: string;

	location: any;
	routes: Route[];
	params: RouteParams;

	relay: any;
	root: GQL.IRoot;
};

interface Context {
	router: InjectedRouter;
}

export class TreeMainComponent extends React.Component<Props, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	constructor(props: Props, context: Context) {
		super(props);
		this._redirectToCanonicalURI(props, context);
	}

	componentWillReceiveProps(nextProps: Props, nextContext: Context): void {
		this._redirectToCanonicalURI(nextProps, nextContext);
	}

	_redirectToCanonicalURI(props: Props, context: Context): void {
		if (props.root.repository && props.repo !== props.root.repository.uri) {
			setTimeout(function(): void {
				let locCopy = cloneDeep(props.location);
				locCopy.pathname = props.location.pathname.replace(new RegExp(props.repo, "g"), props.root.repository.uri);
				context.router.replace(locCopy);
			}, 0);
		}
	}

	render(): JSX.Element | null {
		return (
			<RepoMain
				repo={this.props.repo}
				rev={this.props.rev}
				repository={this.props.root.repository}
				commit={this.props.root.repository && this.props.root.repository.commit}
				location={this.props.location}
				routes={this.props.routes}
				params={this.props.params}
				relay={this.props.relay}
			>
				<div>
					<Panel style={{borderBottom: `1px solid ${colors.coolGray4(0.6)}`}}>
						<div style={{
								padding: `${whitespace[2]} ${whitespace[3]}`,
							}}>
							<RepoLink repo={this.props.repo} rev={this.props.rev} style={{marginRight: 4}} />
							<RepoNavContext params={this.props.params} />
						</div>
					</Panel>
					{/* Refactor once new Panel and Grid code has been merged in */}
					<GridCol col={9} style={{marginRight: "auto", marginLeft: "auto", marginTop: 16, float: "none"}}>
						{this.props.path !== "/" && <Helmet title={`${this.props.path} · ${trimRepo(this.props.repo)}`} />}
						<TreeList
							repo={this.props.repo}
							rev={this.props.rev}
							path={this.props.path}
							tree={this.props.root.repository && this.props.root.repository.commit.commit && this.props.root.repository.commit.commit.tree} />
					</GridCol>
				</div>
			</RepoMain>
		);
	}
}

const TreeMainContainer = Relay.createContainer(TreeMainComponent, {
	initialVariables: {
		repo: "",
		rev: "",
		path: "",
	},
	fragments: {
		root: () => Relay.QL`
			fragment on Root {
				repository(uri: $repo) {
					uri
					description
					commit(rev: $rev) {
						commit {
							sha1
							tree(path: $path) {
								directories {
									name
								}
								files {
									name
								}
							}
						}
						cloneInProgress
					}
				}
			}
		`,
	},
});

export const TreeMain = function(props: {params: any; location: Location, routes: Route[]}): JSX.Element {
	const repoSplat = repoParam(props.params.splat);
	return <Relay.RootContainer
		Component={TreeMainContainer}
		route={{
			name: "Root",
			queries: {
				root: () => Relay.QL`
					query { root }
				`,
			},
			params: {
				repo: repoPath(repoSplat),
				rev: repoRev(repoSplat),
				path: treeParam(props.params.splat),
				location: props.location,
				routes: props.routes,
				params: props.params,
			},
		}}
	/>;
};
