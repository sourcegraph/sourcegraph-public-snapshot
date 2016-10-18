import * as React from "react";

import {ModalComp} from "sourcegraph/components/Modal";
import {Container} from "sourcegraph/quickopen/Container";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";

interface Node {
	nodeName: string;
	parentNode: Node;
	classList: DOMTokenList;
};

interface Event {
	target: Node;
};

export type Props = {
	repo: string | null,
	rev: string | null,
	onDismiss: () => void,
	activateSearch: () => void,
	showModal: boolean,
}

// QuickOpenModal controls when and how to show the search modal.
export class QuickOpenModal extends React.Component<Props, null> {
	constructor() {
		super();
		this.searchModalShortcuts = this.searchModalShortcuts.bind(this);
		this.dismissModal = this.dismissModal.bind(this);
	}

	componentWillMount(): void {
		document.body.addEventListener("keydown", this.searchModalShortcuts);
	}

	componentWillUnmount(): void {
		document.body.removeEventListener("keydown", this.searchModalShortcuts);
	}

	_getEventProps(): any {
		let query = "";
		if (this.refs["searchContainer"]) {
			query = (this.refs as {searchContainer: Container}).searchContainer.state.input;
		}
		return {
				repo: this.props.repo,
				rev: this.props.rev,
				query: query,
			};
	}

	searchModalShortcuts(event: KeyboardEvent & Event): void {
		if (event.keyCode === 27) { // Escape.
			this.dismissModal(false);
		}
		if (event.target.nodeName === "INPUT" || isNonMonacoTextArea(event.target) || event.metaKey || event.ctrlKey) {
			return;
		}
		if (event.keyCode === 191) { // Slash key ('/').
			if (!this.props.showModal) {
				this.dismissModal(false);
			}
			this.props.activateSearch();
			EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_QUICK_OPEN, AnalyticsConstants.ACTION_TOGGLE, "QuickOpenInitiated", this._getEventProps());
		}
		event.preventDefault();
	}

	dismissModal(shouldLog: boolean): void {
		if (shouldLog) {
			EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_QUICK_OPEN, AnalyticsConstants.ACTION_TOGGLE, "QuickOpenDismissed", this._getEventProps());
		}
		this.props.onDismiss();
	}

	render(): JSX.Element {
		if (!this.props.showModal) {
			return <div />;
		}
		const r = this.props.repo ? {URI: this.props.repo, rev: this.props.rev} : null;
		return <ModalComp onDismiss={() => this.dismissModal(true)}>
			<Container
				ref="searchContainer"
				repo={r}
				dismissModal={this.dismissModal} />
		</ModalComp>;
	}
}

function isNonMonacoTextArea(n: Node): boolean {
	if (n.nodeName !== "TEXTAREA") {
		return false;
	}
	let p = n.parentNode;
	for (let i = 0; p && i < 20; p = p.parentNode, i++) {
		if (p.classList.contains("monaco-editor")) {
			return false;
		}
	}
	return true;
}
