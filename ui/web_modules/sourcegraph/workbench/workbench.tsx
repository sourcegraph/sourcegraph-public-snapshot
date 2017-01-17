import * as autobind from "autobind-decorator";
import * as React from "react";
import * as Relay from "react-relay";
import { Route } from "react-router";

import { RouteParams, Router, RouterLocation, getBlobPropsFromRouter, pathFromRouteParams, repoRevFromRouteParams } from "sourcegraph/app/router";
import { EditorController } from "sourcegraph/blob/EditorController";
import "sourcegraph/blob/styles/Monaco.css";
import { ChromeExtensionToast } from "sourcegraph/components/ChromeExtensionToast";
import { OnboardingModals } from "sourcegraph/components/OnboardingModals";
import { TourOverlay } from "sourcegraph/components/TourOverlay";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { repoPath, repoRev } from "sourcegraph/repo";
import { RepoMain } from "sourcegraph/repo/RepoMain";
import { Features } from "sourcegraph/util/features";
import { FileTree } from "sourcegraph/workbench/fileTree";
import { WorkbenchShell } from "sourcegraph/workbench/shell";

interface Props {
	repo: string;
	rev: string | null;
	path: string;
	routes: Route[];
	params: RouteParams;
	selection: RangeOrPosition | null;
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
			return null;
		}
		const files = this.props.root.repository.commit.commit.tree.files;
		if (!Features.workbench.isEnabled()) {
			return <div style={{
				display: "flex",
				flexDirection: "column",
				flex: "auto",
				width: "100%",
			}}>
				<ChromeExtensionToast location={this.props.location} />
				<div style={{
					display: "flex",
					flexDirection: "row",
					flex: "auto",
					width: "100%",
				}}>
					<FileTree
						files={files}
						repo={this.props.repo}
						rev={this.props.rev}
						path={this.props.path} />
					<EditorController {...this.props} />
				</div>
			</div>;

		}
		return <div style={{ display: "flex", height: "100%" }}>
			<RepoMain {...this.props} repository={this.props.root.repository} commit={this.props.root.repository.commit}>
				{this.props.location.query["tour"] && <TourOverlay location={this.props.location} />}
				<OnboardingModals location={this.props.location} />
				<ChromeExtensionToast location={this.props.location} />
				<WorkbenchShell { ...getBlobPropsFromRouter(this.context.router) } />
			</RepoMain>
		</div>;
	}
}

const WorkbenchContainer = Relay.createContainer(WorkbenchComponent, {
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
	const repoRevString = repoRevFromRouteParams(props.params);
	let selection = null;
	if (props.location && props.location.hash && props.location.hash.startsWith("#L")) {
		selection = RangeOrPosition.parse(props.location.hash.replace(/^#L/, ""));
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
				path: pathFromRouteParams(props.params),
				routes: props.routes,
				params: props.params,
				selection: selection,
				location: props.location,
			},
		}}
		/>;
};
