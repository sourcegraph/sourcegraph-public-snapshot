import * as React from "react";
import * as ReactDOM from "react-dom";
import { TitleControl } from "vs/workbench/browser/parts/editor/titleControl";

import { layout } from "sourcegraph/components/utils";
import { URIUtils } from "sourcegraph/core/uri";
import { EditorTitle } from "sourcegraph/editor/EditorTitle";
import { getResource } from "sourcegraph/workbench/utils";

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
		const resource = getResource(editor);
		const pathspec = URIUtils.repoParams(resource);
		const component = <EditorTitle pathspec={pathspec} />;
		ReactDOM.render(component, this.domElement);
		this.domElement.style.height = `${layout.editorToolbarHeight}px`;
	}

	dispose(): void {
		//
	}
}
