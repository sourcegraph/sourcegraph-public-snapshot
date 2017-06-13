import * as React from "react";
import * as backend from "../backend";
import { ExtensionEventLogger } from "../tracking/ExtensionEventLogger";
import * as utils from "../utils";
import { eventLogger } from "../utils/context";
import * as tooltips from "../utils/tooltips";

export class GitHubBackground extends React.Component<{}, {}> {

	componentDidMount(): void {
		document.addEventListener("pjax:end", this.cleanupAndRefresh);
		window.addEventListener("popstate", this.cleanupAndRefresh);
		this.cleanupAndRefresh();
	}

	componentWillUnmount(): void {
		document.removeEventListener("pjax:end", this.cleanupAndRefresh);
		window.removeEventListener("popstate", this.cleanupAndRefresh);
	}

	private cleanupAndRefresh = (): void => {
		// Clean up any tooltips on the page before refreshing (after pjax:success).
		// Otherwise, tooltips may remain on the page because the anchored elem's mousout
		// event may not have fired (and the elem may no longer be on the page).
		// tooltips.hideTooltip();
		// (eventLogger as ExtensionEventLogger).updateIdentity();

		// // Remove all ".sg-annotated"; this allows tooltip event handlers to be re-registered.
		// const sgAnnotated = document.querySelectorAll(".sg-annotated");
		// // tslint:disable-next-line
		// console.log("cleanup found sgannotated", sgAnnotated.length);
		// for (let i = 0; i < sgAnnotated.length; ++i) {
		// 	const item = sgAnnotated.item[i] as HTMLElement;
		// 	if (item && item.classList) {
		// 		item.classList.remove("sg-annotated");
		// 	}
		// }
	}

	render(): JSX.Element | null {
		return null; // the injected app is for bootstrapping; nothing needs to be rendered
	}
}
