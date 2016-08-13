// tslint:disable: typedef ordered-imports

import * as React from "react";

interface Props {
	className?: string;
	width?: string; // appended by "px"
	height?: string; // appended by "px"
	icon: string; // See symbols directory
}

type State = any;

export class Icon extends React.Component<Props, State> {
	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
	};

	static defaultProps = {
		width: "16px",
		height: "auto",
	};

	render(): JSX.Element | null {
		return <img src={`${(this.context as any).siteConfig.assetsRoot}/img/symbols/${this.props.icon}.svg`} width={this.props.width} height={this.props.height} className={this.props.className} />;
	}
}
