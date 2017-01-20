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
import { TourOverlay } from "sourcegraph/components/TourOverlay";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { repoPath, repoRev } from "sourcegraph/repo";
import { RepoMain } from "sourcegraph/repo/RepoMain";
import { Features } from "sourcegraph/util/features";
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

	render(): JSX.Element | null {
		if (!this.props.root.repository || !this.props.root.repository.commit.commit || !this.props.root.repository.commit.commit.tree) {
			return (
				<Header
					title="404"
					subtitle="Repository not found." />
			);
		}
		if (this.props.isSymbolUrl && this.props.root.repository.symbols.length === 0) {
			return null;
		}
		const symbol = this.props.isSymbolUrl ? this.props.root.repository.symbols[0] : null; // Assume for now it's ok to take the first.
		return <div style={{ display: "flex", height: "100%" }}>
			<RepoMain {...this.props} repository={this.props.root.repository} commit={this.props.root.repository.commit}>
				{this.props.location.query["tour"] && <TourOverlay location={this.props.location} />}
				<OnboardingModals location={this.props.location} />
				<ChromeExtensionToast location={this.props.location} layout={() => this.forceUpdate()} />
				<WorkbenchShell
					repo={this.props.repo}
					rev={this.props.rev}
					path={symbol ? symbol.path : pathFromRouteParams(this.props.params)}
					selection={symbol ? RangeOrPosition.fromLSPPosition(symbol).toMonacoRangeAllowEmpty() : this.props.selection} />
				{Features.projectWow.isEnabled() && <InfoPanelLifecycle isSymbolUrl={this.props.isSymbolUrl} repo={this.props.root.repository} />}
			</RepoMain>
		</div>;
	}
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
				repository(uri: $repo) {
					uri
					description
					defaultBranch
					symbols(id: $id, mode: $mode, rev: $rev) @include(if: $isSymbolUrl) {
						path
						line
						character
					}
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
	const repoRevString = repoRevFromRouteParams(props.params);
	let rangeOrPosition: RangeOrPosition = RangeOrPosition.fromOneIndexed(1);
	if (props.location && props.location.hash && props.location.hash.startsWith("#L")) {
		rangeOrPosition = RangeOrPosition.parse(props.location.hash.replace(/^#L/, "")) || rangeOrPosition;
	}
	const isSymbolUrl = getRoutePattern(props.routes) === abs.symbol;
	let id: string | null = null;
	let mode: string | null = null;
	if (isSymbolUrl) {
		id = props.params.splat[1];
		mode = props.params.mode;
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
				repo: repoPath(repoRevString),
				rev: repoRev(repoRevString),
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
