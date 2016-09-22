import * as React from "react";

import {ModalComp} from "sourcegraph/components/Modal";
import {SearchContainer} from "sourcegraph/search/modal/SearchContainer";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";

interface State {
	showModal: boolean;
}

interface Node {
	target: {
		nodeName: string;
	};
};

export type RepoSpec = {
	repo: string,
	commitID?: string,
	rev: string | null,
}

// SearchModal controls when and how to show the search modal.
export class SearchModal extends React.Component<RepoSpec, State> {
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

	_getEventProps(): any {
		return {
				repo: this.props.repo,
				rev: this.props.rev,
				query: (this.refs as {searchContainer: SearchContainer}).searchContainer.state.input,
			};
	}

	searchModalShortcuts(event: KeyboardEvent & Node): void {
		if (event.key === "Escape") {
			this.dismissModal();
		}
		if (event.target.nodeName === "INPUT" || event.metaKey || event.ctrlKey) {
			return;
		}
		if (event.key === "/") {
			this.setState({showModal: !this.state.showModal});
			EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_JUMP_TO, AnalyticsConstants.ACTION_TOGGLE, "JumpToInitiated", this._getEventProps());
		}
		event.preventDefault();
	}

	dismissModal(resultSelected: boolean = false): void {
		if (!resultSelected) {
			EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_JUMP_TO, AnalyticsConstants.ACTION_TOGGLE, "JumpToDismissed", this._getEventProps());
		}
		const state = Object.assign(this.state, {showModal: false});
		this.setState(state);
	}

	render(): JSX.Element {
		if (!this.state.showModal) {
			return <div />;
		}
		return <ModalComp onDismiss={this.dismissModal}>
			<SearchContainer
				ref="searchContainer"
				{...this.props}
				dismissModal={this.dismissModal} />
		</ModalComp>;
	}
}
