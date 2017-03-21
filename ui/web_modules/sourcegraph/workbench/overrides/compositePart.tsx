import * as React from "react";
import * as ReactDOM from "react-dom";

import { Builder } from "vs/base/browser/builder";
import { Composite } from "vs/workbench/browser/composite";
import { ICompositeTitleLabel } from "vs/workbench/browser/parts/compositePart";
import * as vs from "vscode/src/vs/workbench/browser/parts/compositePart";

import { ExplorerTitle } from "sourcegraph/workbench/ExplorerTitle";
import { RouterContext } from "sourcegraph/workbench/utils";

export class CompositePart<T extends Composite> extends vs.CompositePart<T> {

	private domElement: HTMLElement;

	private renderTitle = () => {
		ReactDOM.render(<RouterContext>
			<ExplorerTitle />
		</RouterContext>, this.domElement);
	}

	protected createTitleLabel(parent: Builder): ICompositeTitleLabel {
		this.domElement = parent.getHTMLElement();
		return {
			updateTitle: this.renderTitle,
		};
	}

}
