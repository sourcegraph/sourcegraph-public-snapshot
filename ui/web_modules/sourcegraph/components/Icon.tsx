import * as React from "react";

import {context} from "sourcegraph/app/context";

interface Props {
	className?: string;
	width?: string; // appended by "px"
	height?: string; // appended by "px"
	icon: string; // See symbols directory
}

export class Icon extends React.Component<Props, {}> {

	static defaultProps: {
		width: "16px",
		height: "auto",
	};

	render(): JSX.Element | null {
		return <img src={`${context.assetsRoot}/img/symbols/${this.props.icon}.svg`} width={this.props.width} height={this.props.height} className={this.props.className} />;
	}
}
