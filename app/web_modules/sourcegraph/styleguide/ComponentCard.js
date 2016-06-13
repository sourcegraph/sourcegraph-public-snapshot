// @flow

import React from "react";
import base from "sourcegraph/components/styles/_base.css";

class ComponentCard extends React.Component {
	static propTypes = {
		children: React.PropTypes.any,
		component: React.PropTypes.string,
	};

	render() {
		return <div className={base.pv5}>{this.props.children}</div>;
	}
}

export default ComponentCard;
