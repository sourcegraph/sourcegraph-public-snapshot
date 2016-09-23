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
		{ verticalAlign: "middle" },
		props.style
	);
	return <svg
		{...style({fill: props.color})}
		{...hover({fill: "inherit"})}
		className={props.className}
		width={`${props.width ? props.width : 16}px`}
		style={sx}
		viewBox={props.viewBox}>{props.children}</svg>;
};
