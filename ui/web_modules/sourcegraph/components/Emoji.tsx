// tslint:disable: typedef ordered-imports

import * as React from "react";

interface Props {
	className?: string;
	width?: string; // appended by "px"
	name: string; // See symbols directory
}

type State = any;

export class Emoji extends React.Component<Props, State> {
	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
	};

	static defaultProps = {
		width: "16px",
	};

	render(): JSX.Element | null {
		return <img src={`${(this.context as any).siteConfig.assetsRoot}/img/emoji/${this.props.name}.svg`} width={this.props.width} className={this.props.className} />;
	}
}
