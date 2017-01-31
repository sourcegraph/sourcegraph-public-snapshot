import * as autobind from "autobind-decorator";
import { insertGlobal } from "glamor";
import * as debounce from "lodash/debounce";
import * as React from "react";

import * as dom from "vs/base/browser/dom";
import { IKeyboardEvent } from "vs/base/browser/keyboardEvent";
import URI from "vs/base/common/uri";
import { IElementCallback, ITree } from "vs/base/parts/tree/browser/tree";
import { LegacyRenderer } from "vs/base/parts/tree/browser/treeDefaults";
import { Tree } from "vs/base/parts/tree/browser/treeImpl";
import { Location } from "vs/editor/common/modes";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { Controller as VSController } from "vs/editor/contrib/referenceSearch/browser/referencesWidget";
import { IEditorService } from "vs/platform/editor/common/editor";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";

import { getEditorInstance } from "sourcegraph/editor/Editor";
import { ReferenceItem } from "sourcegraph/workbench/info/referenceItem";
import { FileReferences, OneReference, ReferencesModel } from "sourcegraph/workbench/info/referencesModel";
import { DataSource } from "sourcegraph/workbench/info/referencesWidget";
import { RepositoryHeader } from "sourcegraph/workbench/info/repositoryHeader";
import { Services } from "sourcegraph/workbench/services";
import { Disposables, scrollToLine } from "sourcegraph/workbench/utils";

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
const fileRefsHeight = 36;
const refBaseHeight = 68;
const refWithCommitInfoHeight = 95;

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

		const editor = getEditorInstance();
		scrollToLine(editor, editor.getSelection().startLineNumber);
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
			RepositoryHeader(
				element,
				container,
				userToggledState,
				firstToggleAdded,
				fileRefsHeight,
				this._contextService,
				this._editorURI.path,
			);
		} else if (element instanceof OneReference) {
			ReferenceItem(
				element,
				container,
				firstToggleAdded,
				refBaseHeight,
				refWithCommitInfoHeight,
			);
		}

		return null;
	}
}

/**
 * Controller extends the default Controller. It adds the functionality of
 *  pressing enter to jump to the selected reference. It avoids the behavior of
 *  the default controller where double clicking an entry will jump to the
 *  definition.
 */
class Controller extends VSController {
	onEnter(tree: Tree, event: IKeyboardEvent): boolean {
		const selections = tree.getSelection();
		const reference = selections[0];
		if (selections.length > 1 || !(reference instanceof OneReference)) {
			return false;
		}
		const editorService = Services.get(IEditorService) as IEditorService;
		editorService.openEditor({
			resource: reference.uri,
			options: {
				selection: reference.range,
			},
		});
		return true;
	}
}
