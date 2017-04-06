import { css } from "glamor";
import * as React from "react";

import { Loader } from "sourcegraph/components/Loader";
import { colors, typography, whitespace } from "sourcegraph/components/utils";

export interface ButtonProps extends React.HTMLFactory<HTMLButtonElement> {
	animation?: boolean;
	children?: React.ReactNode[];
	backgroundColor?: String;
	block?: boolean;
	disabled?: boolean;
	outline?: boolean;
	size?: "tiny" | "small" | "large";
	loading?: boolean;
	color?: "white" | "blue" | "green" | "orange" | "purple" | "red" | "blueGray";
	onClick?: React.EventHandler<React.MouseEvent<HTMLButtonElement>>;
	className?: string;
	style?: React.CSSProperties;

	type?: "button" | "submit" | "reset";
	id?: string;
	tabIndex?: number;
}

export const sizeSx = {
	tiny: css({
		paddingBottom: "0.14rem",
		paddingTop: "0.2rem",
		paddingLeft: whitespace[2],
		paddingRight: whitespace[2],
	}, typography.size[7]),
	small: css({
		paddingBottom: whitespace[1],
		paddingTop: whitespace[1],
	}, typography.size[6]),
	large: css({
		paddingBottom: whitespace[2],
		paddingTop: whitespace[2],
		paddingLeft: whitespace[3],
		paddingRight: whitespace[3],
	}, typography.size[4]),
};

export function Button(props: ButtonProps): JSX.Element {
	const {
		block = false,
		outline = false,
		size,
		loading = false,
		color = "blueGray",
		children,
		type = "button",
		animation = true,
		backgroundColor,
		...transferredProps,
	} = props;

	const btnColor = colors[color]();
	const btnHoverColor = color === "white" ? colors.blueD1() : colors[`${color}D1`]();
	const btnActiveColor = color === "white" ? colors.blueD2() : colors[`${color}D2`]();

	const outlineSx = css(
		{
			backgroundColor: "transparent",
			borderColor: color === "blueGray" ? colors.blueGrayL2(0.6) : btnColor,
			color: color === "blueGray" ? colors.blue() : btnColor,
		},
		{
			":hover": {
				backgroundColor: "transparent",
				borderColor: color === "blueGray" ? colors.blueGrayL2() : btnHoverColor,
				color: color === "blueGray" ? colors.blueD1() : btnHoverColor,
			}
		},
		{
			":active": {
				backgroundColor: "transparent",
				borderColor: btnActiveColor,
				color: btnActiveColor,
			}
		},

	);

	const whiteSx = css(
		{
			color: colors.blue(),
			backgroundColor: "white",
		},
		{
			":hover":
			{
				backgroundColor: "white",
				color: btnHoverColor,
			},
		},
		{ ":active": { color: btnActiveColor } }
	);

	return <button
		type={type}
		{...transferredProps}
		{...css(
			{
				borderWidth: outline ? 2 : 0,
				borderStyle: "solid",
				borderColor: "transparent",
				backgroundColor: backgroundColor || btnColor,
				color: "white",
				textAlign: "center",
				fontWeight: "bold",
				outline: "none",
				paddingLeft: whitespace[3],
				paddingRight: whitespace[3],
				paddingTop: "0.45rem",
				paddingBottom: "0.42rem",
				transition: animation ? "all 0.4s" : "none",
				borderRadius: 4,
				boxSizing: "border-box",
				cursor: "pointer",
				userSelect: "none",
				overflow: "hidden",
				display: block ? "block" : "inline-block",
				width: block ? "100%" : "auto",
			},
			{ ":hover": { backgroundColor: btnHoverColor } },
			{ ":active": { backgroundColor: btnActiveColor } },
			size ? sizeSx[size] : {},
			outline ? outlineSx : {},
			color === "white" ? whiteSx : {},
			{
				"[disabled]": {
					backgroundColor: colors.blueGrayL2(),
					color: colors.blueGray(0.7),
					cursor: "not-allowed",
				}
			}
		)
		}>
		{loading && <Loader />}
		{!loading && children}
	</button>;
}
