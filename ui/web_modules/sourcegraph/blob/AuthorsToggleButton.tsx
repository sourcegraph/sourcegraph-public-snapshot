import * as React from "react";
import { EventListener } from "sourcegraph/Component";
import { Key, ToggleButton } from "sourcegraph/components";
import { layout, typography, whitespace } from "sourcegraph/components/utils";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { Features } from "sourcegraph/util/features";

interface Props {
	shortcut: string;
	onClick: () => void;
}

interface State {
	on: boolean;
}

export class AuthorsToggleButton extends React.Component<Props, State> {
	constructor(props: Props) {
		super(props);
		this.state = {
			on: Features.codeLens.isEnabled() || false,
		};
	}

	_toggleAuthors(): void {
		this.setState({ on: !this.state.on });
		this.props.onClick();
	}

	_showAuthorsClickHandler(): void {
		this._toggleAuthors();
		AnalyticsConstants.Events.AuthorsToggle.logEvent({ toggleAuthors: this.state.on, type: "click" });
	}

	_showAuthorsKeyHandler(event: Event & KeyboardEvent): void {
		const AKeyCode = 65;
		if (event.key === "a" || event.keyCode === AKeyCode) {
			this._toggleAuthors();
			AnalyticsConstants.Events.AuthorsToggle.logEvent({ toggleAuthors: this.state.on, type: "shortcut" });
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
				onClick={() => this._showAuthorsClickHandler()}>
				Show authors
				<Key shortcut={shortcut} style={{ marginLeft: whitespace[2] }} />
			</ToggleButton>

			<EventListener
				target={global.document.body}
				event="keydown"
				callback={(e) => this._showAuthorsKeyHandler(e as KeyboardEvent)} />
		</div>;
	}
};
