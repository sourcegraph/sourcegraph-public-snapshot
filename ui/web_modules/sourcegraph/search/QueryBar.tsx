import * as React from "react";

import { Query } from "sourcegraph/search/types";
import { Dispatcher } from "sourcegraph/workbench/utils";
import { IKeyboardEvent } from "vs/base/browser/keyboardEvent";
import { FindInput } from "vs/base/browser/ui/findinput/findInput";

interface P {
	dispatcher: Dispatcher<Query>;
	initialQuery: string;
}

export class QueryBar extends React.Component<P, {}> {

	findInput: FindInput | null;

	private triggerSearch(): void {
		if (!this.findInput) {
			return;
		}
		this.props.dispatcher.dispatch({
			query: {
				pattern: this.findInput.getValue(),
				isCaseSensitive: this.findInput.getCaseSensitive(),
				isMultiline: false,
				isRegExp: this.findInput.getRegex(),
				isWordMatch: this.findInput.getWholeWords(),
			}
		});
	}

	private onKeyDown = (e: IKeyboardEvent) => {
		if (e.browserEvent.key === "Enter") {
			this.triggerSearch();
		}
	}

	private ref = (container: HTMLDivElement | null) => {
		if (!container) {
			return;
		}
		this.findInput = new FindInput(container, null as any, { width: 798, label: "" });
		this.findInput.setValue(this.props.initialQuery);
		this.findInput.onKeyDown(this.onKeyDown);
	}

	render(): JSX.Element {
		return <div style={{ margin: "20px 0" }} ref={this.ref} />;
	}
}
