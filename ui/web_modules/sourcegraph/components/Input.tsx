import * as React from "react";
import { FormElement } from "sourcegraph/components/FormElement";
import * as Icons from "sourcegraph/components/symbols/Primaries";
import { colors, forms, typography, whitespace } from "sourcegraph/components/utils";

interface Props extends React.HTMLAttributes<HTMLInputElement> {
	block?: boolean;
	className?: string;
	// domRef is like ref, but it is called with the <input> DOM element,
	// not this pure wrapper component. <Input domRef={...}> is equivalent
	// to <input ref={...}>.
	domRef?: (c: HTMLInputElement) => void;
	label?: string;
	helperText?: string;
	error?: boolean;
	errorText?: string;
	optionalText?: string;
	style?: React.CSSProperties;
	containerStyle?: React.CSSProperties;
	compact?: boolean;
	icon?: string; // Choose from Primaries icon names
	iconPosition?: "right";
}

export function Input(props: Props): JSX.Element {
	const {
		block,
		className,
		errorText,
		helperText,
		label,
		optionalText,

		compact,
		containerStyle,
		domRef,
		error,
		icon,
		iconPosition,
		style,

		...inputProps
	} = props;

	const formElProps = { block, className, errorText, helperText, label, optionalText };

	const inputSx = {
		display: block ? "block" : "inline-block",
		paddingBottom: compact ? whitespace[1] : whitespace[2],
		paddingTop: compact ? whitespace[1] : whitespace[2],
		paddingLeft: icon && !iconPosition ? "2.5rem" : whitespace[3],
		paddingRight: icon && iconPosition ? "2.5rem" : whitespace[3],
		width: block && "100%",
		boxShadow: error && forms.error,
		...style,
		...compact ? typography.size[7] : {},
	};

	return <FormElement {...formElProps} style={containerStyle}>
		<div style={{ position: "relative" }}>
			<input {...inputProps} {...forms.style} style={inputSx} ref={domRef} />
			{icon && Icons[icon]({
				color: colors.blueGrayL1(),
				width: compact ? 16 : 24,
				style: {
					margin: whitespace[2],
					position: "absolute",
					right: iconPosition && 0,
					top: 0,
				},
			})}
		</div>
	</FormElement>;
}
