import { fetchBlobContent } from "app/backend";
import { BlobPosition } from "app/util/types";
import * as React from "react";

interface Props extends BlobPosition {
	// How many lines to show in the excerpt.
	previewLines?: number;
}

interface State {
	blobLines?: string[];
}

export class CodeExcerpt extends React.Component<Props, State> {
	constructor(props: Props) {
		super(props);
		this.state = {};
	}

	componentDidMount(): void {
		fetchBlobContent(this.props.uri, this.props.rev, this.props.path).then(content => {
			const blobLines = content.split("\n");
			this.setState({ blobLines });
		});
	}

	render(): JSX.Element | null {
		if (!this.state.blobLines) {
			return null;
		}
		return <div>
			{this.state.blobLines[this.props.line]}
		</div>;
	}
}
