import { style } from "glamor";
import * as React from "react";
import { Alert } from "sourcegraph/components/symbols";
import { colors, typography, whitespace } from "sourcegraph/components/utils";

interface Props extends React.HTMLAttributes<HTMLInputElement> {
	block?: boolean;
	className?: string;
	// domRef is like ref, but it is called with the <input> DOM element,
	// not this pure wrapper component. <Input domRef={...}> is equivalent
	// to <input ref={...}>.
	domRef?: (c: HTMLInputElement) => void;
	label?: string;
	placeholder?: string;
	helperText?: string;
	error?: boolean;
	errorText?: string;
	style?: React.CSSProperties;
	containerStyle?: React.CSSProperties;
	inputSize?: "small";
}

export function Input(props: Props): JSX.Element {
	const other = Object.assign({}, props);
	delete other.block;
	delete other.className;
	delete other.domRef;
	delete other.label;
	delete other.placeholder;
	delete other.helperText;
	delete other.error;
	delete other.errorText;
	delete other.style;
	delete other.inputSize;

	const errorTextSx = Object.assign(
		{
			display: "block",
			marginTop: whitespace[2],
		},
		typography.size[7],
	);

	const inputSx = Object.assign(
		{
			appearance: "none",
			borderRadius: 3,
			backgroundColor: "white",
			borderColor: props.error ? colors.red1() : colors.coolGray3(0.3),
			borderWidth: 1,
			borderStyle: "solid",
			boxSizing: "border-box",
			color: colors.coolGray2(),
			display: props.block ? "block" : "inline-block",
			paddingBottom: props.inputSize === "small" ? whitespace[1] : whitespace[2],
			paddingLeft: props.inputSize === "small" ? "0.8rem" : whitespace[3],
			paddingRight: props.inputSize === "small" ? "0.8rem" : whitespace[3],
			paddingTop: props.inputSize === "small" ? "0.3rem" : whitespace[2],
			transition: "all 0.25s ease-in-out",
			width: props.block ? "100%" : null,
		},
		props.style,
		props.inputSize === "small" ? typography.size[7] : null,
	);

	const placeholderSx = { color: colors.coolGray3(0.7) };

	const sx = Object.assign(
		{
			width: props.block ? "100%" : null,
		},
		props.containerStyle,
	);

	return <div className={props.className} style={sx}>
		{props.label && <div style={{ marginBottom: whitespace[2] }}>{props.label}</div>}
		<input {...other} style={inputSx} ref={props.domRef} placeholder={props.placeholder}
			{...style({
				":focus": {
					borderColor: `${colors.coolGray3(0.7)} !important`,
					outline: "none",
				},
				"::-webkit-input-placeholder": placeholderSx,
				"::-moz-placeholder": placeholderSx,
				":-moz-placeholder": placeholderSx,
				":-ms-input-placeholder": placeholderSx,
			}) }
			/>
		{props.helperText && <em style={errorTextSx}>{props.helperText}</em>}
		{props.errorText && <div style={{
			color: colors.redText(),
			marginBottom: whitespace[2],
			marginTop: whitespace[2],
		}}>
			<Alert width={16} style={{
				fill: colors.redText(),
				marginRight: whitespace[2],
				marginTop: -4,
			}} />
			{props.errorText}
		</div>}
	</div>;
}
