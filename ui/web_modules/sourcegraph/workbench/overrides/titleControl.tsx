import * as React from "react";
import * as ReactDOM from "react-dom";
import { TitleControl } from "vs/workbench/browser/parts/editor/titleControl";
import { getResource } from "vs/workbench/common/editor";

import { URIUtils } from "sourcegraph/core/uri";
import { EditorTitle } from "sourcegraph/editor/EditorTitle";

export class NoTabsTitleControl extends TitleControl {
	domElement: HTMLElement;

	create(parent: HTMLElement): void {
		this.domElement = parent;
		this.render();
	}

	doRefresh(): void {
		this.render();
	}

	render(): void {
		if (!this.context) {
			return;
		}
		const editor = this.context && this.context.activeEditor;
		const resource = getResource(editor) || (editor as any).resource;
		const pathspec = URIUtils.repoParams(resource);
		const component = <EditorTitle pathspec={pathspec} />;
		ReactDOM.render(component, this.domElement);
		this.domElement.style.height = "auto";
	}

	dispose(): void {
		ReactDOM.unmountComponentAtNode(this.domElement);
	}
}
