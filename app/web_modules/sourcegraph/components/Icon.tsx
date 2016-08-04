// tslint:disable

import * as React from "react";

class Icon extends React.Component<any, any> {
	static propTypes = {
		className: React.PropTypes.string,
		width: React.PropTypes.string, // appended by "px"
		height: React.PropTypes.string, // appended by "px"
		icon: React.PropTypes.string.isRequired, // See symbols directory
	};

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

export default Icon;
