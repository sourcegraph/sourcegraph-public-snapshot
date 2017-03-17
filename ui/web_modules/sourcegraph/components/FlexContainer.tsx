import * as React from "react";

interface Props extends React.DOMAttributes<HTMLElement> {
	direction?: "left-right" | "right-left" | "top-bottom" | "bottom-top";
	wrap?: boolean;
	justify?: "start" | "end" | "center" | "between" | "around";
	items?: "start" | "end" | "center" | "baseline" | "stretch";
	content?: "start" | "end" | "center" | "between" | "around" | "stretch";
	className?: string;
	children?: React.ReactNode[];
	style?: React.CSSProperties;
}

export function FlexContainer(props: Props): JSX.Element {
	const {
		direction = "left_right",
		wrap = false,
		justify = "start",
		items = "stretch",
		content = "stretch",
		className,
		children,
		style,
		...transferredProps,
	} = props;

	return <div className={className} {...transferredProps} style={Object.assign({
		display: "flex",
		flexWrap: wrap ? "wrap" : "nowrap",
		flexDirection: directionAttrs[direction],
		alignContent: contentAttrs[content],
		alignItems: itemsAttrs[items],
		justifyContent: justifyAttrs[justify],
	}, style)}>{children}</div>;
}

const directionAttrs = {
	"left-right": "row",
	"right-left": "row-reverse",
	"top-bottom": "column",
	"bottom-top": "column-reverse",
};

const justifyAttrs = {
	"start": "flex-start",
	"end": "flex-end",
	"center": "center",
	"between": "space-between",
	"around": "space-around",
};

const itemsAttrs = {
	"start": "flex-start",
	"end": "flex-end",
	"center": "center",
	"baseline": "baseline",
	"stretch": "stretch",
};

const contentAttrs = {
	"start": "flex-start",
	"end": "flex-end",
	"center": "center",
	"around": "space-around",
	"between": "space-between",
	"stretch": "stretch",
};
