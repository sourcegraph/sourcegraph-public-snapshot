// @flow

import React from "react";

class Affix extends React.Component {
	static propTypes = {
		className: React.PropTypes.string,
		style: React.PropTypes.object,
		children: React.PropTypes.any,
		offset: React.PropTypes.number,
	};

	componentDidMount() {
		window.addEventListener("scroll", () => this.affixEl());
	}

	componentWillUnmount() {
		window.removeEventListener("scroll", () => this.affixEl());
	}

	affixEl() {
		const initialOffset = this._affix.offsetTop;
		if (initialOffset <= window.scrollY) {
			this._affix.style.position = "fixed";
			this._affix.style.top = `${this.props.offset}px`;
		} else if (initialOffset > window.scrollY) {
			this._affix.style.position = "relative";
		}
	}

	render() {
		const {className, style, children} = this.props;
		return (
			<div className={className} style={style}>
				<div ref={(el) => { this._affix = el; }}>{children}</div>
			</div>
		);
	}
}

export default Affix;
