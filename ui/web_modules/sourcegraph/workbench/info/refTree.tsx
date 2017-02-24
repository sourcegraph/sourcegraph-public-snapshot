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
	private focusedReference: FileReferences | OneReference | undefined = undefined;
	private updatingTree: boolean = false;
	private visibleModel: ReferencesModel | undefined = undefined;

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
		this.refreshTreeIfModelChanged(nextProps.model);
		return false;
	}

	private refreshTreeIfModelChanged(model: ReferencesModel): void {
		if (this.updatingTree) {
			return;
		}
		if (model === this.visibleModel) {
			return;
		}
		this.updatingTree = true;

		// Pre-resolve all models before we set the model on the tree.
		// If we don't do this, the tree will appear blank as files are resolved (can be slow).
		const modelService = Services.get(ITextModelResolverService);
		Promise.all(model.groups.map(fileReferences => {
			return fileReferences.resolve(modelService);
		})).then(() => {
			let elementsToExpand: Array<FileReferences> = [];

			if (model.groups.length && !firstToggleAdded) {
				userToggledState.add(model.workspace.toString());
				firstToggleAdded = true;
			}
			for (let fileReferences of model.groups) {
				const workspace = fileReferences.uri.with({ fragment: "" }).toString();
				if (userToggledState.has(workspace)) {
					elementsToExpand.push(fileReferences);
				}
			}

			// Maintain the user's current scroll position.
			let revealReference: FileReferences | OneReference | undefined = undefined;
			let relativeTop: number | undefined = undefined;
			const scrollPosition = this.tree.getScrollPosition();
			if (this.visibleModel && this.visibleModel.references.length > 0 && scrollPosition > 0) {
				// Tree gives us access to a scroll position, which is a % of scroll (0=top, 1=bottom)
				// We want to map this to an actual reverence that we can reveal after the re-render to maintain scroll position.
				// As an approximation we collect all expanded elements and pretend that they have the same height
				// to find the approximate index of the item that has been scrolled to.
				// Our approximation only needs to be good enought to make sure the item at this index is actually visible.
				// After render, we reveal this item while maintaining its relative top so that it doesn't appear to move.
				// We may have to be more precise about this when external references stream in.
				const expandedElements: FileReferences[] = this.tree.getExpandedElements();
				const expandedReferences = expandedElements.reduce<(FileReferences | OneReference)[]>((collected, fileRef) => {
					collected.push(fileRef);
					collected.push(...fileRef.children);
					return collected;
				}, []);

				const visibleScrollIndex = Math.floor((expandedReferences.length - 1) * scrollPosition);
				revealReference = expandedReferences[visibleScrollIndex];
				relativeTop = this.tree.getRelativeTop(revealReference);
			}

			this.tree.setInput(model).then(() => {
				this.tree.expandAll(elementsToExpand).then(() => {
					if (this.focusedReference) {
						this.tree.setSelection([this.focusedReference]);
						this.tree.setFocus(this.focusedReference);
					}
					if (revealReference) {
						this.tree.reveal(revealReference, relativeTop);
					}
					this.tree.layout();
					this.updatingTree = false;
					this.visibleModel = model;

					// check to see if a new model was set while we were refreshing.
					this.refreshTreeIfModelChanged(this.props.model);
				});
			});
		});
	}

	private treeItemFocused(reference: FileReferences | OneReference): void {
		this.focusedReference = reference;
		if (!(reference instanceof OneReference)) {
			return;
		}

		const modelService = Services.get(ITextModelResolverService);
		modelService.createModelReference(reference.uri).then((ref) => {
			this.props.focus(reference);
			this.tree.layout();
		});

		const editor = getEditorInstance();
		if (!editor) {
			throw new Error("Editor not set.");
		}
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
			controller: new Controller({}, { ref: undefined })
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
		if (editor && editor.getModel()) {
			this._editorURI = editor.getModel().uri;
		}
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

		if (element instanceof FileReferences && this._editorURI) {
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
