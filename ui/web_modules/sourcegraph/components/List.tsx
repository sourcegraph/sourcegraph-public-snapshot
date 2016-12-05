import * as React from "react";
import { whitespace } from "sourcegraph/components/utils";

interface Props {
	children?: React.ReactElement<any>[];
	style?: React.CSSProperties;
	itemStyle?: React.CSSProperties;
}

export function List({children, style, itemStyle}: Props): JSX.Element {
	const sx = Object.assign(
		{ paddingLeft: whitespace[4] },
		style,
	);

	const itemSx = Object.assign(
		{ marginBottom: whitespace[3] },
		itemStyle,
	);

	if (!children || !children.length || children.length === 0) { return <li></li>; };

	const listItems = children.map((child, i) => {
		return child.type === "li"
			? <li key={i} style={itemSx}>{child.props.children}</li>
			: child;
	});

	return <ul style={sx}>{listItems}</ul>;
}
