import { $, Builder } from "vs/base/browser/builder";
import { IDisposable } from "vs/base/common/lifecycle";

import "sourcegraph/workbench/styles/leftRightWidget.css";

export interface IRenderer {
	(container: HTMLElement): IDisposable;
}

export class LeftRightWidget {

	private $el: Builder;
	private toDispose: IDisposable[];

	constructor(container: Builder, renderLeftFn: IRenderer, renderRightFn: IRenderer);
	constructor(container: HTMLElement, renderLeftFn: IRenderer, renderRightFn: IRenderer);
	constructor(container: any, renderLeftFn: IRenderer, renderRightFn: IRenderer) {
		this.$el = $(".monaco-left-right-widget").appendTo(container);

		this.toDispose = [
			renderLeftFn($(".left-right-widget_left").appendTo(this.$el).getHTMLElement()),
			renderRightFn($(".left-right-widget_right").appendTo(this.$el).getHTMLElement()),
		].filter(x => !!x);
	}

	public setClassNames(classNames: string): void {
		const classList = this.$el.getHTMLElement().classList;
		classList.add.apply(classList, arguments);
	}

	public dispose(): void {
		if (this.$el) {
			this.$el.destroy();
			this.$el = null as any;
		}
	}
}
