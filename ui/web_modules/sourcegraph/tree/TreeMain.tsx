import * as cloneDeep from "lodash/cloneDeep";
import * as React from "react";
import Helmet from "react-helmet";
import * as Relay from "react-relay";
import {InjectedRouter, Route} from "react-router";
import {RouteParams} from "sourcegraph/app/routeParams";
import {GridCol, Panel, RepoLink} from "sourcegraph/components";
import {colors} from "sourcegraph/components/utils";
import {whitespace} from "sourcegraph/components/utils/index";
import {trimRepo} from "sourcegraph/repo";
import {RepoMain} from "sourcegraph/repo/RepoMain";
import {treeParam} from "sourcegraph/tree";
import {RepoNavContext} from "sourcegraph/tree/RepoNavContext";
import {TreeList} from "sourcegraph/tree/TreeList";

interface Props {
	location?: any;
	repo: string;
	rev: string;
	route?: Route;
	routeParams: RouteParams;

	repoObj?: any;
	isCloning?: boolean;
	routes: any[];
};

type PropsWithRoot = Props & {root: GQL.IRoot};

interface Context {
	router: InjectedRouter;
}

export class TreeMainComponent extends React.Component<PropsWithRoot, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	constructor(props: PropsWithRoot, context: Context) {
		super(props);
		this._redirectToCanonicalURI(props, context);
	}

	componentWillReceiveProps(nextProps: PropsWithRoot, nextContext: Context): void {
		this._redirectToCanonicalURI(nextProps, nextContext);
	}

	_redirectToCanonicalURI(props: PropsWithRoot, context: Context): void {
		if (props.repo !== props.root.repository.uri) {
			setTimeout(function(): void {
				let locCopy = cloneDeep(props.location);
				locCopy.pathname = props.location.pathname.replace(new RegExp(props.repo, "g"), props.root.repository.uri);
				context.router.replace(locCopy);
			}, 0);
		}
	}

	render(): JSX.Element | null {
		const path = treeParam(this.props.routeParams.splat);

		return (
			<RepoMain
				location={this.props.location}
				repo={this.props.repo}
				rev={this.props.rev}
				commit={this.props.root.repository.commit}
				repoObj={this.props.repoObj}
				isCloning={this.props.isCloning}
				route={this.props.route}
				routes={this.props.routes}
			>
				<div>
					<Panel style={{borderBottom: `1px solid ${colors.coolGray4(0.6)}`}}>
						<div style={{
								padding: `${whitespace[2]} ${whitespace[3]}`,
							}}>
							<RepoLink repo={this.props.repo} rev={this.props.rev} style={{marginRight: 4}} />
							<RepoNavContext params={this.props.routeParams} />
						</div>
					</Panel>
					{/* Refactor once new Panel and Grid code has been merged in */}
					<GridCol col={9} style={{marginRight: "auto", marginLeft: "auto", marginTop: 16, float: "none"}}>
						{path !== "/" && <Helmet title={`${path} · ${trimRepo(this.props.repo)}`} />}
						<TreeList
							repo={this.props.repo}
							rev={this.props.rev}
							path={path}
							tree={this.props.root.repository.commit && this.props.root.repository.commit.tree} />
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
					commit(rev: $rev) {
						tree(path: $path) {
							directories {
								name
							}
							files {
								name
							}
						}
					}
				}
			}
		`,
	},
});

export const TreeMain = function(props: Props): JSX.Element {
	return <Relay.RootContainer
		Component={TreeMainContainer}
		route={{
			name: "Root",
			queries: {
				root: () => Relay.QL`
					query { root }
				`,
			},
			params: props,
		}}
	/>;
};
