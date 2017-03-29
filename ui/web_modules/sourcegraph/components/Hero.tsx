import * as React from "react";
import { BGContainer } from "sourcegraph/components/BGContainer";
import { colors, whitespace } from "sourcegraph/components/utils";

import { context } from "sourcegraph/app/context";

interface Props {
	className?: string;
	color?: "transparent" | "white" | "purple" | "blue" | "green" | "dark";
	children?: React.ReactNode[];
	pattern?: "objects";
	style?: React.CSSProperties;
}

export function Hero({ color = "white", className, pattern, children, style }: Props): JSX.Element {
	return <BGContainer img={pattern ? patterns[pattern] : ""} className={className} style={{
		backgroundColor: bgColors[color],
		color: color === "white" || color === "transparent" ? "inherit" : colors.white(),
		textAlign: "center",
		paddingBottom: whitespace[5],
		paddingTop: whitespace[5],
		...style
	}}>{children}</BGContainer>;
}

const bgColors = {
	"transparent": "transparent",
	"white": colors.white(),
	"purple": colors.purple(),
	"blue": colors.blue(),
	"dark": colors.blueGrayD1(),
	"green": colors.green(),
};

const patterns = {
	"objects": `${context.assetsRoot}/img/backgrounds/pattern.svg`,
};
