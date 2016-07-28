// @flow

import * as React from "react";

class Icon extends React.Component {
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

	render() {
		return <img src={`${this.context.siteConfig.assetsRoot}/img/symbols/${this.props.icon}.svg`} width={this.props.width} height={this.props.height} className={this.props.className} />;
	}
}

export default Icon;
