import {hover, style} from "glamor";
import * as React from "react";

interface Props {
	className?: string;
	width?: number; // appended by "px"
	style?: Object;
	color?: any;
	viewBox?: string;
	children?: any;
}

export const Symbol = (props: Props) => {
	const sx = Object.assign({},
		{
			display: "inline",
			verticalAlign: "middle",
			position: "relative",
			top: -1,
		},
		props.style
	);
	return <svg
		{...style({fill: props.color ? props.color : "currentColor"})}
		{...hover({fill: "currentColor"})}
		className={props.className}
		width={`${props.width ? props.width : 16}px`}
		style={sx}
		viewBox={props.viewBox}>{props.children}</svg>;
};
