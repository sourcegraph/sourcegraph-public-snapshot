import * as React from "react";
import URI from "vs/base/common/uri";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";
import { IEditorService } from "vs/platform/editor/common/editor";

import { getEditorInstance } from "sourcegraph/editor/Editor";
import { infoStore } from "sourcegraph/workbench/info/sidebar";

import { Disposables } from "sourcegraph/workbench/utils";
import { Location } from "vs/editor/common/modes";

import * as autobind from "autobind-decorator";
import { $ } from "vs/base/browser/builder";
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
		if (this.tree) {
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
		}} ref={this.treeDiv}>

		</div>;
	}
}

class Renderer extends LegacyRenderer {
	private _contextService: IWorkspaceContextService;

	constructor(
		@IWorkspaceContextService contextService: IWorkspaceContextService
	) {
		super();
		this._contextService = contextService;
	}

	public getHeight(tree: ITree, element: any): number {
		if (element instanceof OneReference) {
			return 100;
		}
		return 0;
	}

	// NOTE: This is a HUGE todo and will be cleaned up ASAP. Letting this sit here for now during code review, but no need to add comments
	// around this function. Will provide an update once this is refactored and cleaned up. - Kingy (1/17/2017)
	protected render(tree: ITree, element: FileReferences | OneReference, container: HTMLElement): IElementCallback | any {
		dom.clearNode(container);

		if (element instanceof OneReference) {
			const preview = element.parent.preview.preview(element.range);
			if (element.info && element.info.hunk.author && element.info.hunk.author.person) {
				let imgURL = "https://secure.gravatar.com/avatar?d=mm&f=y&s=128";
				let gravatarHash = element.info.hunk.author.person.gravatarHash;
				if (gravatarHash) {
					imgURL = `https://secure.gravatar.com/avatar/${gravatarHash}?s=128&d=retro`;
				}
				$(".sidebar-references").innerHtml(
					strings.format(
						`<div class="code-content">
							<div class="function">
								<code>
									{0}
								</code>
								<code>
									{1}
								</code>
								<code>
									{2}
								</code>
								<div class="author-details">
									<img src="${imgURL}" />
									<div class="name">{3} {4}</div>
								</div>
								<div class="file-details">
									{5}: Line
									{6}
								</div>
							</div>
							<div class="divider-container">
								<div class="divider"/>
							</div>
						</div>`,
						strings.escape(preview.before),
						strings.escape(preview.inside),
						strings.escape(preview.after),
						element.info.hunk.author.person.name,
						element.info.hunk.author.date,
						element.info.loc.uri.fragment,
						element.info.loc.range.startLineNumber)).appendTo(container);
			} else {
				$(".sidebar-references").innerHtml(
					strings.format(
						`<div class="code-content">
							<div class="function">
								<span>{0}</span><span class="referenceMatch">{1}</span><span>{2}</span>
							</div>
							<div class="divider-container">
								<div class="divider"/>
							</div>
						</div>`,
						strings.escape(preview.before),
						strings.escape(preview.inside),
						strings.escape(preview.after))).appendTo(container);
			}
		}

		return null;
	}
}
