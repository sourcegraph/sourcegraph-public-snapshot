import * as React from "react";
import { colors } from "sourcegraph/components/utils";

interface Props {
	style?: React.CSSProperties;
	shortcut: string;
}

export function Key({style, shortcut}: Props): JSX.Element {
	return <span style={Object.assign({
		backgroundColor: colors.black(0.3),
		borderRadius: 3,
		lineHeight: 1,
		display: "inline-block",
		padding: "3px 5px",
	}, style)}>{shortcut}</span>;
}
