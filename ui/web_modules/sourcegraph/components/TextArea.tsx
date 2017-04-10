import * as React from "react";
import { FormElement } from "sourcegraph/components/FormElement";
import { colors, forms, whitespace } from "sourcegraph/components/utils";

interface Props extends React.HTMLAttributes<HTMLTextAreaElement> {
	block?: boolean;
	className?: string;
	// domRef is like ref, but it is called with the <input> DOM element,
	// not this pure wrapper component. <Input domRef={...}> is equivalent
	// to <input ref={...}>.
	domRef?: (c: HTMLTextAreaElement) => void;
	label?: string;
	helperText?: string;
	error?: boolean;
	errorText?: string;
	optionalText?: string;
	style?: React.CSSProperties;
	containerStyle?: React.CSSProperties;
}

export function TextArea(props: Props): JSX.Element {
	const {
		block,
		className,
		errorText,
		helperText,
		label,
		optionalText,

		domRef,
		children,
		containerStyle,
		error,
		style,

		...textareaProps
	} = props;

	const formElProps = { block, className, errorText, helperText, label, optionalText };

	const textareaSx = {
		display: block ? "block" : "inline-block",
		padding: `${whitespace[2]} ${whitespace[3]}`,
		width: block && "100%",
		boxShadow: error && forms.error,
		background: colors.white(),
		...style,
	};

	return <FormElement {...formElProps} style={containerStyle}>
		<textarea {...textareaProps} {...forms.style} style={textareaSx} ref={domRef}>
			{children}
		</textarea>
	</FormElement>;
}
