import * as React from "react";
import URI from "vs/base/common/uri";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { IEditorService } from "vs/platform/editor/common/editor";

import { getEditorInstance } from "sourcegraph/editor/Editor";
import { infoStore } from "sourcegraph/workbench/info/sidebar";

import { Disposables } from "sourcegraph/workbench/utils";
import { Location } from "vs/editor/common/modes";

import * as autobind from "autobind-decorator";
import { $, Builder } from "vs/base/browser/builder";
import * as strings from "vs/base/common/strings";
import { Tree } from "vs/base/parts/tree/browser/treeImpl";
import { Controller } from "vs/editor/contrib/referenceSearch/browser/referencesWidget";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";

import { Services } from "sourcegraph/workbench/services";

import { LegacyRenderer } from "vs/base/parts/tree/browser/treeDefaults";
import { IWorkspaceContextService } from "vs/platform/workspace/common/workspace";

import * as dom from "vs/base/browser/dom";
import { IElementCallback, ITree } from "vs/base/parts/tree/browser/tree";

import { FileReferences, OneReference, ReferencesModel } from "sourcegraph/workbench/info/referencesModel";
import { DataSource } from "sourcegraph/workbench/info/referencesWidget";
import { WorkspaceBadge } from "sourcegraph/workbench/ui/badges/workspaceBadge";
import { FileLabel } from "sourcegraph/workbench/ui/fileLabel";
import { LeftRightWidget } from "sourcegraph/workbench/ui/leftRightWidget";
import { scrollToLine } from "sourcegraph/workbench/utils";

interface Props {
	model?: ReferencesModel;
	focus(resource: Location): void;
}

interface State {
	previewResource: Location | null;
}

@autobind
export class RefTree extends React.Component<Props, State> {

	private tree: Tree;
	private toDispose: Disposables = new Disposables();

	state: State = {
		previewResource: null,
	};

	componentWillUnmount(): void {
		this.toDispose.dispose();
	}

	private scrollEditorForRef(): void {
		const editor = getEditorInstance();
		const line = editor.getSelection().startLineNumber - 5;
		scrollToLine(editor, line);
	}

	private treeItemFocused(reference: FileReferences | OneReference): void {
		if (!(reference instanceof OneReference)) {
			return;
		}

		const modelService = Services.get(ITextModelResolverService);
		modelService.createModelReference(reference.uri).then((ref) => {
			this.scrollEditorForRef();
			this.props.focus(reference);
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
		if (this.tree && this.tree.getInput() !== model) {
			this.tree.setInput(model);
		}
	}

	render(): JSX.Element {
		if (!this.props.model) {
			return <div></div>;
		}

		this.updateTree(this.props.model);
		return <div style={{
			zIndex: 1,
			flex: "1 1 100%",
			height: "100%", // Necessary for Safari to render the child element at 100% height.
		}} ref={this.treeDiv}>

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
				return 100;
			}
			return 71;
		} else if (element instanceof FileReferences) {
			return 50;
		}

		return 0;
	}

	protected render(tree: ITree, element: FileReferences | OneReference, container: HTMLElement): IElementCallback | any {
		dom.clearNode(container);

		if (element instanceof FileReferences || element instanceof FileReferences) {
			const fileReferencesContainer: Builder = $(".reference-file");
			// tslint:disable
			let workspaceURI = URI.from({
				scheme: element.uri.scheme,
				authority: element.uri.authority,
				path: element.uri.path,
				query: element.uri.query,
				fragment: element.uri.path,
			});
			new LeftRightWidget(fileReferencesContainer, (left: HTMLElement) => {
				new FileLabel(left, workspaceURI, this._contextService);
				return null as any;

			}, (right: HTMLElement) => {

				const workspace = workspaceURI.path === this._editorURI.path ? "Local" : "External";
				const badge = new WorkspaceBadge(right, workspace);

				if (element.failure) {
					badge.setTitleFormat("Failed to resolve file.");
				} else if (workspace === "Local") {
					badge.setTitleFormat("Local");
				} else {
					badge.setTitleFormat("External");
				}

				return badge as any;
			});

			fileReferencesContainer.appendTo(container);
		}

		if (element instanceof OneReference) {
			const preview = element.preview.preview(element.range);

			let authorInfo = "";
			if (element.commitInfo && element.commitInfo.hunk.author && element.commitInfo.hunk.author.person) {
				let imgURL = "https://secure.gravatar.com/avatar?d=mm&f=y&s=128";
				let gravatarHash = element.commitInfo.hunk.author.person.gravatarHash;
				if (gravatarHash) {
					imgURL = `https://secure.gravatar.com/avatar/${gravatarHash}?s=128&d=retro`;
				}
				authorInfo = strings.format(
					`<div class="author-details">
						<img src="${imgURL}" />
						<div class="name">{0} {1}</div>
					</div>`,
					element.commitInfo.hunk.author.person.name,
					element.commitInfo.hunk.author.date,
				);
			}

			$(".sidebar-references").innerHtml(
				strings.format(
					`<div class="code-content">
							<div class="function">
								<code>{0}</code><code>{1}</code><code>{2}</code>
								{3}
								<div class="file-details">{4} - Line: {5}
								</div>
							</div>
							<div class="divider-container">
								<div class="divider"/>
							</div>
						</div>`,
					strings.escape(preview.before),
					strings.escape(preview.inside),
					strings.escape(preview.after),
					authorInfo,
					element.uri.fragment,
					element.range.startLineNumber)).appendTo(container);
		}

		return null;
	}
}
