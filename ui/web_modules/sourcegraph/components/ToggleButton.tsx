import * as React from "react";
import { Button } from "sourcegraph/components";
import { colors, whitespace } from "sourcegraph/components/utils";

interface Props {
	children?: React.ReactNode[];
	on?: boolean;
	onClick: React.EventHandler<React.SyntheticEvent<any>>;
	style?: React.CSSProperties;
	size?: "small" | "large";
}

export function ToggleButton({ children, on, onClick, style, size }: Props): JSX.Element {
	const sx = Object.assign(
		{
			backgroundColor: on ? null : colors.blueGrayD1(),
			paddingLeft: "0.75rem",
			paddingRight: whitespace[2],
		},
		style,
	);
	return <Button color={on ? "blue" : "blueGray"} style={sx} size={size} onClick={onClick}>{children}</Button>;
}
