import * as autobind from "autobind-decorator";
import * as React from "react";
import * as Relay from "react-relay";
import { Route } from "react-router";
import { IRange } from "vs/editor/common/editorCommon";

import { abs, getRoutePattern } from "sourcegraph/app/routePatterns";
import { RouteParams, Router, RouterLocation, pathFromRouteParams, repoRevFromRouteParams } from "sourcegraph/app/router";
import "sourcegraph/blob/styles/Monaco.css";
import { Heading, PageTitle } from "sourcegraph/components";
import { ChromeExtensionToast } from "sourcegraph/components/ChromeExtensionToast";
import { OnboardingModals } from "sourcegraph/components/OnboardingModals";
import { whitespace } from "sourcegraph/components/utils";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { repoPath, repoRev } from "sourcegraph/repo";
import { CloningRefresher, RepoMain } from "sourcegraph/repo/RepoMain";
import { Paywall, TrialEndingWarning } from "sourcegraph/user/Paywall";
import { InfoPanelLifecycle } from "sourcegraph/workbench/info/sidebar";
import { WorkbenchShell } from "sourcegraph/workbench/shell";

interface Props {
	repo: string | null;
	rev: string | null;
	isSymbolUrl: boolean;
	routes: Route[];
	params: RouteParams;
	selection: IRange | null;
	location: RouterLocation;

	relay: any;
	root: GQL.IRoot;
}

// WorkbenchComponent loads the VSCode workbench shell. To learn about VSCode and the
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
		let selection: IRange | null;
		let path: string;
		if (this.props.isSymbolUrl) {
			if (!this.props.root.symbols || this.props.root.symbols.length === 0) {
				return <Error
					code="404"
					desc="Symbol not found" />;
			}
			// Assume that there is only one symbol for now
			const symbol = this.props.root.symbols[0];
			repository = symbol.repository;
			path = symbol.path;
			selection = RangeOrPosition.fromLSPPosition(symbol).toMonacoRangeAllowEmpty();
		} else {
			if (!this.props.root.repository) {
				return <Error
					code="404"
					desc="Repository not found" />;
			}
			repository = this.props.root.repository;
			selection = this.props.selection;
			path = pathFromRouteParams(this.props.params);
		}

		if (repository.revState.cloneInProgress) {
			return <CloningRefresher relay={this.props.relay} />;
		}

		if (!repository.revState.commit && !repository.revState.zapRef) {
			return <Error
				code="404"
				desc="Revision not found" />;
		}

		const commitID = repository.revState.zapRef ? repository.revState.zapRef.base : repository.revState.commit!.sha1;
		return <div style={{ display: "flex", height: "100%" }}>
			<BlobPageTitle repo={this.props.repo} path={path} />
			{/* TODO(john): repo main takes the commit state for the 'y' hotkey, assume for now revState can be passed. */}
			<RepoMain {...this.props} repository={repository} commit={repository.revState}>
				<OnboardingModals location={this.props.location} />
				<ChromeExtensionToast location={this.props.location} layout={() => this.forceUpdate()} />
				<TrialEndingWarning layout={() => this.forceUpdate()} repo={repository} />
				<WorkbenchShell
					location={this.props.location}
					routes={this.props.routes}
					params={this.props.params}
					repo={repository.uri}
					rev={this.props.rev}
					commitID={commitID}
					zapRef={repository.revState.zapRef ? this.props.rev! : undefined}
					branch={repository.revState.zapRef ? repository.revState.zapRef.branch : undefined}
					path={path}
					selection={selection} />
				<InfoPanelLifecycle repo={repository} fileEventProps={{ repo: repository.uri, rev: commitID, path: path }} />
			</RepoMain>
		</div>;
	}
}

function Error({ code, desc }: { code: string, desc: string }): JSX.Element {
	return <div style={{ textAlign: "center", marginTop: whitespace[8] }}>
		<Heading level={2}>{code}</Heading>
		<Heading level={4} color="gray">{desc}</Heading>
	</div>;
}

function BlobPageTitle({ repo, path }: { repo: string | null, path: string }): JSX.Element {
	const base = path.split("/").pop() || path;
	if (!repo) {
		return <PageTitle title={base} />;
	}
	repo = repo.replace(/^github.com\//, "");
	let title = base === "/" ? repo : `${base} · ${repo}`;
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
						expirationDate
						revState(rev: $rev) {
							zapRef {
								base
								branch
							}
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
					expirationDate
					revState(rev: $rev) {
						zapRef {
							base
							branch
						}
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
		renderFailure={Paywall}
	/>;
};
