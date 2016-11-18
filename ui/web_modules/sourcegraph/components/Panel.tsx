import {hover as gHover} from "glamor";
import * as React from "react";
import {colors} from "sourcegraph/components/utils/index";

interface Props {
	className?: string;
	children?: any;
	color?: "blue" | "purple" | "green" | "orange" | "white" | "coolGray3";
	hoverLevel?: "high" | "low";
	hover?: boolean;
	hoverBorder?: boolean;
	style?: React.CSSProperties;
}

export function Panel({
	className,
	children,
	color = "white",
	hoverBorder,
	hoverLevel,
	hover,
	style,
}: Props): JSX.Element {
	const sx = Object.assign(
		{
			backgroundColor: colors[color](),
			borderRadius: "3px",
			color: color !== "white" ? "white" : "inherit",
			borderWidth: 1,
			borderColor: colors.coolGray3(0.2),
			borderStyle: "solid",
		},
		hoverLevel ? {
			boxShadow: `0 ${hoverLevel === "high" ? "2px 20px" : "1px 4px"} 0 ${colors.coolGray1(0.08)}`,
		} : null,
		style,
	);

	const hoverSx = gHover({
		transition: "all 550ms cubic-bezier(0.175, 0.885, 0.320, 1)",
		boxShadow: `0 ${hoverLevel === "high" ? "10px 35px" : "2px 6px"} 3px ${colors.coolGray1(0.05)}`,
		borderColor: hoverBorder ? `${colors.coolGray3(0.4)} !important` : null,
	});

	return <div className={className} {...hoverSx} style={sx}>{children}</div>;
};
