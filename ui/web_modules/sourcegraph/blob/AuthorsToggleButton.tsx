import * as autobind from "autobind-decorator";
import * as React from "react";

import { IDisposable } from "vs/base/common/lifecycle";

import { EventListener, isNonMonacoTextArea } from "sourcegraph/Component";
import { ToggleButton } from "sourcegraph/components";
import { layout, typography, whitespace } from "sourcegraph/components/utils";
import { isOnZapRef } from "sourcegraph/editor/config";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { isCodeLensEnabled } from "sourcegraph/workbench/ConfigurationService";
import { onWorkspaceUpdated } from "sourcegraph/workbench/services";

interface Props {
	keyCode: number;
	shortcut: string;
	toggleAuthors: () => void;
}

interface State {
	on: boolean;
	isViewingZapRef: boolean;
}

@autobind
export class AuthorsToggleButton extends React.Component<Props, State> {
	disposables: IDisposable[];

	constructor(props: Props) {
		super(props);
		this.disposables = [];
		this.state = {
			on: isCodeLensEnabled(),
			isViewingZapRef: isOnZapRef(),
		};
	}

	componentDidMount(): void {
		this.disposables.push(onWorkspaceUpdated(workspace => this.setState({
			isViewingZapRef: Boolean(workspace.revState && workspace.revState!.zapRef),
		} as State)));
	}

	componentWillUnmount(): void {
		this.disposables.forEach(disposable => disposable.dispose());
	}

	toggleAuthors(): void {
		this.setState({ on: !this.state.on } as State);
		this.props.toggleAuthors();
	}

	showAuthorsClickHandler(): void {
		this.toggleAuthors();
		Events.AuthorsToggle.logEvent({ toggleAuthors: this.state.on, type: "click" });
	}

	showAuthorsKeyHandler(event: KeyboardEvent & Event): void {
		if (isOnZapRef() || this.state.isViewingZapRef) {
			return;
		}
		// Don't toggle if in an input on textarea
		const eventTarget = event.target as Node;
		if (eventTarget.nodeName === "INPUT" || isNonMonacoTextArea(eventTarget) || event.metaKey || event.ctrlKey) {
			return;
		}
		const keyCode = this.props.keyCode;
		if (event.key === this.props.shortcut || event.keyCode === keyCode) {
			this.toggleAuthors();
			Events.AuthorsToggle.logEvent({ toggleAuthors: this.state.on, type: "keyboardShortcut" });
			event.preventDefault();
		}
	}

	render(): JSX.Element | null {
		const toggleButtonSx = Object.assign({
			position: "relative",
		}, typography.size[7]);

		return this.state.isViewingZapRef ? null : <div style={{ display: "inline-block", marginLeft: whitespace[2] }} { ...layout.hide.sm}>
			<ToggleButton
				size="small"
				on={this.state.on}
				style={toggleButtonSx}
				onClick={this.showAuthorsClickHandler}>
				Show authors
			</ToggleButton>

			<EventListener
				target={global.document.body}
				event="keydown"
				callback={this.showAuthorsKeyHandler} />
		</div>;
	}
};
