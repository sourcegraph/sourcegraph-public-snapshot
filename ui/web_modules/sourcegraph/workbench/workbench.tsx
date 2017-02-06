import * as autobind from "autobind-decorator";
import * as React from "react";
import * as Relay from "react-relay";
import { Route } from "react-router";
import { IRange } from "vs/editor/common/editorCommon";

import { abs, getRoutePattern } from "sourcegraph/app/routePatterns";
import { RouteParams, Router, RouterLocation, pathFromRouteParams, repoRevFromRouteParams } from "sourcegraph/app/router";
import "sourcegraph/blob/styles/Monaco.css";
import { ChromeExtensionToast } from "sourcegraph/components/ChromeExtensionToast";
import { Header } from "sourcegraph/components/Header";
import { OnboardingModals } from "sourcegraph/components/OnboardingModals";
import { PageTitle } from "sourcegraph/components/PageTitle";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { repoPath, repoRev } from "sourcegraph/repo";
import { RepoMain } from "sourcegraph/repo/RepoMain";
import { InfoPanelLifecycle } from "sourcegraph/workbench/info/sidebar";
import { WorkbenchShell } from "sourcegraph/workbench/shell";

interface Props {
	repo: string;
	rev: string | null;
	isSymbolUrl: boolean;
	routes: Route[];
	params: RouteParams;
	selection: IRange;
	location: RouterLocation;

	relay: any;
	root: GQL.IRoot;
}

// WorkbenchComponent loads the VSCode workbench shell, or our home made file
// tree and Editor, depending on configuration. To learn about VSCode and the
// way we use it, read README.vscode.md.
@autobind
class WorkbenchComponent extends React.Component<Props, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };

	componentWillMount(): void {
		document.title = "Sourcegraph";
	}

	render(): JSX.Element | null {
		let repository: GQL.IRepository;
		let symbol: GQL.ISymbol | undefined;
		let selection: IRange;
		let path: string;
		if (this.props.isSymbolUrl) {
			if (!this.props.root.symbols || this.props.root.symbols.length === 0) {
				return (
					<Header
						title="404"
						subtitle="Symbol not found." />
				);
			}
			// Assume that there is only one symbol for now
			symbol = this.props.root.symbols[0];
			repository = symbol.repository;
			path = symbol.path;
			selection = RangeOrPosition.fromLSPPosition(symbol).toMonacoRangeAllowEmpty();
		} else {
			if (!this.props.root.repository || !this.props.root.repository.commit.commit || !this.props.root.repository.commit.commit.tree) {
				return (
					<Header
						title="404"
						subtitle="Repository not found." />
				);
			}
			repository = this.props.root.repository;
			selection = this.props.selection;
			path = pathFromRouteParams(this.props.params);
		}
		const commitID = repository.commit.commit!.sha1;
		return <div style={{ display: "flex", height: "100%" }}>
			<BlobPageTitle repo={this.props.repo} path={path} />
			<RepoMain {...this.props} repository={repository} commit={repository.commit}>
				<OnboardingModals location={this.props.location} />
				<ChromeExtensionToast location={this.props.location} layout={() => this.forceUpdate()} />
				<WorkbenchShell
					repo={repository.uri}
					commitID={commitID}
					path={path}
					selection={symbol ? RangeOrPosition.fromLSPPosition(symbol).toMonacoRangeAllowEmpty() : this.props.selection} />
				<InfoPanelLifecycle repo={repository} fileEventProps={{ repo: repository.uri, rev: commitID, path: path }} />
			</RepoMain>
		</div>;
	}
}

function BlobPageTitle({ repo, path }: { repo: string | null, path: string }): JSX.Element {
	const base = path.split("/").pop() || path;
	if (!repo) {
		return <PageTitle title={base} />;
	}
	repo = repo.replace(/^github.com\//, "");
	const title = `${base} · ${repo}`;
	return <PageTitle title={title} />;
}

const WorkbenchContainer = Relay.createContainer(WorkbenchComponent, {
	initialVariables: {
		repo: "",
		rev: "",
		id: "",
		mode: "",
		isSymbolUrl: false
	},
	fragments: {
		root: () => Relay.QL`
			fragment on Root {
				symbols(id: $id, mode: $mode) @include(if: $isSymbolUrl) {
					path
					line
					character
					repository {
						uri
						description
						defaultBranch
						commit(rev: $rev) {
							commit {
								sha1
								languages
								tree(recursive: true) {
									files {
										name
									}
								}
							}
							cloneInProgress
						}
					}
				}
				repository(uri: $repo) @skip(if: $isSymbolUrl) {
					uri
					description
					defaultBranch
					commit(rev: $rev) {
						commit {
							sha1
							languages
							tree(recursive: true) {
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

// TODO(john): make this use router context.
export function Workbench(props: { params: any; location: RouterLocation, routes: Route[] }): JSX.Element {
	let rangeOrPosition: RangeOrPosition = RangeOrPosition.fromOneIndexed(1);
	if (props.location && props.location.hash && props.location.hash.startsWith("#L")) {
		rangeOrPosition = RangeOrPosition.parse(props.location.hash.replace(/^#L/, "")) || rangeOrPosition;
	}
	const isSymbolUrl = getRoutePattern(props.routes) === abs.goSymbol;
	let id: string | null = null;
	let mode: string | null = null;
	let repo: string | null = null;
	let rev: string | null = null;
	if (isSymbolUrl) {
		id = props.params.splat;
		mode = "go";
	} else {
		const repoRevString = repoRevFromRouteParams(props.params);
		repo = repoPath(repoRevString);
		rev = repoRev(repoRevString);
	}
	return <Relay.RootContainer
		Component={WorkbenchContainer}
		route={{
			name: "Root",
			queries: {
				root: () => Relay.QL`
					query { root }
				`,
			},
			params: {
				repo,
				rev,
				id,
				mode,
				isSymbolUrl,
				routes: props.routes,
				params: props.params,
				selection: rangeOrPosition.toMonacoRangeAllowEmpty(),
				location: props.location,
			},
		}}
	/>;
};
