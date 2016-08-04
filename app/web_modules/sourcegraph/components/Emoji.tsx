// tslint:disable

import * as React from "react";

class Emoji extends React.Component<any, any> {
	static propTypes = {
		className: React.PropTypes.string,
		width: React.PropTypes.string, // appended by "px"
		name: React.PropTypes.string.isRequired, // See symbols directory
	};

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

export default Emoji;
