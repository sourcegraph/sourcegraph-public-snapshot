import {hover as gHover} from "glamor";
import * as React from "react";
import {colors} from "sourcegraph/components/utils/index";

interface Props {
	className?: string;
	children?: any;
	color?: "blue" | "purple" | "green" | "orange" | "white" | "coolGray3";
	hoverLevel?: "high" | "low";
	hover?: boolean;
	style?: React.CSSProperties;
}

export function Panel(props: Props): JSX.Element {
	const {
		className,
		children,
		color = "white",
		hoverLevel,
		hover,
		style,
	} = props;

	const sx = Object.assign({},
		{
			backgroundColor: colors[color](),
			borderRadius: "3px",
			color: color !== "white" ? "white" : "inherit",
		},
		hoverLevel ? {
			boxShadow: `0 ${hoverLevel === "high" ? "2px 25px" : "2px 5px"} 2px ${colors.black(0.05)}`,
		} : null,
		style,
	);

	const hoverSx = gHover(
		{
			transition: "all 550ms cubic-bezier(0.175, 0.885, 0.320, 1)",
			boxShadow: `0 ${hoverLevel === "high" ? "10px 35px" : "2px 6px"} 3px ${colors.black(0.05)}`,
		}
	);

	return <div className={className}
		{...hoverSx}
		style={sx}>
			{children}
		</div>;
};
