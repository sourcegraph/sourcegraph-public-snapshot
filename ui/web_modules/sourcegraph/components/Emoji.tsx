// tslint:disable: typedef ordered-imports

import * as React from "react";

type Props = {
	className?: string,
	width?: string, // appended by "px"
	name: string, // See symbols directory
};

export class Emoji extends React.Component<Props, any> {
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
