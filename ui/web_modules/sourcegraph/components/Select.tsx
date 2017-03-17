import * as React from "react";
import { FormElement } from "sourcegraph/components/FormElement";
import { ChevronDown } from "sourcegraph/components/symbols/Primaries";
import { colors, forms, whitespace } from "sourcegraph/components/utils";

interface Props extends React.HTMLAttributes<HTMLSelectElement> {
	block?: boolean;
	children?: React.ReactNode[];
	containerStyle?: React.CSSProperties;
	defaultValue?: string;
	label?: string;
	helperText?: string;
	error?: boolean;
	errorText?: string;
	optionalText?: string;
	placeholder?: string;
	style?: React.CSSProperties;
}

export function Select(props: Props): JSX.Element {
	const {
		block = true,
		errorText,
		helperText,
		label,
		optionalText,

		children,
		containerStyle,
		defaultValue,
		error,
		placeholder,
		style,

		...selectProps
	} = props;

	const formElProps = { block, errorText, helperText, label, optionalText };

	const sx = {
		color: "inherit",
		marginTop: whitespace[2],
		padding: `${whitespace[2]} ${whitespace[3]}`,
		fontFamily: "inherit",
		width: block && "100%",
		boxShadow: error && forms.error,
		...style,
	};

	return <FormElement {...formElProps} style={containerStyle}>
		<select
			{...selectProps}
			{...forms.style}
			style={sx}
			required={true}
			defaultValue={defaultValue ? defaultValue : ""}>
			{placeholder && <option value="" disabled={true}>{placeholder}</option>}
			{children}
		</select>
		<ChevronDown color={colors.blueGray()} width={24} style={{ marginLeft: "-32px" }} />
	</FormElement>;
}
