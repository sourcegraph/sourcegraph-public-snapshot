import React from "react";

import Component from "sourcegraph/Component";

import CSSModules from "react-css-modules";
import styles from "./styles/icons.css";

class Icon extends Component {
	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
	}

	render() {
		return (
			<span className={this.state.className}><i styleName={`icon icon-${this.state.name}`} /></span>
		);
	}
}

Icon.propTypes = {
	name: React.PropTypes.string.isRequired,
};

export default CSSModules(Icon, styles, {allowMultiple: true});
