import * as React from "react";

import { Query, SearchQuery } from "sourcegraph/search/types";
import { Dispatcher } from "sourcegraph/workbench/utils";

interface P {
	dispatcher: Dispatcher<Query>;
}

interface S extends SearchQuery {
}

const sx = {
	width: 400,
	margin: 20,
};

export class QueryBar extends React.Component<P, S> {
	private onChange = (e: React.FormEvent<HTMLInputElement>) => {
		this.setState({
			pattern: (e.target as any).value,
		});
	}

	private onKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
		if (e.key !== "Enter") {
			return;
		}
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

	render(): JSX.Element {
		return <input onKeyDown={this.onKeyDown} onChange={this.onChange} style={sx}>
		</input>;
	}
}
