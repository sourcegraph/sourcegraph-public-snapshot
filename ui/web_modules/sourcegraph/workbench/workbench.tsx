import * as React from "react";
import * as Relay from "react-relay";
import { Route } from "react-router";

import { EditorController, Props as ControllerProps } from "sourcegraph/blob/EditorController";
import { ChromeExtensionToast } from "sourcegraph/components/ChromeExtensionToast";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { repoParam, repoPath, repoRev } from "sourcegraph/repo";
import { treeParam } from "sourcegraph/tree";
import { Features } from "sourcegraph/util/features";
import { FileTree } from "sourcegraph/workbench/fileTree";
import { WorkbenchShell } from "sourcegraph/workbench/shell";

import { RepoMain } from "sourcegraph/repo/RepoMain";

// WorkbenchComponent loads the VSCode workbench shell, or our home made file
// tree and Editor, depending on configuration. To learn about VSCode, and the
// way we use it, read README.vscode.md.
class WorkbenchComponent extends React.Component<ControllerProps, {}> {
	render(): JSX.Element | null {
		if (!this.props.root.repository || !this.props.root.repository.commit.commit || !this.props.root.repository.commit.commit.tree) {
			return null;
		}
		const files = this.props.root.repository.commit.commit.tree.files;
		if (Features.workbench.isEnabled()) {
			return <div style={{ display: "flex", height: "100%" }}>
				<RepoMain {...this.props} {...this.props.root} {...this.props.root.repository}>
					<WorkbenchShell />
				</RepoMain>
			</div>;
		}
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

export function Workbench(props: { params: any; location: Location, routes: Route[] }): JSX.Element {
	const repoSplat = repoParam(props.params.splat);
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
				repo: repoPath(repoSplat),
				rev: repoRev(repoSplat),
				path: treeParam(props.params.splat),
				routes: props.routes,
				params: props.params,
				selection: selection,
				location: props.location,
			},
		}}
		/>;
};
