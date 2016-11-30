import * as React from "react";
import * as Relay from "react-relay";
import { Route } from "react-router";

import { EditorController, Props as ControllerProps } from "sourcegraph/blob/EditorController";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";
import { repoParam, repoPath, repoRev } from "sourcegraph/repo";
import { treeParam } from "sourcegraph/tree";
import { FileTree } from "sourcegraph/workbench/fileTree";

class WorkbenchComponent extends React.Component<ControllerProps, {}> {
	render(): JSX.Element {
		const files = this.props.root.repository.commit.commit.tree.files;
		return <div style={{
			display: "flex",
			flexDirection: "row",
			flex: "auto",
		}}>
			<FileTree
				files={files}
				repo={this.props.repo}
				rev={this.props.rev}
				path={this.props.path} />
			<EditorController {...this.props} />
		</div>;
	}
}

export const WorkbenchContainer = Relay.createContainer(WorkbenchComponent, {
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
