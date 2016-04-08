import React from "react";
import {Link as RouterLink} from "react-router";

import Component from "sourcegraph/Component";

import CSSModules from "react-css-modules";
import styles from "./styles/link.css";

class Link extends Component {
	contextTypes: {
		router: React.PropTypes.object.isRequired,
	}

	constructor(props) {
		super(props);
	}

	shouldComponentUpdate() {
		if (super.shouldComponentUpdate) return true;
		if (global.window) {
			// Active state may have changed based on window
			// location.
			return this.state.currLocation !== window.location.href;
		}
		return false;
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		if (global.window) state.currLocation = window.location.href;
	}

	render() {
		let style = this.state.styl;
		if (style && !style.match(/default|primary/)) {
			// assume "button" alternative styling; give a default color
			style = `${style} default`;
		}

		return (
			this.state.to ?
				<RouterLink styleName={style ? style : "link"}
					activeClassName={style ? styles.active : null}
					to={this.state.to}
					onClick={this.state.onClick}>
					{this.state.children}
				</RouterLink> :
				<a styleName={style ? style : "link"}
					href={this.state.href}
					onClick={this.state.onClick}>
					{this.state.children}
				</a>
		);
	}
}

Link.propTypes = {
	to: React.PropTypes.string,
	href: React.PropTypes.string, // circumvent react router handling
	styl: React.PropTypes.string, // e.g. "button primary"
	onClick: React.PropTypes.func,
};

export default CSSModules(Link, styles, {allowMultiple: true});
