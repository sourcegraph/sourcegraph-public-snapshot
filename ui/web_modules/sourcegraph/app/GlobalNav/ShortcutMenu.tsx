import { css, lastChild } from "glamor";
import * as React from "react";

import { Router, RouterLocation } from "sourcegraph/app/router";
import { EventListener, isNonMonacoTextArea } from "sourcegraph/Component";
import { Key } from "sourcegraph/components/Key";
import { LocationStateModal, dismissModal, setLocationModalState } from "sourcegraph/components/Modal";
import { whitespace } from "sourcegraph/components/utils";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";

interface Props {
	location: RouterLocation;
	onDismiss?: () => void;
	router: Router;
};

const modalName = "keyboardShortcuts";

export class ShortcutModal extends React.Component<Props, {}> {

	shortcut = (event: KeyboardEvent & Event): void => {
		const { location, router } = this.props;
		if ((event.target as any).nodeName === "INPUT" || isNonMonacoTextArea((event.target as Node))) {
			return;
		}

		const SlashKeyCode = 191;
		if (event.shiftKey && (event.key === "/" || event.keyCode === SlashKeyCode)) {
			if (location && (location as any).state && (location as any).state.modal === modalName) {
				this.dismissModal();
			} else {
				setLocationModalState(router, modalName, true);
			}
			event.preventDefault();
		}
	}

	dismissModal = (): void => {
		Events.ShortcutMenu_Dismissed.logEvent();
		dismissModal(modalName, this.props.router)();
	}

	render(): JSX.Element {
		const { onDismiss } = this.props;
		return <div>
			<LocationStateModal title="Keyboard shortcuts" modalName={modalName} onDismiss={onDismiss} style={{ maxWidth: "20rem" }}>
				{shortcutElement("?", "Show keyboard shortcuts")}
				{shortcutElement("/", "Open quick search")}
				{shortcutElement("a", "Toggle authors")}
				{shortcutElement("g", "View on GitHub")}
				{shortcutElement("y", "Show commit hash in URL bar")}
			</LocationStateModal>
			<EventListener target={global.document.body} event="keydown" callback={this.shortcut} />
		</div>;
	}
}

const shortcutStyle = css(
	lastChild({ marginBottom: 0 }),
	{
		marginBottom: whitespace[3],
		display: "block"
	},
);

function shortcutElement(letter: string, text: string): JSX.Element {
	return <div {...shortcutStyle}>
		<Key shortcut={letter} />
		<span style={{ marginLeft: whitespace[2] }}>{text}</span>
	</div>;
}
