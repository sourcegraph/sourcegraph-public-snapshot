import * as React from "react";

interface Props {
	title: string;
}

let titleSet = false;

export class PageTitle extends React.Component<Props, {}> {
	componentDidMount(): void {
		if (titleSet) {
			console.error("more than one PageTitle used at the same time");
		}
		titleSet = true;
		this._updateTitle(this.props.title);
	}

	componentWillReceiveProps(nextProps: Props): void {
		this._updateTitle(nextProps.title);
	}

	_updateTitle(title: string): void {
		document.title = `${title} · Sourcegraph`;
	}

	componentWillUnmount(): void {
		titleSet = false;
		document.title = "Sourcegraph";
	}

	render(): JSX.Element | null {
		return null;
	}
}
