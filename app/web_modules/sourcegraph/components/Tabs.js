// @flow

import React from "react";
import {Link} from "react-router";
import CSSModules from "react-css-modules";
import styles from "sourcegraph/components/styles/tabs.css";

class Tabs extends React.Component {
	static propTypes = {
		className: React.PropTypes.string,
		links: React.PropTypes.array.isRequired, // [tabText, tabLink]
		active: React.PropTypes.number, // index of the active tab
		color: React.PropTypes.string, // blue, purple
		size: React.PropTypes.string, // null, small, large,
	};

	static defaultProps = {
		active: null,
		color: "blue",
		size: null,
	};

	renderTabs() {
		const {links, active, color} = this.props;

		return links.map((item, i) => {
			const isActive = (function() { return active === i; })();
			return (
				<Link to={item[1]} key={i} styleName={
					`${isActive ? "tab-active " : "tab-inactive"} ${isActive ? color : ""}`
				}>
					{item[0]}
				</Link>
			);
		});
	}

	render() {
		const {className, size} = this.props;
		return (
			<div styleName={`container ${size ? size : ""}`} className={className}>{this.renderTabs()}</div>
		);
	}
}

export default CSSModules(Tabs, styles, {allowMultiple: true});
