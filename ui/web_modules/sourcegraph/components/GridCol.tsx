import {media, merge} from "glamor";
import * as React from "react";
import {breakpoints} from "sourcegraph/components/utils/layout";

interface Props {
	align?: "left" | "right" | "none";
	col: number;
	colSm?: number;
	colMd?: number;
	colLg?: number;
	children?: JSX.Element[];
	style?: React.CSSProperties;
}

export function GridCol({
	align = "none",
	col,
	colSm,
	colMd,
	colLg,
	children,
	style,
}: Props): JSX.Element {
	// Generate column sizes
	const colSize: number = 100 / 12;
	const unit = "%";
	const colSizes = Array.from(Array(13), (_, i) => colSize * i);
	const column = colSizes.map((val) => val + unit );

	const sx = merge(
		media(breakpoints["sm"], { width: colSm ? column[colSm] : "" }),
		media(breakpoints["md"], { width: colMd ? column[colMd] : "" }),
		media(breakpoints["lg"], { width: colLg ? column[colLg] : "" }),
		{
			boxSizing: "border-box",
			width: column[col],
			float: align,
		},
		style || {},
	);

	return <div style={style}
		{...sx}>
		{children}
	</div>;
}
