import {style as gStyle} from "glamor";
import * as React from "react";
import {CoachMark, Panel} from "sourcegraph/components/index";

interface Props {
	color?: "blue" | "purple" | "orange" | "green";
	children?: React.ReactNode[];
	pulseColor?: "blue" | "purple" | "orange" | "green" | "white";
	open: boolean;
	containerStyle?: React.CSSProperties;
	tooltipStyle?: React.CSSProperties;
	annotationPosition?: "left" | "right";
	active: boolean;
}

export function Annotation ({
	color = "blue",
	pulseColor = "blue",
	active = false,
	open = true,
	children,
	containerStyle,
	tooltipStyle,
	annotationPosition = "right",
}: Props): JSX.Element {

	const sx = gStyle(Object.assign({},
		{ position: "relative" },
		containerStyle,
	));

	let leftOffset = "22.5px";
	if (annotationPosition === "left") {
		leftOffset = (tooltipStyle && tooltipStyle["width"]) ? "-" + tooltipStyle["width"] : "-350px";
	}

	const tooltipSx = Object.assign({},
		{
			position: "absolute",
			top: 22.5,
			left: leftOffset,
			maxWidth: 350,
		},
		tooltipStyle,
	);

	return <div {...sx}>
		<CoachMark color={color} pulseColor={pulseColor} active={active} />
		{open && <Panel hoverLevel="low" style={tooltipSx}>{children}</Panel>}
	</div>;
};
