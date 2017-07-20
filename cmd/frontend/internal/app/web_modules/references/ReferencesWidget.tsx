import { fetchReferences } from "app/backend/lsp";
import { CloseIcon } from "app/components/Icons";
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
	}

	getRangeString(ref: Reference): string {
		return `${ref.range.start.line}:${ref.range.start.character}-${ref.range.end.line}:${ref.range.end.character}`;
	}

	render(): JSX.Element | null {
		if (!this.state.context) {
			return null;
		}
		return <div>
			<div style={{ display: "flex", alignItems: "center", borderBottom: "1px solid #e1e4e8", padding: "10px" }}>
				<div style={{ flex: 1 }}>
					{this.state.context.coords.word}
				</div>
				<CloseIcon onClick={() => setReferences({ ...store.getValue, docked: false })} />
			</div>
			{
				this.state.data && this.state.data.references &&
				this.state.data.references.map((ref, i) => <div key={i} style={{ padding: "5px", borderBottom: "1px solid #e1e4e8" }}>
					{`${URI.parse(ref.uri).fragment} @ ${this.getRangeString(ref)}`}
				</div>)
			}
		</div>;
	}
}
