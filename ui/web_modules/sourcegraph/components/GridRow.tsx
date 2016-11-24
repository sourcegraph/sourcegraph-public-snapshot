import * as React from "react";
import { clearFix } from "sourcegraph/components/utils/layout";

interface Props {
	style?: React.CSSProperties;
	children?: JSX.Element[];
}

export function GridRow(props: Props): JSX.Element {
	return (
		<div
			{...clearFix}
			style={props.style}>
			{props.children}
		</div>
	);
}
