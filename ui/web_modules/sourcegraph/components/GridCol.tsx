import * as React from "react";
import {breakpoints} from "sourcegraph/components/utils/layout";

import {media, merge} from "glamor";

interface Props {
	align?: "left" | "right";
	col: number;
	colSm?: number;
	colMd?: number;
	colLg?: number;
	children?: JSX.Element[];
	style?: React.CSSProperties;
}

export function GridCol(props: Props): JSX.Element {
	// Generate column sizes
	const colSize: number = 100 / 12;
	const unit = "%";
	const colSizes = Array.from(Array(13), (_, i) => colSize * i);
	const column = colSizes.map((val) => val + unit );

	const sx = merge(
		media(breakpoints["sm"], { width: props.colSm ? column[props.colSm] : "" }),
		media(breakpoints["md"], { width: props.colMd ? column[props.colMd] : "" }),
		media(breakpoints["lg"], { width: props.colLg ? column[props.colLg] : "" }),
		{
			boxSizing: "border-box",
			width: column[props.col],
			float: props.align,
		},
		props.style || {},
	);

	return <div style={props.style}
		{...sx}>
		{props.children}
	</div>;
}
