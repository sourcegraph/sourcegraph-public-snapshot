import { css } from "glamor";
import * as React from "react";
import { Alert } from "sourcegraph/components/symbols";
import { ChevronDown } from "sourcegraph/components/symbols/Zondicons";
import { colors, typography, whitespace } from "sourcegraph/components/utils";

interface Props {
	block?: boolean;
	children?: React.ReactNode[];
	containerSx?: React.CSSProperties;
	label?: string;
	placeholder?: string;
	helperText?: string;
	error?: boolean;
	errorText?: string;
	style?: React.CSSProperties;
	defaultValue?: string;
}

export function Select({block = true, containerSx, children, label, placeholder, helperText, error, errorText, style, defaultValue}: Props): JSX.Element {
	const sx = css({
		appearance: "none",
		backgroundColor: "white",
		borderColor: error ? colors.red() : colors.blueGray(),
		borderRadius: 3,
		borderWidth: 1,
		boxSizing: "border-box",
		marginTop: whitespace[2],
		padding: "8px 38px 8px 15px",
		transition: "all 0.25s ease-in-out",
		fontFamily: "inherit",
		fontWeight: "inherit",
		outline: "none",
		width: block ? "100%" : "",
	});
	return <div style={containerSx}>
		{label && <div>{label} <br /></div>}
		<select
			{...sx}
			style={style}
			required={true}
			defaultValue={defaultValue}
			placeholder={placeholder ? placeholder : ""}>
			{children}
		</select>
		<ChevronDown style={{ fill: colors.blueGray(), marginLeft: "-28px" }} width={11} />
		{helperText && <em { ...css({
			display: "block",
			marginTop: whitespace[2],
		}, typography.small) }> {helperText}</em>}
		{errorText &&
			<div style={{
				color: colors.red(),
				marginTop: whitespace[2],
				marginBottom: whitespace[2],
			}}>
				<Alert width={16} style={{
					fill: "currentColor",
					marginRight: whitespace[2],
					marginTop: -4
				}} />
				This is an error message.
			</div>
		}
	</div>;
}
