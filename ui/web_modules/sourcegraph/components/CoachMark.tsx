import {hover, keyframes, style as gStyle} from "glamor";
import * as React from "react";
import {colors} from "sourcegraph/components/utils";

interface Props {
	color?: "blue" | "purple" | "orange" | "green";
	pulseColor?: "blue" | "purple" | "orange" | "green" | "white";
	style?: React.CSSProperties;
	active: boolean;
}

const hoverSx = hover({
	transform: "scale(1.2)",
});

export function CoachMark ({
	color = "blue",
	pulseColor = "blue",
	active = false,
	style,
}: Props): JSX.Element {

	const baseShadow = `0 2px 7px ${colors.black(0.3)}`;
	const pulseShadowStart = `0 0 0 0 ${colors[pulseColor](0.3)}`;
	const pulseShadowFinish = `0 0 0 50px ${colors[pulseColor](0)}`;

	const bounce = keyframes({
			"100%": { boxShadow: `${baseShadow}, ${pulseShadowFinish}` },
	});

	const sx = gStyle(Object.assign({},
		{
			animation: active ? `${bounce} 2.5s infinite` : "",
			background: colors[color](),
			borderRadius: "50%",
			border: "4px solid white",
			boxShadow: `${baseShadow}, ${pulseShadowStart}`,
			cursor: "pointer",
			transition: "transform 0.2s ease-in-out",
			height: 18,
			width: 18,
		},
		style,
	));

	return <div {...sx} {...hoverSx}></div>;
};
