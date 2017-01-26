import { css, insertGlobal } from "glamor";
import * as debounce from "lodash/debounce";
import * as React from "react";
import * as ReactDOM from "react-dom";
import URI from "vs/base/common/uri";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { IEditorService } from "vs/platform/editor/common/editor";

import { getEditorInstance } from "sourcegraph/editor/Editor";
import { infoStore } from "sourcegraph/workbench/info/sidebar";

import { Disposables } from "sourcegraph/workbench/utils";
import { Builder } from "vs/base/browser/builder";
import { IDisposable } from "vs/base/common/lifecycle";
import { Location } from "vs/editor/common/modes";

import * as autobind from "autobind-decorator";
import { $ } from "vs/base/browser/builder";
import { Tree } from "vs/base/parts/tree/browser/treeImpl";
import { Controller } from "vs/editor/contrib/referenceSearch/browser/referencesWidget";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";

import { Services } from "sourcegraph/workbench/services";

import { LegacyRenderer } from "vs/base/parts/tree/browser/treeDefaults";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";

import * as dom from "vs/base/browser/dom";
import { IElementCallback, ITree } from "vs/base/parts/tree/browser/tree";

import { ReferenceCard } from "sourcegraph/components";
import { List } from "sourcegraph/components/symbols/Primaries";
import { colors, paddingMargin, whitespace } from "sourcegraph/components/utils";
import { FileReferences, OneReference, ReferencesModel } from "sourcegraph/workbench/info/referencesModel";
import { DataSource } from "sourcegraph/workbench/info/referencesWidget";
import { WorkspaceBadge } from "sourcegraph/workbench/ui/badges/workspaceBadge";
import { FileLabel } from "sourcegraph/workbench/ui/fileLabel";
import { LeftRightWidget } from "sourcegraph/workbench/ui/leftRightWidget";

import "sourcegraph/workbench/styles/tree.css";

interface Props {
	model: ReferencesModel;
	focus(resource: Location): void;
}

interface State {
	previewResource: Location | null;
}

// This set is used to keep track of which rows in the tree the user has
// expanded or closed. Due to race conditions caused by the loading state of
// the tree in VSCode, this hack is necessary to properly expand tree elements
// during updates.
let userToggledState: Set<string>;
let firstToggleAdded: boolean;

// The height of the tree elements must be explicitly defined and enforced,
// otherwise scrolling functionality will not work properly.
const fileRefsHeight: number = 36;
const refBaseHeight: number = 68;
const refWithCommitInfoHeight: number = 95;

@autobind
export class RefTree extends React.Component<Props, State> {

	private tree: Tree;
	private toDispose: Disposables = new Disposables();
	private treeID: string = "reference-tree";

	state: State = {
		previewResource: null,
	};

	constructor() {
		super();
		userToggledState = new Set<string>();
		firstToggleAdded = false;
	}

	componentDidMount(): void {
		this.resetMonacoStyles();
		this.onResize = debounce(this.onResize, 200);
		window.addEventListener("resize", this.onResize, false);
	}

	componentDidUpdate(): void {
		this.updateTree(this.props.model);
	}

	componentWillUnmount(): void {
		this.toDispose.dispose();
		window.removeEventListener("resize", this.onResize);
	}

	onResize(): void {
		this.tree.layout();
	}

	shouldComponentUpdate(nextProps: Props, nextState: State): boolean {
		let expandedElements = new Array<FileReferences>();
		const scrollPosition = this.tree.getScrollPosition();

		if (nextProps.model !== this.props.model) {
			if (nextProps.model && nextProps.model.groups.length && !firstToggleAdded) {
				userToggledState.add(nextProps.model.groups[0].uri.toString());
				firstToggleAdded = true;
			}
			if (nextProps.model) {
				for (let r of nextProps.model.groups) {
					if (userToggledState.has(r.uri.toString())) {
						expandedElements.push(r);
					}
				}
			}
			this.tree.setInput(nextProps.model).then(() => {
				this.tree.expandAll(expandedElements).then(() => {
					this.tree.setScrollPosition(scrollPosition);
					this.tree.layout();
				});
			});
		}
		return false;
	}

