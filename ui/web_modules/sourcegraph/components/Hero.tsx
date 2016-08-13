// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as styles from "sourcegraph/components/styles/hero.css";
import * as classNames from "classnames";

interface Props {
	className?: string;
	pattern?: string;
	color?: string; // white, purple, blue, green, dark
	children?: any;
}

type State = any;

export class Hero extends React.Component<Props, State> {
	render(): JSX.Element | null {
		const {color, pattern, children, className} = this.props;

		return (
			<div className={classNames(styles.hero, colorClasses[color || ""] || styles.white, patternClasses[pattern || ""], className)}>
				{children}
			</div>
		);
	}
}

const colorClasses = {
	"transparent": styles.transparent,
	"white": styles.white,
	"purple": styles.purple,
	"blue": styles.blue,
	"dark": styles.dark,
	"green": styles.green,
};

const patternClasses = {
	"objects": styles.bg_img_objects,
	"objects_fade": styles.bg_img_objects_fade,
};
