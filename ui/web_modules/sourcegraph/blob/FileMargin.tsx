// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as ReactDOM from "react-dom";

type Props = {
	lineFromByte?: () => void,
	selectionStartLine?: any,
	startByte?: number,

	[key: string]: any,
};

export class FileMargin extends React.Component<Props, any> {
	componentDidUpdate() {
		const content = this.refs["content"] as HTMLElement;
		if (content) {
			const lineOffsetFromTop = this.getOffsetFromTop();
			const isNearBottom = lineOffsetFromTop > (content.parentNode as HTMLElement).clientHeight - content.clientHeight;

			content.style.top = isNearBottom ? "" : `${lineOffsetFromTop}px`;
			content.style.bottom = isNearBottom ? "0px" : "";
		}
	}

	getOffsetFromTop() {
		if (this.props.selectionStartLine) {
			return (ReactDOM.findDOMNode(this.props.selectionStartLine) as HTMLElement).offsetTop;
		}
		return 0;
	}

	render(): JSX.Element | null {
		let passthroughProps = Object.assign({}, this.props);
		delete passthroughProps.children;
		delete passthroughProps.lineFromByte;

		return (
			<div {...passthroughProps} style={{position: "relative"}}>
				{React.Children.map(this.props.children, (child, i) => (
					<div key={i} ref="content" style={{position: "absolute"}}>{child}</div>
				))}
			</div>
		);
	}
}
