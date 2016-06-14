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
		const {_affix} = this.refs;
		const initialOffset = _affix.offsetTop;
		window.addEventListener("scroll", (e) => {
			if (initialOffset <= window.scrollY) {
				_affix.style.position = "fixed";
				_affix.style.top = `${this.props.offset}px`;
			} else if (initialOffset > window.scrollY) {
				_affix.style.position = "relative";
			}
		});
	}

	render() {
		const {className, style, children} = this.props;
		return (
			<div className={className} style={style}>
				<div ref="_affix">{children}</div>
			</div>
		);
	}
}

export default Affix;
