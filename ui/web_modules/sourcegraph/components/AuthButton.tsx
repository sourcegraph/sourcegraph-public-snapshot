import * as React from "react";
import { Button, SplitButton } from "sourcegraph/components";
import { ButtonProps } from "sourcegraph/components/Button";
import { GitHubLogo, Google } from "sourcegraph/components/symbols";
import { colors, typography, whitespace } from "sourcegraph/components/utils";
import { AuthProps, getAuthAction } from "sourcegraph/user/Auth";

interface Props extends ButtonProps {
	iconType: "github" | "google";
	secondaryText?: string;
	authInfo: AuthProps;
}

export function AuthButton(props: Props): JSX.Element {
	const {
		iconType,
		secondaryText,
		children,
		authInfo,
		...btnProps
	} = props;

	const iconSx = props.size === "small" ? typography.size[5] : typography.size[4];
	const iconSize = 24;

	const icon = <span>
		{iconType === "github" && <GitHubLogo width={iconSize} style={iconSx} />}
		{iconType === "google" && <Google width={iconSize} style={iconSx} />}
	</span>;

	const { submit, form } = getAuthAction(authInfo);

	if (secondaryText) {
		return <SplitButton onClick={submit} {...btnProps} secondaryText={secondaryText}>
			{form}
			{icon}
			{children}
		</SplitButton>;
	}

	return <Button {...btnProps} onClick={submit} style={{
		padding: 0,
		textAlign: "left",
	}}>
		{form}
		<span style={{
			backgroundColor: colors.black(0.25),
			display: "inline-block",
			padding: whitespace[2],
		}}>{icon}</span>
		<span style={{
			marginLeft: whitespace[3],
			marginRight: whitespace[3],
		}}>{children}</span>
	</Button>;
}
