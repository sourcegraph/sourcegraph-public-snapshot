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
		const initialOffset = this._getOffset();
		window.addEventListener("scroll", () => this._affixEl(initialOffset));
	}

	componentWillUnmount() {
		const initialOffset = this._getOffset();
		window.removeEventListener("scroll", () => this._affixEl(initialOffset));
	}

	_affix: {
		offsetTop: number,
		style: any,
	};

	_getOffset(): number {
		return this._affix.offsetTop;
	}

	_affixEl(initialOffset: number): any {
		if (!this._affix) return;
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
				<div ref={(el) => this._affix = el}>{children}</div>
			</div>
		);
	}
}

export default Affix;