	private treeItemFocused(reference: FileReferences | OneReference): void {
		if (!(reference instanceof OneReference)) {
			return;
		}

		const modelService = Services.get(ITextModelResolverService);
		modelService.createModelReference(reference.uri).then((ref) => {
			this.props.focus(reference);
			this.tree.layout();
		});
	}

	private treeItemSelected(resource: URI): void {
		const editorService = Services.get(IEditorService) as IEditorService;
		editorService.openEditor({ resource });
		infoStore.dispatch(null);
	}

	private treeDiv(parent: HTMLDivElement): void {
		if (!parent) {
			return;
		}

		const instantiationService = Services.get(IInstantiationService);
		const config = {
			dataSource: instantiationService.createInstance(DataSource),
			renderer: instantiationService.createInstance(Renderer),
			controller: new Controller()
		};

		const options = {
			allowHorizontalScroll: false,
			twistiePixels: 20,
		};

		this.tree = new Tree(parent, config, options);

		this.toDispose.add(this.tree);
		this.toDispose.add(this.tree.addListener2(Controller.Events.SELECTED, (ref) => {
			if (ref instanceof OneReference) {
				this.treeItemSelected(ref.uri);
			}
		}));
		this.toDispose.add(this.tree.addListener2(Controller.Events.FOCUSED, this.treeItemFocused));
		this.forceUpdate();
	}

	private updateTree(model: ReferencesModel): void {
		if (this.tree) {
			this.tree.layout();
			if (this.tree.getInput() !== model) {
				this.tree.setInput(model);
			}
		}
	}

	private resetMonacoStyles(): void {
		insertGlobal(`#${this.treeID} .monaco-tree-row`, {
			backgroundColor: "initial",
			height: "initial !important",
			paddingLeft: "initial !important",
			overflow: "visible",
		});
		insertGlobal(`#${this.treeID} .monaco-tree:focus`, { outline: "none" });
		insertGlobal(`#${this.treeID} .monaco-scrollable-element`, { position: "absolute !important", top: 0, left: 0, bottom: 0, right: 0 });
		insertGlobal(`#${this.treeID} .monaco-tree-row .content:before`, { display: "none" });
		insertGlobal(`#${this.treeID} .monaco-tree-row.selected`, { backgroundColor: "initial" });
		insertGlobal(`#${this.treeID} .monaco-tree-row:hover`, { backgroundColor: "initial" });
	}

	render(): JSX.Element {
		return <div ref={this.treeDiv} id={this.treeID} style={{
			zIndex: 1,
			flex: "1 1 100%",
			display: "flex",
			flexDirection: "column",
			outline: "none",
		}}>
		</div>;
	}
}

class Renderer extends LegacyRenderer {
	private _contextService: IWorkspaceContextService;
	private _editorURI: URI;

	constructor(
		@IWorkspaceContextService contextService: IWorkspaceContextService
	) {
		super();
		this._contextService = contextService;
		const editor = getEditorInstance();
		this._editorURI = editor.getModel().uri;
	}

	public getHeight(tree: ITree, element: any): number {
		if (element instanceof OneReference) {
			if (element.commitInfo) {
				return refWithCommitInfoHeight;
			}
			return refBaseHeight;
		} else if (element instanceof FileReferences) {
			return fileRefsHeight;
		}

		return 0;
	}

