import { hover, select } from "glamor";
import * as React from "react";
import { InjectedRouter, Link } from "react-router";
import { ITree } from "vs/base/parts/tree/browser/tree";
import { Tree } from "vs/base/parts/tree/browser/treeImpl";

import { urlToBlob } from "sourcegraph/blob/routes";
import { Heading } from "sourcegraph/components/Heading";
import { colors } from "sourcegraph/components/utils/colors";
import { Events } from "sourcegraph/util/constants/AnalyticsConstants";
import { urlTo } from "sourcegraph/util/urlTo";
import { Controller, FileTreeDataSource, Node, Renderer, makeTree, nodePathFromPath } from "sourcegraph/workbench/fileTreeModel";

interface Props {
	files: GQL.IFile[];
	repo: string;
	rev: string | null;
	path: string;
}

const DownChevron = `<svg viewBox="0 0 24 24" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"><path d="M12.5240625,16.27845 C12.2315625,16.571325 11.7568125,16.5717 11.4639375,16.278825 C11.4635625,16.278825 11.4631875,16.278825 11.4631875,16.27845 L5.8366875,10.64895 C5.5396875,10.359825 5.5329375,9.8847 5.8220625,9.5877 C6.1111875,9.291075 6.5863125,9.2847 6.8833125,9.57345 L6.8975625,9.5877 L11.9934375,14.686575 L17.0896875,9.5877 C17.3788125,9.290325 17.8535625,9.28395 18.1509375,9.5727 C18.4483125,9.86145 18.4546875,10.336575 18.1659375,10.633575 C18.1610625,10.638825 18.1561875,10.644075 18.1509375,10.64895 L12.5240625,16.27845 Z" id="Fill" fill="white"></path></svg>`;

const RightChevron = `<svg viewBox="0 0 24 24" version="1.1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"><path d="M10.656375,18.1633125 C10.3635,18.4569375 9.888375,18.4569375 9.59475,18.1640625 C9.3015,17.8711875 9.3015,17.3956875 9.594375,17.1028125 L14.696625,12.0065625 L9.594375,6.9099375 C9.3015,6.6166875 9.3015,6.1415625 9.59475,5.8486875 C9.888375,5.5554375 10.3635,5.5558125 10.656375,5.8494375 L16.289625,11.4759375 C16.582875,11.7684375 16.583625,12.2428125 16.291125,12.5356875 C16.290375,12.5360625 16.290375,12.5364375 16.289625,12.5368125 L10.656375,18.1633125 Z" id="Fill" fill="white"></path></svg>`;

export class FileTree extends React.Component<Props, {}> {

	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	private _treeImpl: ITree;
	context: { router: InjectedRouter };

	constructor() {
		super();
		this.divMounted = this.divMounted.bind(this);
	}

	componentWillUnmount(): void {
		if (this._treeImpl) {
			this._treeImpl.dispose();
		}
	}

	divMounted(domElement: HTMLElement): void {
		if (!domElement) { return; }

		const config = {
			dataSource: new FileTreeDataSource(),
			controller: new Controller(this.selectElement.bind(this)),
			renderer: new Renderer(),
		};
		this._treeImpl = new Tree(domElement, config);
		this.setTreeInput(this.props.files);
	}

	selectElement(node: Node): boolean {
		if (node.children.length !== 0 || node.label === "") {
			return false;
		}
		Events.FileTree_Navigated.logEvent({
			repo: this.props.repo,
			file: node.id,
			rev: this.props.rev,
		});
		const url = urlToBlob(this.props.repo, this.props.rev, node.id);
		this.context.router.push(url);
		return true;
	}

	componentWillReceiveProps(nextProps: Props): void {
		if (this.props.repo === nextProps.repo) { return; }
		this.setTreeInput(nextProps.files);
	}

	setTreeInput(files: GQL.IFile[]): void {
		if (!this._treeImpl) { return; }
		const data = makeTree(files);
		this._treeImpl.setInput(data);

		// Set and expand the selection
		const nodePath = nodePathFromPath(data, this.props.path);
		this._treeImpl.expandAll(nodePath);
		this._treeImpl.setSelection([nodePath[nodePath.length - 1]]);
	}

	render(): JSX.Element {
		// This weird styling is necessary because we need to use real
		// selectors to override the default Monaco CSS.
		const style = Object.assign(
			select(` .monaco-tree.focused .monaco-tree-rows > .monaco-tree-row.focused:not(.highlighted),
					.monaco-tree.focused .monaco-tree-rows > .monaco-tree-row.selected:not(.highlighted),
					.monaco-tree.focused .monaco-tree-rows > .monaco-tree-row.focused.selected:not(.highlighted),
					.monaco-tree .monaco-tree-rows > .monaco-tree-row.focused.selected,
					.monaco-tree .monaco-tree-rows > .monaco-tree-row.selected:not(.highlighted),
					.monaco-tree .monaco-tree-rows > .monaco-tree-row:hover:not(.highlighted):not(.selected):not(.focused)`, {
					background: colors.blueText(),
				}),
			select(" .monaco-tree div", {
				color: colors.coolGray4(),
			}),
			select(" .monaco-tree", {
				background: colors.coolGray2(),
			}),
			select(" .monaco-tree .monaco-tree-row > .content", {
				lineHeight: "30px",
			}),
			select(" .monaco-tree .monaco-tree-rows > .monaco-tree-row.has-children > .content:before", {
				backgroundImage: `url('data:image/svg+xml, ${RightChevron}')`,
			}),
			select(" .monaco-tree .monaco-tree-rows > .monaco-tree-row.expanded > .content:before", {
				backgroundImage: `url('data:image/svg+xml, ${DownChevron}')`,
			}),
		);
		return <div>
			<Title repo={this.props.repo} />
			<div {...style} ref={this.divMounted} style={{
				minWidth: 300,
				height: "calc(100% - 50px)",
			}}>
			</div>
		</div>;
	}
}

function Title({repo}: { repo: string }): JSX.Element {
	return <Heading level={5} style={{
		boxShadow: `rgba(0, 0, 0, 0.4) 0px 2px 6px 0px`,
		zIndex: 1,
		background: colors.coolGray2(),
		position: "relative",
		margin: 0,
		padding: "10px 20px 10px",
	}}>
		<Link to={urlTo("repo", { splat: repo })}
			{...hover({ color: `${colors.white()} !important` }) }
			style={{ color: colors.coolGray4() }}
			>
			{repo.replace(/^github.com\//, "")}
		</Link>
	</Heading >;
}
