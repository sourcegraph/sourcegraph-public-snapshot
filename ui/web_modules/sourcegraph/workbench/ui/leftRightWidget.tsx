import * as ReactDOM from "react-dom";

import { $, Builder } from "vs/base/browser/builder";
import { IDisposable } from "vs/base/common/lifecycle";

import "sourcegraph/workbench/styles/leftRightWidget.css";

export interface IRenderer {
	(container: HTMLElement): IDisposable;
}

export class LeftRightWidget {

	private $el: Builder;
	private toDispose: IDisposable[];
	private iconElement: HTMLElement | null;

	constructor(container: Builder, renderLeftFn: IRenderer, renderRightFn: IRenderer);
	constructor(container: HTMLElement, renderLeftFn: IRenderer, renderRightFn: IRenderer);
	constructor(container: any, renderLeftFn: IRenderer, renderRightFn: IRenderer) {
		this.$el = $(".monaco-left-right-widget").appendTo(container);

		this.toDispose = [
			renderLeftFn($(".left-right-widget_left").appendTo(this.$el).getHTMLElement()),
			renderRightFn($(".left-right-widget_right").appendTo(this.$el).getHTMLElement()),
		].filter(x => Boolean(x));
	}

	public setIcon(icon: JSX.Element): void {
		if (!this.iconElement) {
			const widgetContainer = this.$el.getHTMLElement();
			this.iconElement = document.createElement("span");
			this.iconElement.setAttribute("class", "left-right-widget_icon");
			widgetContainer.insertBefore(this.iconElement, widgetContainer.firstChild);
		}
		ReactDOM.render(icon, this.iconElement);
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
