import {style as gStyle} from "glamor";
import * as React from "react";
import {CoachMark, Panel} from "sourcegraph/components/index";

interface Props {
	color?: "blue" | "purple" | "orange" | "green";
	children?: React.ReactNode[];
	pulseColor?: "blue" | "purple" | "orange" | "green" | "white";
	open: boolean;
	markOnClick?: () => void;
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
	markOnClick,
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
			opacity: open ? 1 : 0,
			display: open ? "hidden" : "block",
			transform: `scale(${ open ? 1 : 0})`,
			transformOrigin: `${annotationPosition === "right" ? "left" : "right"} top`,
			top: 22.5,
			left: leftOffset,
			maxWidth: 350,
			transition: "all 300ms cubic-bezier(0.500, -0.405, 0.205, 1.345)",
		},
		tooltipStyle,
	);

	return <div {...sx}>
		<CoachMark color={color} pulseColor={pulseColor} active={active} onClick={markOnClick} />
		<Panel hoverLevel="low" style={tooltipSx}>{children}</Panel>
	</div>;
};
