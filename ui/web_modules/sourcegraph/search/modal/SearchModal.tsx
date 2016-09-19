import * as React from "react";

import {ModalComp} from "sourcegraph/components/Modal";
import {SearchContainer} from "sourcegraph/search/modal/SearchContainer";

interface State {
	showModal: boolean;
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
			showModal: false,
		};
	}

	componentWillMount(): void {
		document.body.addEventListener("keydown", this.searchModalShortcuts);
	}

	componentWillUnmount(): void {
		document.body.removeEventListener("keydown", this.searchModalShortcuts);
	}

	searchModalShortcuts(event: KeyboardEvent & Node): void {
		if (event.key === "Escape") {
			this.setState({showModal: false});
		}
		if (event.target.nodeName === "INPUT" || event.metaKey || event.ctrlKey) {
			return;
		}
		if (event.key === "s") {
			this.setState({showModal: !this.state.showModal});
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
				dismissModal={this.dismissModal} />
		</ModalComp>;
	}
}
