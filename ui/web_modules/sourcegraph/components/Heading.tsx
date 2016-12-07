import * as React from "react";
import { colors, typography, whitespace } from "sourcegraph/components/utils/index";

interface Props {
	children?: React.ReactNode[];
	level?: number; //  1 is the largest
	underline?: Color;
	color: Color;
	align?: "left" | "right" | "center"; // left, right, center
	compact?: boolean;
	style?: Object;
}

enum Color {
	"blue",
	"purple",
	"white",
	"orange",
	"green",
	"gray"
}

export const Heading = (props: Props): any => {
	const headingColors = {
		blue: colors.blue(),
		purple: colors.purple(),
		white: colors.white(),
		orange: colors.orange(),
		green: colors.green(),
		gray: colors.coolGray3(),
	};

	const sx = Object.assign(
		{
			color: headingColors[props.color],
			fontWeight: typography.weight[2],
			marginBottom: props.compact ? 0 : whitespace[2],
			marginTop: props.compact ? 0 : whitespace[2],
			textTransform: props.level === 7 ? "uppercase" : "auto",
			textAlign: props.align,
		},
		typography.size[props.level ? props.level : 3],
		props.style,
	);

	const underlineSx = {
		borderColor: headingColors[props.underline ? props.underline : "white"],
		borderWidth: "4px",
		display: "inline-block",
		width: "6rem",
		marginBottom: whitespace[3],
		marginTop: whitespace[3],
	};

	return <div style={sx}>
		{props.children} {props.underline && <br />}
		{props.underline && <hr style={underlineSx} />}
	</div>;
};
