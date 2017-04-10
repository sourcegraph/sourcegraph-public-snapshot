import * as React from "react";
import * as ReactDOM from "react-dom";
import { context } from "sourcegraph/app/context";
import { GlobalNav } from "sourcegraph/app/GlobalNav";
import { hiddenNavRoutes } from "sourcegraph/app/GlobalNav";
import { abs, isAtRoute } from "sourcegraph/app/routePatterns";
import { __getRouterForWorkbenchOnly } from "sourcegraph/app/router";
import { EDITOR_TITLE_HEIGHT } from "sourcegraph/components/utils/layout";
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
		container.style.height = `${EDITOR_TITLE_HEIGHT}px`;
		container.style.flex = "1";
		container.style.flexDirection = "column";
		ReactDOM.render(<RouterContext><GlobalNav /></RouterContext>, container);
		parent.append(container);
		return super.createContentArea(parent);
	}

	layout(dimension: Dimension): Dimension[] {
		// VSCode attempts to calculate the height initially by using the standard zoom factor, which
		// defaults to Infinity when running in the browser. This can cause jumpiness so if it's set to Infinity initially use the default nav height instead.
		if (dimension.height === Infinity) {
			dimension.height = EDITOR_TITLE_HEIGHT;
		}
		const isHomeRoute = isAtRoute(__getRouterForWorkbenchOnly(), abs.home);
		const shouldHide = hiddenNavRoutes.has(location.pathname) || (isHomeRoute && !context.user && context.authEnabled);
		const titleBuilder = Build.withElementById("workbench.parts.titlebar");
		if (titleBuilder) {
			if (!shouldHide) {
				titleBuilder.display("flex").removeClass("builder-hidden");
			} else {
				titleBuilder.display("none");
			}
		}
		return super.layout(dimension);
	}
}
