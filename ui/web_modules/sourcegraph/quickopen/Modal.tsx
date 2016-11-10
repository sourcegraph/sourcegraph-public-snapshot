import * as React from "react";
import * as Relay from "react-relay";

import {EventListener, isNonMonacoTextArea} from "sourcegraph/Component";
import {ModalComp} from "sourcegraph/components/Modal";
import {Container} from "sourcegraph/quickopen/Container";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

interface Event {
	target: Node;
};

export type Props = {
	repo: string | null,
	rev: string | null,
	onDismiss: () => void,
	activateSearch: (eventProps?: any) => void,
	showModal: boolean,
}

// QuickOpenModal controls when and how to show the search modal.
class QuickOpenModalComponent extends React.Component<Props & {root: GQL.IRoot}, {}> {
	constructor() {
		super();
		this.searchModalShortcuts = this.searchModalShortcuts.bind(this);
		this.dismissModal = this.dismissModal.bind(this);
	}

	static logQuickOpenModalInitiatedEvent(eventProps?: any): void {
		AnalyticsConstants.Events.Quickopen_Initiated.logEvent(eventProps);
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
		if (event.target.nodeName === "INPUT" || isNonMonacoTextArea(event.target) || event.metaKey || event.ctrlKey) {
			return;
		}
		const SlashKeyCode = 191;
		if (event.key === "/" || event.keyCode === SlashKeyCode) {
			if (!this.props.showModal) {
				this.dismissModal(false);
			}
			this.props.activateSearch(this._getEventProps());
			event.preventDefault();
		}
	}

	dismissModal(shouldLog: boolean): void {
		if (shouldLog) {
			AnalyticsConstants.Events.Quickopen_Dismissed.logEvent(this._getEventProps());
		}
		this.props.onDismiss();
	}

	render(): JSX.Element {
		const r = this.props.repo ? {URI: this.props.repo, rev: this.props.rev} : null;
		return <div>
			{this.props.showModal && <ModalComp onDismiss={() => this.dismissModal(true)}>
				<Container
					ref="searchContainer"
					repo={r}
					files={this.props.root.repository ? this.props.root.repository.commit.tree.files : []}
					languages={this.props.root.repository ? this.props.root.repository.commit.languages : []}
					dismissModal={this.dismissModal} />
			</ModalComp>}
			<EventListener target={global.document.body} event="keydown" callback={this.searchModalShortcuts} />
		</div>;
	}
}

const QuickOpenModalContainer = Relay.createContainer(QuickOpenModalComponent, {
	initialVariables: {
		repo: "",
		rev: "",
	},
	fragments: {
		root: () => Relay.QL`
			fragment on Root {
				repository(uri: $repo) {
					commit(rev: $rev) {
						tree(recursive: true) {
							files {
								name
							}
						}
						languages
					}
				}
			}
		`,
	},
});

export const QuickOpenModal = function(props: Props): JSX.Element {
	return <Relay.RootContainer
		Component={QuickOpenModalContainer}
		renderLoading={() => <div></div>}
		route={{
			name: "Root",
			queries: {
				root: () => Relay.QL`
					query { root }
				`,
			},
			params: props,
		}}
	/>;
};
