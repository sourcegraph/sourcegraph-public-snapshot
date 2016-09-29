import * as React from "react";

import {Base} from "sourcegraph/components/Base";
import {colors, typography, whitespace} from "sourcegraph/components/utils/index";

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
	children?: any;
	level?: number; //  1 is the largest
	underline?: Color;
	color: Color;
	align?: "left" | "right" | "center"; // left, right, center
	style?: Object;
}

enum Color {
	"blue",
	"purple",
	"white",
	"orange",
	"green",
	"gray"
}

export const Heading = (props: Props): any => {
	const headingColors = {
		blue: colors.blue(),
		purple: colors.purple(),
		white: colors.white(),
		orange: colors.orange(),
		green: colors.green(),
		gray: colors.coolGray3(),
	};

	const sx = Object.assign({},
		typography.size[props.level ? props.level : 3],
		{
			color: headingColors[props.color],
			fontWeight: typography.weight[2],
			textTransform: props.level === 7 ? "uppercase" : "auto",
			textAlign: props.align,
		},
	);

	const underlineSx = {
		borderColor: headingColors[props.underline ? props.underline : "white"],
		borderWidth: "4px",
		display: "inline-block",
		width: "6rem",
		marginBottom: whitespace[3],
		marginTop: whitespace[3],
	};

	return <Base
		m={props.m}
		mt={props.mt}
		mb={props.mb || props.mb === 0 ? props.mb : 2}
		ml={props.ml}
		mr={props.mr}
		my={props.my}
		mx={props.mx}
		p={props.p}
		pt={props.pt}
		pb={props.pb}
		pl={props.pl}
		pr={props.pr}
		py={props.py}
		px={props.px}
		style={sx}>
		{props.children} <br />
		{props.underline && <hr style={underlineSx} />}
	</Base>;
};