	protected render(tree: ITree, element: FileReferences | OneReference, container: HTMLElement): IElementCallback | any {
		dom.clearNode(container);

		if (element instanceof FileReferences) {
			const repositoryHeader = $(".refs-repository-title");
			let workspaceURI = URI.from({
				scheme: element.uri.scheme,
				authority: element.uri.authority,
				path: element.uri.path,
				query: element.uri.query,
				fragment: element.uri.path,
			});

			// tslint:disable:no-new
			new LeftRightWidget(repositoryHeader, left => {
				const repoTitleContent = new FileLabel(left, workspaceURI, this._contextService);
				repoTitleContent.setIcon(<List width={18} style={{ marginLeft: -2, color: colors.blueGrayL1() }} />);
				return null as any;
			}, right => {

				const workspace = workspaceURI.path === this._editorURI.path ? "Local" : "External";
				const badge = new WorkspaceBadge(right, workspace);

				if (element.failure) {
					badge.setTitleFormat("Failed to resolve file.");
				} else if (workspace === "Local") {
					badge.setTitleFormat("Local");
					badge.setColor(colors.green());
				} else {
					badge.setTitleFormat("External");
					badge.setColor(colors.orangeL1());
				}

				return badge as any;
			}).setClassNames((css as any)({
				paddingLeft: whitespace[2],
				paddingRight: whitespace[2],
			}));

			const borderSx = `1px solid ${colors.blueGrayL1(0.2)}`;
			const refHeaderSx = (css as any)(
				paddingMargin.padding("x", 2),
				{
					backgroundColor: colors.blueGrayL3(),
					borderBottom: borderSx,
					borderTop: borderSx,
					boxShadow: `0 2px 2px 0 ${colors.black(0.05)}`,
					color: colors.text(),
					display: "flex",
					fontWeight: "bold",
					alignItems: "center",
					height: fileRefsHeight,
				},
			);

			insertGlobal(`.monaco-tree-row.has-children .${refHeaderSx}:before`, {
				content: `""`,
				height: 15,
				width: 9,
				marginLeft: 9,
				transition: "all 300ms ease-in-out",
				backgroundRepeat: "no-repeat",
				backgroundImage: "url(data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNyIgaGVpZ2h0PSIxMiIgdmlld0JveD0iMjQgMTUgNyAxMiIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cGF0aCBkPSJNMjUuNDcyIDI2LjE0NWMtLjI2LjI2LS42ODIuMjYtLjk0MyAwLS4yNjItLjI2LS4yNjItLjY4Mi0uMDAyLS45NDNsNC41MzYtNC41My00LjUzNi00LjUzYy0uMjYtLjI2LS4yNi0uNjgzIDAtLjk0My4yNjItLjI2Mi42ODQtLjI2Ljk0NCAwbDUuMDA4IDVjLjEyNS4xMjYuMTk2LjI5NS4xOTYuNDcycy0uMDcuMzQ3LS4xOTYuNDcybC01LjAwOCA1eiIgZmlsbD0iIzc3OTNBRSIgZmlsbC1ydWxlPSJldmVub2RkIi8+PC9zdmc+)",
			});

			insertGlobal(`.monaco-tree-row.has-children.expanded .${refHeaderSx}:before`, {
				transform: "rotate(90deg)",
			});

			repositoryHeader.getHTMLElement().classList.add(refHeaderSx);
			repositoryHeader.on("click", (e: Event, builder: Builder, unbind: IDisposable): void => {
				const stateKey = element.uri.toString();
				if (userToggledState.has(stateKey)) {
					userToggledState.delete(stateKey);
				} else {
					userToggledState.add(stateKey);
				}
			});
			repositoryHeader.appendTo(container);
		}

		if (element instanceof OneReference) {
			const preview = element.preview.preview(element.range);
			const fileName = element.uri.fragment;
			const line = element.range.startLineNumber;
			const fnSignature = preview.before.concat(preview.inside, preview.after);
			const refContainer = $("div");
			let height = refBaseHeight;
			let defaultAvatar;
			let gravatarHash;
			let avatar;
			let authorName;
			let date;

			if (element && element.commitInfo && element.commitInfo.hunk.author && element.commitInfo.hunk.author.person) {
				defaultAvatar = "https://secure.gravatar.com/avatar?d=mm&f=y&s=128";
				gravatarHash = element.commitInfo.hunk.author.person.gravatarHash;
				avatar = gravatarHash ? `https://secure.gravatar.com/avatar/${gravatarHash}?s=128&d=retro` : defaultAvatar;
				authorName = element.commitInfo.hunk.author.person.name;
				date = element.commitInfo.hunk.author.date;
				height = refWithCommitInfoHeight;
			}

			refContainer.appendTo(container);

			ReactDOM.render(
				<ReferenceCard
					fnSignature={fnSignature}
					authorName={authorName}
					avatar={avatar}
					date={date}
					fileName={fileName}
					height={height}
					line={line} />,
				refContainer.getHTMLElement(),
			);
		}

		return null;
	}
}
