import * as React from "react";

import {ModalComp} from "sourcegraph/components/Modal";
import { Category, SearchContainer } from "sourcegraph/search/modal/SearchContainer";

// Keyboard shortcuts
export const shortcuts = {
	repo: "r",
	file: "t",
	def: "f",
	search: "s",
};

interface State {
	showModal: boolean;
	start: Category | null;
}

interface Node {
	target: {
		nodeName: string;
	};
};

export type RepoRev = {
	repo: string,
	commitID: string,
}

// SearchModal controls when and how to show the search modal.
export class SearchModal extends React.Component<RepoRev, State> {
	constructor() {
		super();
		this.searchModalShortcuts = this.searchModalShortcuts.bind(this);
		this.dismissModal = this.dismissModal.bind(this);
		this.state = {
			showModal: true,
			start: null,
		};
	}

	componentWillMount(): void {
		document.body.addEventListener("keydown", this.searchModalShortcuts);
	}

	searchModalShortcuts(event: KeyboardEvent & Node): void {
		if (event.key === "Escape") {
			this.setState({showModal: false, start: null});
		}
		if (event.target.nodeName === "INPUT" || event.metaKey || event.ctrlKey) {
			return;
		}
		if (event.key === shortcuts.search) {
			this.setState({showModal: !this.state.showModal, start: null});
		}
		if (event.key === shortcuts.file) {
			this.setState({showModal: true, start: Category.file});
		}
		if (event.key === shortcuts.repo) {
			this.setState({showModal: true, start: Category.repository});
		}
		if (event.key === shortcuts.def) {
			this.setState({showModal: true, start: Category.definition});
		}
		event.preventDefault();
	}

	dismissModal(): void {
		const state = Object.assign(this.state, {showModal: false});
		this.setState(state);
	}

	render(): JSX.Element {
		if (!this.state.showModal) {
			return <div />;
		}
		return <ModalComp onDismiss={this.dismissModal}>
			<SearchContainer
				repo={this.props.repo}
				commitID={this.props.commitID}
				start={this.state.start}
				dismissModal={this.dismissModal} />
		</ModalComp>;
	}
}
