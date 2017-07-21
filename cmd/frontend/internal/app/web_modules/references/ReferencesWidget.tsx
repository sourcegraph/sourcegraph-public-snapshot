import { fetchDependencyReferences } from "app/backend";
import { fetchReferences } from "app/backend/lsp";
import { ReferencesState, setReferences, store } from "app/references/store";
import { Reference } from "app/util/types";
import * as React from "react";
import * as URI from "urijs";

const deepEqual = require("deep-equal");

export class ReferencesWidget extends React.Component<{}, ReferencesState> {
	subscription: any;

	constructor() {
		super();
		this.state = store.getValue();
	}

	componentDidMount(): void {
		this.subscription = store.subscribe((state) => {
			if (!deepEqual(state.context, this.state.context)) {
				this.setState(state, () => {
					if (state.context) {
						this.fetchReferences();
					}
				});
			}
		});
	}

	componentWillUnmount(): void {
		if (this.subscription) {
			this.subscription.unsubscribe();
		}
	}

	fetchReferences(): void {
		const context = this.state.context!;
		fetchReferences(context.coords.char, context.path, context.coords.line, context.repoRevSpec).then((res) => {
			if (res) {
				this.setState({ ...store.getValue(), data: { references: res } });
			}
		});
		// fetchDependencyReferences(context.repoRevSpec.repoURI, context.repoRevSpec.rev, context.path, context.coords.line, context.coords.char).then((res) => {
		// 	console.log("got xdendency references", res);
		// });
	}

	getRangeString(ref: Reference): string {
		return `${ref.range.start.line}:${ref.range.start.character}-${ref.range.end.line}:${ref.range.end.character}`;
	}

	getRefURL(ref: Reference): string {
		const uri = URI.parse(ref.uri);
		return `http://localhost:3080/${uri.hostname}/${uri.path}@${uri.query}/-/blob/${uri.fragment}#L${ref.range.start.line}`;
	}

	render(): JSX.Element | null {
		if (!this.state.context) {
			return null;
		}
		return <div>
			<div style={{ display: "flex", alignItems: "center", borderBottom: "1px solid #2A3A51", padding: "10px" }}>
				<div style={{ flex: 1 }}>
					{this.state.context.coords.word}
				</div>
				<div onClick={() => setReferences({ ...store.getValue, docked: false })}>X</div>
			</div>
			{
				this.state.data && this.state.data.references &&
				this.state.data.references.map((ref, i) => <div key={i} style={{ padding: "5px", borderBottom: "1px solid #2A3A51" }}>
					<a href={this.getRefURL(ref)}>{`${URI.parse(ref.uri).fragment} @ ${this.getRangeString(ref)}`}</a>
				</div>)
			}
		</div>;
	}
}
