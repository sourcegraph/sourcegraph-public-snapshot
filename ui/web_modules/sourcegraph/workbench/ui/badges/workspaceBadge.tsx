"use strict";

import { $, Builder } from "vs/base/browser/builder";
import * as strings from "vs/base/common/strings";

export class WorkspaceBadge {

	private $el: Builder;
	private workspace: "Local" | "External";
	private titleFormat: string;

	constructor(container: Builder, workspace: "Local" | "External", titleFormat?: string);
	constructor(container: HTMLElement, workspace: "Local" | "External", titleFormat?: string);
	constructor(container: any, workspace: "Local" | "External", titleFormat?: string) {
		this.$el = $(`.workspace-references-badge-${workspace.toLowerCase()}`).appendTo(container);
		this.titleFormat = titleFormat || "";
		this.setWorkspace(workspace);
	}

	public setWorkspace(workspace: "Local" | "External"): void {
		this.workspace = workspace;
		this.render();
	}

	public setTitleFormat(titleFormat: string): void {
		this.titleFormat = titleFormat;
		this.render();
	}

	private render(): void {
		this.$el.text("" + this.workspace);
		this.$el.title(strings.format(this.titleFormat, this.workspace));
	}

	public dispose(): void {
		if (this.$el) {
			this.$el.destroy();
			this.$el = null as any;
		}
	}
}
