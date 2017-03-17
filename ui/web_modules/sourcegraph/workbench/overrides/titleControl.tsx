import * as React from "react";
import * as ReactDOM from "react-dom";

import { TitleControl } from "vs/workbench/browser/parts/editor/titleControl";
import { toResource } from "vs/workbench/common/editor";

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
		try {
			// TODO(john): saw this error at extracting .details once when doing a jump-to-repo via quickopen.
			// This code is super gross and we need a better way...
			const resource = toResource(editor) || ((editor as any).details && (editor as any).details.resource) || (editor as any).resource;
			const pathspec = URIUtils.repoParams(resource);
			const component = <EditorTitle pathspec={pathspec} />;
			ReactDOM.render(component, this.domElement);
			this.domElement.style.height = "auto";
		} catch (e) {
			// swallow
		}
	}

	dispose(): void {
		ReactDOM.unmountComponentAtNode(this.domElement);
	}
}
