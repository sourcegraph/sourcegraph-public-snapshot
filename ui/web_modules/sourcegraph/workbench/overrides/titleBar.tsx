import * as React from "react";
import * as ReactDOM from "react-dom";
import { context } from "sourcegraph/app/context";
import { GlobalNav } from "sourcegraph/app/GlobalNav";
import { hiddenNavRoutes } from "sourcegraph/app/GlobalNav";
import { abs, isAtRoute } from "sourcegraph/app/routePatterns";
import { __getRouterForWorkbenchOnly } from "sourcegraph/app/router";
import { RouterContext } from "sourcegraph/workbench/utils";

import { Build, Builder, Dimension } from "vs/base/browser/builder";
import { TitlebarPart as VSTitlebarPart } from "vscode/src/vs/workbench/browser/parts/titlebar/titlebarPart";

export class TitlebarPart extends VSTitlebarPart {
	updateTitle(): void {
		//
	}

	public createContentArea(parent: Builder): Builder {
		const container = document.createElement("div");
		container.style.display = "flex";
		container.style.width = `${window.innerWidth}`;
		container.style.flex = "1";
		container.style.flexDirection = "column";
		ReactDOM.render(<RouterContext><GlobalNav /></RouterContext>, container);
		parent.append(container);
		return super.createContentArea(parent);
	}

	layout(dimension: Dimension): Dimension[] {
		// Adjust render for if the nav bar is going to be visible.
		const isHomeRoute = isAtRoute(__getRouterForWorkbenchOnly(), abs.home);
		const shouldHide = hiddenNavRoutes.has(location.pathname) || (isHomeRoute && !context.user && context.authEnabled);
		const titleBuilder = Build.withElementById("workbench.parts.titlebar");
		if (titleBuilder) {
			titleBuilder.display(shouldHide ? "none" : "flex");
		}
		return super.layout(dimension);
	}
}
