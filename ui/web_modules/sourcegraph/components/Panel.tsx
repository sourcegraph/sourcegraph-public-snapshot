import * as classnames from "classnames";
import { hover as gHover } from "glamor";
import * as React from "react";
import { colors } from "sourcegraph/components/utils/index";

interface Props {
	className?: string;
	children?: any;
	color?: "blue" | "purple" | "green" | "orange" | "white" | "coolGray3";
	hoverLevel?: "high" | "low";
	hover?: boolean;
	onClick?: () => void;
	style?: React.CSSProperties;
}

export function Panel({
	className,
	children,
	color = "white",
	hoverLevel,
	hover,
	onClick,
	style,
}: Props): JSX.Element {
	const sx = Object.assign(
		{
			backgroundColor: colors[color](),
			borderRadius: "3px",
			color: color !== "white" ? "white" : "inherit",
			transition: "all 550ms cubic-bezier(0.175, 0.885, 0.320, 1)",
		},
		hoverLevel ? {
			boxShadow: `0 ${hoverLevel === "high" ? "2px 24px" : "2px 8px"} 0 ${colors.blueGrayD1(0.25)}`,
		} : {},
		style,
	);

	const hoverSx = gHover({
		boxShadow: `0 ${hoverLevel === "high" ? "10px 32px" : "2px 16px"} 3px ${colors.blueGrayD2(0.25)} !important`,
	}).toString();

	return <div className={classnames(hover ? hoverSx : "", className)} onClick={onClick} style={sx}>{children}</div>;
};
