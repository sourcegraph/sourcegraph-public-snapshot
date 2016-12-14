import * as React from "react";
import { colors } from "sourcegraph/components/utils";

interface Props {
	size?: "small" | "large" | "default";
	children?: React.ReactNode[];
	direction?: "vertical" | "horizontal";
	style?: React.CSSProperties;
}

export function Tabs({
	children,
	direction = "horizontal",
	style,
}: Props): JSX.Element {

	const sx = Object.assign({
		borderColor: colors.coolGray2(0.2),
		borderWidth: 0,
		borderLeftWidth: direction === "vertical" ? 1 : 0,
		borderBottomWidth: direction === "horizontal" ? 1 : 0,
		borderStyle: "solid",
	}, style);

	return <div style={sx}>{children}</div>;
}
