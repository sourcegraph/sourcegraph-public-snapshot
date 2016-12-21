import * as React from "react";
import * as ReactDOM from "react-dom";
import { TitleControl } from "vs/workbench/browser/parts/editor/titleControl";

import { BlobTitle } from "sourcegraph/blob/BlobTitle";
import { layout } from "sourcegraph/components/utils";
import { URIUtils } from "sourcegraph/core/uri";
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
		const {repo, rev, path} = URIUtils.repoParams(resource);
		ReactDOM.render(<BlobTitle
			repo={repo}
			rev={rev}
			path={path}

			// TODO:
			routeParams={{ splat: "" }}
			toggleAuthors={() => {/* */ } }
			routes={[]}
			toast=""
			/>, this.domElement);
		this.domElement.style.height = `${layout.editorToolbarHeight}px`;
	}

	dispose(): void {
		//
	}
}
