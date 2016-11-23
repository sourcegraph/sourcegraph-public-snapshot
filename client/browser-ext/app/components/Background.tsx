import * as backend from "../backend";
import * as utils from "../utils";
import {EventLogger} from "../utils/EventLogger";
import * as React from "react";

export class Background extends React.Component<{}, {}> {
	constructor(props: {}) {
		super(props);
		this._refresh = this._refresh.bind(this);
		this._cleanupAndRefresh = this._cleanupAndRefresh.bind(this);
		this._popstateUpdate = this._popstateUpdate.bind(this);
	}

	componentDidMount(): void {
		document.addEventListener("pjax:end", this._cleanupAndRefresh);
		window.addEventListener("popstate", this._popstateUpdate);
		this._cleanupAndRefresh();
	}

	componentWillUpdate(nextProps: {}): void {
		// Call refresh with new props (since this.props are not updated until this method completes).
		this._refresh(nextProps);
	}

	componentWillUnmount(): void {
		document.removeEventListener("pjax:end", this._cleanupAndRefresh);
		document.removeEventListener("popstate", this._popstateUpdate);
	}

	removePopovers(): void {
		const popovers = document.getElementsByClassName("sg-popover");
		for (let i = popovers.length; i > 0; ) {
			popovers[--i].remove();
		}
	}

	_cleanupAndRefresh(): void {
		// Clean up any popovers on the page before refreshing (after pjax:success).
		// Otherwise, popovers may remain on the page because the anchored elem's mousout
		// event may not have fired (and the elem may no longer be on the page).
		this.removePopovers();
		this._refresh();
	}

	_popstateUpdate(): void {
		this.removePopovers();
	}

	_refresh(props?: {}): void {
		if (utils.isSourcegraphURL(window.location)) {
			return;
		}

		if (!props) {
			props = this.props;
		}

		let urlProps = utils.parseURL(window.location);

		if (urlProps.repoURI) {
			backend.ensureRepoExists(urlProps.repoURI);
		}

		chrome.runtime.sendMessage({type: "getIdentity"}, (identity) => {
			if (identity) {
				EventLogger.updatePropsForUser(identity);
			}
		});
	}

	render(): JSX.Element | null {
		return null; // the injected app is for bootstrapping; nothing needs to be rendered
	}
}
