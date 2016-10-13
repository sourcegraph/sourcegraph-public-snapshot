import {style as gStyle} from "glamor";
import * as React from "react";
import {CoachMark, Panel} from "sourcegraph/components/index";
import {whitespace} from "sourcegraph/components/utils";

interface Props {
	color?: "blue" | "purple" | "orange" | "green";
	children?: React.ReactNode[];
	pulseColor?: "blue" | "purple" | "orange" | "green" | "white";
	open: boolean;
	containerStyle?: React.CSSProperties;
	tooltipStyle?: React.CSSProperties;
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
}: Props): JSX.Element {

	const sx = gStyle(Object.assign({},
		{ position: "relative" },
		containerStyle,
	));

	const tooltipSx = Object.assign({},
		{
			padding: whitespace[3],
			position: "absolute",
			top: 22.5,
			left: 22.5,
			maxWidth: "250px",
		},
		tooltipStyle,
	);

	return <div {...sx}>
		<CoachMark color={color} pulseColor={pulseColor} active={active} />
		{open && <Panel hoverLevel="low" style={tooltipSx}>{children}</Panel>}
	</div>;
};
