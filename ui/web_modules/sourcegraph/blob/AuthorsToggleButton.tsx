import * as autobind from "autobind-decorator";
import * as React from "react";

import { EventListener, isNonMonacoTextArea } from "sourcegraph/Component";
import { Key, ToggleButton } from "sourcegraph/components";
import { layout, typography, whitespace } from "sourcegraph/components/utils";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { Features } from "sourcegraph/util/features";

interface Props {
	keyCode: number;
	shortcut: string;
	toggleAuthors: () => void;
}

interface State {
	on: boolean;
}

@autobind
export class AuthorsToggleButton extends React.Component<Props, State> {
	constructor(props: Props) {
		super(props);
		this.state = {
			on: Features.codeLens.isEnabled(),
		};
	}

	toggleAuthors(): void {
		this.setState({ on: !this.state.on });
		this.props.toggleAuthors();
	}

	showAuthorsClickHandler(): void {
		this.toggleAuthors();
		AnalyticsConstants.Events.AuthorsToggle.logEvent({ toggleAuthors: this.state.on, type: "click" });
	}

	showAuthorsKeyHandler(event: KeyboardEvent & Event): void {
		// Don't toggle if in an input on textarea
		const eventTarget = event.target as Node;
		if (eventTarget.nodeName === "INPUT" || isNonMonacoTextArea(eventTarget) || event.metaKey || event.ctrlKey) {
			return;
		}
		const keyCode = this.props.keyCode;
		if (event.key === this.props.shortcut || event.keyCode === keyCode) {
			this.toggleAuthors();
			AnalyticsConstants.Events.AuthorsToggle.logEvent({ toggleAuthors: this.state.on, type: "keyboardShortcut" });
			event.preventDefault();
		}
	}

	render(): JSX.Element {
		const { shortcut } = this.props;

		const toggleButtonSx = Object.assign({
			marginRight: whitespace[1],
			position: "relative",
		}, typography.size[7]);

		return <div style={{ display: "inline-block" }} { ...layout.hide.sm}>
			<ToggleButton
				size="small"
				on={this.state.on}
				style={toggleButtonSx}
				onClick={this.showAuthorsClickHandler}>
				Show authors
				<Key shortcut={shortcut} style={{ marginLeft: whitespace[2] }} />
			</ToggleButton>

			<EventListener
				target={global.document.body}
				event="keydown"
				callback={this.showAuthorsKeyHandler} />
		</div>;
	}
};
