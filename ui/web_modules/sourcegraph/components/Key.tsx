import * as React from "react";
import { blueGray, blueGrayL3 } from "sourcegraph/components/utils/colors";
import { weight } from "sourcegraph/components/utils/typography";

interface Props {
	style?: React.CSSProperties;
	shortcut: string;
}

const defaultKeyStyle = {
	backgroundColor: blueGrayL3(),
	border: "solid 1px rgba(201, 211, 227, 0.57)",
	borderRadius: 3,
	boxShadow: "0 2px 0 0 rgba(201, 211, 227, 0.6)",
	color: blueGray(),
	display: "inline-block",
	textTransform: "uppercase",
	width: "23px",
	height: "21.8px",
	lineHeight: "21.8px",
	fontSize: 13,
	textAlign: "center",
	verticalAlign: "middle",
	fontWeight: weight[2],
};

export function Key({ style, shortcut }: Props): JSX.Element {
	return <span style={Object.assign({}, defaultKeyStyle, style)}>{shortcut}</span>;
}
