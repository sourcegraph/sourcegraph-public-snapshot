"use strict";

import * as dom from "vs/base/browser/dom";
import { HighlightedLabel } from "vs/base/browser/ui/highlightedlabel/highlightedLabel";
import { IMatch } from "vs/base/common/filters";
import { IWorkspaceProvider, getPathLabel } from "vs/base/common/labels";
import * as paths from "vs/base/common/paths";
import * as types from "vs/base/common/types";
import uri from "vs/base/common/uri";

export interface IIconLabelCreationOptions {
	supportHighlights?: boolean;
}

export interface IIconLabelOptions {
	title?: string;
	extraClasses?: string[];
	italic?: boolean;
	matches?: IMatch[];
}

export class IconLabel {
	private domNode: HTMLElement;
	private labelNode: HTMLElement | HighlightedLabel;
	private descriptionNode: HTMLElement;

	constructor(container: HTMLElement, options?: IIconLabelCreationOptions) {
		this.domNode = dom.append(container, dom.$(".monaco-icon-label"));
		if (options && options.supportHighlights) {
			this.labelNode = new HighlightedLabel(dom.append(this.domNode, dom.$("a.label-name")));
		} else {
			this.labelNode = dom.append(this.domNode, dom.$("a.label-name"));
		}
		this.descriptionNode = dom.append(this.domNode, dom.$("span.label-description"));
	}

	public get element(): HTMLElement {
		return this.domNode;
	}

	public get labelElement(): HTMLElement {
		const labelNode = this.labelNode;
		if (labelNode instanceof HighlightedLabel) {
			return labelNode.element;
		} else {
			return labelNode;
		}
	}

	public get descriptionElement(): HTMLElement {
		return this.descriptionNode;
	}

	public setValue(label?: string, description?: string, options?: IIconLabelOptions): void {
		const labelNode = this.labelNode;
		if (labelNode instanceof HighlightedLabel) {
			labelNode.set(label || "", options ? options.matches : void 0);
		} else {
			labelNode.textContent = label || "";
		}

		this.descriptionNode.textContent = description || "";

		if (!description) {
			dom.addClass(this.descriptionNode, "empty");
		} else {
			dom.removeClass(this.descriptionNode, "empty");
		}

		this.domNode.title = options && options.title ? options.title : "";

		const classes = ["monaco-icon-label"];
		if (options) {
			if (options.extraClasses) {
				classes.push(...options.extraClasses);
			}

			if (options.italic) {
				classes.push("italic");
			}
		}

		this.domNode.className = classes.join(" ");
	}

	public dispose(): void {
		const labelNode = this.labelNode;
		if (labelNode instanceof HighlightedLabel) {
			labelNode.dispose();
		}
	}
}

export class FileLabel extends IconLabel {

	constructor(container: HTMLElement, file: uri, provider: IWorkspaceProvider) {
		super(container);

		this.setFile(file, provider);
	}

	public setFile(file: uri, provider: IWorkspaceProvider): void {
		const path = getPath(file);
		const parent = paths.dirname(path);

		this.setValue(paths.basename(path), parent && parent !== "." ? getPathLabel(parent, provider) : "", { title: path });
	}
}

function getPath(arg1: uri | IWorkspaceProvider): string {
	if (!arg1) {
		return "";
	}

	if (types.isFunction((arg1 as IWorkspaceProvider).getWorkspace)) {
		const ws = (arg1 as IWorkspaceProvider).getWorkspace();

		return ws ? ws.resource.fsPath : "";
	}

	return (arg1 as uri).fsPath;
}
