import * as React from "react";

import {Base} from "sourcegraph/components/Base";
import {breakpoints} from "sourcegraph/components/utils/layout";

import {media, merge} from "glamor";

interface Props {
	m?: number;
	mt?: number;
	mb?: number;
	ml?: number;
	mr?: number;
	my?: number;
	mx?: number;
	p?: number;
	pt?: number;
	pb?: number;
	pl?: number;
	pr?: number;
	py?: number;
	px?: number;
	align?: "left" | "right";
	col: number;
	colSm?: number;
	colMd?: number;
	colLg?: number;
	children?: Array<JSX.Element>;
	className?: string;
	style?: Object;
}

export function GridCol(props: Props): JSX.Element {

	// Generate column sizes
	const colSize: number = 8.333333;
	const unit = "%";
	const colSizes = Array.from(Array(12), (_, i) => colSize * i);
	const column = colSizes.map((val) => val + unit );

	const sx = merge(
		media(breakpoints["sm"], { width: props.colSm ? column[props.colSm] : "" }),
		media(breakpoints["md"], { width: props.colMd ? column[props.colMd] : "" }),
		media(breakpoints["lg"], { width: props.colLg ? column[props.colLg] : "" }),
		{
			boxSizing: "border-box",
			width: column[props.col],
		},
		props.style ? props.style : {},
	);

	return <Base
		{...sx}
		align={props.align}
		className={props.className}>
			{props.children}
	</Base>;
}
