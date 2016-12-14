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
		paddingLeft: `calc(${whitespace[4]} - ${borderWidth}px)`,
		paddingRight: whitespace[3],
	};

	const tabSize = {
		"small": typography.size[7],
		"large": typography.size[3],
	};

	const sx = merge(
		{
			borderWidth: 0,
			borderColor: active ? colors[color]() : "transparent",
			borderStyle: "solid",
			fontSize: "inherit",
		},

		direction === "vertical" ? verticalSx : horizontalSx,
		tabSize && size ? tabSize[size] : null,
		hideMobile ? layout.hide.sm : {},

		select(" a",
			active
				? { color: inverted ? "white" : colors[color]() }
				: { color: colors.coolGray3() }
		),

		select(" a:hover",
			active
				? { color: inverted ? "white" : colors[color]() }
				: { color: inverted ? colors.coolGray4() : colors.coolGray2() }
		)
	);

	return <span {...sx} style={style}>{children}</span>;
}
