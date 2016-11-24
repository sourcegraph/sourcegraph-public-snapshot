import * as classNames from "classnames";
import * as React from "react";
import * as styles from "sourcegraph/components/styles/hero.css";
import { colors, whitespace } from "sourcegraph/components/utils";

interface Props {
	className?: string;
	color?: "transparent" | "white" | "purple" | "blue" | "green" | "dark";
	children?: React.ReactNode[];
	pattern?: string;
	style?: React.CSSProperties;
}

export function Hero({color = "white", className, pattern, children, style}: Props): JSX.Element {
	return <div className={classNames(pattern ? patternClasses[pattern] : null, className)} style={{
		backgroundColor: bgColors[color],
		color: color === "white" || color === "transparent" ? "inherit" : colors.white(),
		textAlign: "center",
		paddingBottom: whitespace[4],
		paddingTop: whitespace[4],
	}}>{children}</div>;
}

const bgColors = {
	"transparent": "transparent",
	"white": colors.white(),
	"purple": colors.purple(),
	"blue": colors.blue(),
	"dark": colors.coolGray2(),
	"green": colors.green(),
};

const patternClasses = {
	"objects": styles.bg_img_objects,
	"objects_fade": styles.bg_img_objects_fade,
};
