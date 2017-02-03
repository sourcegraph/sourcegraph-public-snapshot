import * as classNames from "classnames";
import { css } from "glamor";
import * as React from "react";

import { Button, FlexContainer } from "sourcegraph/components";
import { ButtonProps } from "sourcegraph/components/Button";
import { colors, whitespace } from "sourcegraph/components/utils";

import { sizeSx } from "sourcegraph/components/Button";

interface SplitButtonProps extends ButtonProps {
	secondaryText?: string;
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
		onClick,
		...transferredProps,
	} = props;

	const btnHoverColor = colors[`${color}D1`]();
	const secondaryColor = colors[`${color}D1`]();

	const containerSx = css({
		":hover button": { backgroundColor: btnHoverColor },
		":hover": { color: "white !important" },
	}).toString();

	return <FlexContainer items="stretch"
		onClick={onClick}
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
		<Button
			{...transferredProps}
			{...css({ flex: 2 }) }
			block={block}
			color={color}
			size={size}
			style={{
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
