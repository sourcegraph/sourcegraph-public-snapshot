import React from "react";

// This componet renders the Sourcegraph logo, logomark, or logomark with tagline at custom sizes.

class Logo extends React.Component {

	static propTypes = {
		className: React.PropTypes.string,
		type: React.PropTypes.string,
		width: React.PropTypes.string,
	};

	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
	};

	render() {
		const {width, type, className} = this.props;

		let logoImg = "sourcegraph-mark.svg";
		if (type === "logotype") logoImg = "sourcegraph-logo.svg";
		if (type === "logotype-with-tag") logoImg = "sourcegraph-logo-tagline.svg";

		return <img src={`${this.context.siteConfig.assetsRoot}/img/${logoImg}`} width={width} className={className} />;
	}
}

export default Logo;
