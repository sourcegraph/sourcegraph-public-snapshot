import { merge, select } from "glamor";
import * as React from "react";
import { colors, layout, typography, whitespace } from "sourcegraph/components/utils";

interface Props {
	active?: boolean;
	children?: any;
	color?: "blue" | "purple";
	direction?: "vertical" | "horizontal";
	hideMobile?: boolean;
	inverted?: boolean;
	size?: "small" | "large";
	style?: React.CSSProperties;
}

export function TabItem({
	active,
	children,
	color = "blue",
	direction = "horizontal",
	hideMobile,
	inverted,
	size,
	style,
}: Props): JSX.Element {

	const borderWidth = 3;

	const horizontalSx = {
		display: "inline-block",
		borderBottomWidth: `${borderWidth}px !important`,
		margin: whitespace[2],
		marginBottom: -1,
		marginTop: 0,
		padding: whitespace[2],
		paddingTop: 12,
	};

	const verticalSx = {
		display: "block",
		borderLeftWidth: `${borderWidth}px !important`,
		marginBottom: whitespace[2],
		marginLeft: -1,
		padding: whitespace[1],
		paddingLeft: `calc(${whitespace[3]} - ${borderWidth}px)`,
		paddingRight: whitespace[3],
	};

	const tabSize = {
		"small": typography.size[7],
		"large": typography.size[3],
	};

	const sx = merge(
		{
			borderWidth: 0,
			borderColor: active ? colors[`${color}L1`]() : "transparent",
			borderStyle: "solid",
			fontSize: "inherit",
			fontWeight: active ? typography.weight[2] : null,
		},

		direction === "vertical" ? verticalSx : horizontalSx,
		tabSize && size ? tabSize[size] : {},
		hideMobile ? layout.hide.sm : {},

		select(" a",
			active
				? { color: inverted ? "white" : colors[color]() }
				: { color: inverted ? colors.blueGrayL1() : colors.blueGray() }
		),

		select(" a:hover",
			active
				? { color: inverted ? "white" : colors[color]() }
				: { color: inverted ? colors.blueGrayL3() : colors.blueGrayD1() }
		)
	);

	return <span {...sx} style={style}>{children}</span>;
}
