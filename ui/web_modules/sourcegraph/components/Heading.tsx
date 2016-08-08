// tslint:disable

import * as React from "react";
import * as styles from "./styles/heading.css";
import * as classNames from "classnames";

class Heading extends React.Component<any, any> {
	static propTypes = {
		className: React.PropTypes.string,
		children: React.PropTypes.any,
		level: React.PropTypes.string, //  1 is the largest
		underline: React.PropTypes.string, // blue, purple, white, orange, green
		color: React.PropTypes.string, // purple, blue, green, orange, cool_mid_gray
		align: React.PropTypes.string, // left, right, center
		style: React.PropTypes.object,
	};

	static defaultProps = {
		level: "3", //  1 is the largest
		underline: null,
		color: null,
		align: null,
	};

	render(): JSX.Element | null {
		const {className, children, level, color, underline, align, style} = this.props;

		return (
			<div className={classNames(levelClasses[level] || styles.h3, colorClasses[color], alignClasses[align], className)} style={style}>
				{children}<br />
				{underline && <hr className={classNames(styles.line, underlineClasses[underline])} />}
			</div>
		);
	}
}

const levelClasses = {
	"1": styles.h1,
	"2": styles.h2,
	"3": styles.h3,
	"4": styles.h4,
	"5": styles.h5,
	"6": styles.h6,
	"7": styles.h7,
};

const colorClasses = {
	"purple": styles.purple,
	"blue": styles.blue,
	"green": styles.green,
	"orange": styles.orange,
	"cool_mid_gray": styles.cool_mid_gray,
};

const alignClasses = {
	"left": styles.left,
	"right": styles.right,
	"center": styles.center,
};

const underlineClasses = {
	"blue": styles.l_blue,
	"purple": styles.l_purple,
	"white": styles.l_white,
	"orange": styles.l_orange,
	"green": styles.l_green,
};

export default Heading;
