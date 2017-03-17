import * as React from "react";
import * as icons from "sourcegraph/components/symbols/Primaries";
import { colors, whitespace } from "sourcegraph/components/utils";

interface Props {
	className?: string;
	style?: React.CSSProperties;
	color?: "blue" | "purple" | "orange" | "green" | "yellow" | "red" | "gray";
	text?: string;
	icon?: string;
	compact?: boolean;
}

export function Label({
	className,
	style,
	color = "blue",
	text,
	icon,
	compact,
}: Props): JSX.Element {
	return <span className={className} style={Object.assign({
		backgroundColor: colors[color](),
		borderRadius: 3,
		color: color === "yellow" ? colors.black(0.7) : "white",
		display: "inline-block",
		padding: `${whitespace[compact ? 0 : 1]} ${whitespace[2]}`,
	}, style)}>{icon && icons[icon]({ style: { top: -1 } })} {text}</span>;
}
