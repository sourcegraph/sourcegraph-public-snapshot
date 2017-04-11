import * as React from "react";

import { Query, SearchQuery } from "sourcegraph/search/types";
import { Dispatcher } from "sourcegraph/workbench/utils";

interface P {
	dispatcher: Dispatcher<Query>;
	initialQuery: string;
}

interface S extends SearchQuery {
}

const sx = {
	width: 400,
	margin: 20,
};

export class QueryBar extends React.Component<P, S> {
	state: S = {
		pattern: this.props.initialQuery,
	};

	private onChange = (e: React.FormEvent<HTMLInputElement>) => {
		this.setState({
			pattern: (e.target as any).value,
		});
	}

	private triggerSearch(): void {
		this.props.dispatcher.dispatch({
			query: {
				pattern: this.state.pattern,
				isCaseSensitive: false,
				isMultiline: false,
				isRegExp: false,
				isWordMatch: false,
			}
		});
	}

	private onKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
		if (e.key !== "Enter") {
			return;
		}
		this.triggerSearch();
	}

	render(): JSX.Element {
		return <input
			onKeyDown={this.onKeyDown}
			value={this.state.pattern}
			onChange={this.onChange}
			style={sx} />;
	}
}
