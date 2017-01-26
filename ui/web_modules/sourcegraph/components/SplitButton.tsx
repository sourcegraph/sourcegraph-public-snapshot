import * as classNames from "classnames";
import { css } from "glamor";
import * as omit from "lodash/omit";
import * as React from "react";

import { Button, FlexContainer } from "sourcegraph/components";
import { colors, whitespace } from "sourcegraph/components/utils";

import { sizeSx } from "sourcegraph/components/Button";

interface SplitButtonProps extends React.ClassAttributes<HTMLButtonElement> {
	block?: boolean;
	children?: React.ReactNode[];
	size?: "tiny" | "small" | "large";
	color?: "white" | "blue" | "green" | "orange" | "purple" | "red" | "blueGray";
	className?: string;
	secondaryText?: string;
	style?: React.CSSProperties;
}

export function SplitButton(props: SplitButtonProps): JSX.Element {

	const {
		block = false,
		children,
		className,
		color = "blueGray",
		secondaryText,
		size,
		style,
	} = props;

	const transferredProps = omit(props, ["secondaryText"]);

	const btnHoverColor = colors[`${color}D1`]();
	const secondaryColor = colors[`${color}D1`]();

	const containerSx = css({
		":hover button": { backgroundColor: btnHoverColor },
		":hover": { color: "white !important" },
	}).toString();

	return <FlexContainer items="stretch"
		className={classNames(containerSx, className)}
		style={Object.assign({
			backgroundColor: secondaryColor,
			borderRadius: 3,
			color: "white",
			cursor: "pointer",
			fontWeight: "bold",
			overflow: "hidden"
		},
			!block ? { display: "inline-block" } : {},
			style,
		)}>
		<Button {...transferredProps} {...css({ flex: 2 }) } style={{
			borderTopRightRadius: 0,
			borderBottomRightRadius: 0,
			margin: 0,
		}}>
			{children}
		</Button>
		<div {...css(
			{
				display: "inline-block",
				flex: 1,
				paddingLeft: whitespace[3],
				paddingRight: whitespace[3],
				paddingTop: "0.45rem",
				paddingBottom: "0.42rem",
				textAlign: "center",
			},
			size ? sizeSx[size] : {},
		) }>{secondaryText}</div>
	</FlexContainer>;
};
