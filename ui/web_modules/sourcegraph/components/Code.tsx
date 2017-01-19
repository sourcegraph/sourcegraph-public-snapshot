import * as React from "react";
import { colors, typography, whitespace } from "sourcegraph/components/utils";

interface Props {
	children?: React.ReactNode[];
	style?: React.CSSProperties;
}

export function Code({ children, style }: Props): JSX.Element {
	return <span style={Object.assign({
		backgroundColor: colors.blueGrayL3(),
		borderRadius: 3,
		color: colors.text(),
		fontFamily: typography.fontStack.code,
		padding: whitespace[1],
		marginRight: whitespace[1],
	}, typography.size[7], style)}>{children}</span>;
}
