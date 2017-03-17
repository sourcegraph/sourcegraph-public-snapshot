import * as React from "react";
import { FlexContainer, Heading } from "sourcegraph/components";
import { colors, typography, whitespace } from "sourcegraph/components/utils";

interface Props extends React.HTMLAttributes<HTMLInputElement> {
	block?: boolean;
	className?: string;
	label?: string;
	helperText?: string;
	errorText?: string;
	optionalText?: string;
	style?: React.CSSProperties;
}

export function FormElement(props: Props): JSX.Element {
	const {
		block,
		children,
		label,
		helperText,
		errorText,
		style,
		className,
		optionalText,
	} = props;

	return <div className={className} style={{
		marginBottom: whitespace[3],
		position: "relative",
		width: block && "100%",
		...style
	}}>
		<FlexContainer justify="between">
			{label && <div style={{ marginBottom: whitespace[2] }}>{label}</div>}
			{optionalText && <Heading level={7} color="gray" style={{ marginTop: 0 }}>{optionalText}</Heading>}
		</FlexContainer>
		{children}
		{helperText && <Context type="helper" text={helperText} />}
		{errorText && <Context type="error" text={errorText} />}
	</div>;
}

export function Context({ type, text }: { type: "helper" | "error", text: string }): JSX.Element {
	return <div style={{
		color: type === "error" ? colors.orange() : colors.blueGray(),
		fontStyle: type === "helper" && "italic",
		marginBottom: whitespace[2],
		marginTop: whitespace[2],
		...typography.small,
	}}>{text}</div>;
}
