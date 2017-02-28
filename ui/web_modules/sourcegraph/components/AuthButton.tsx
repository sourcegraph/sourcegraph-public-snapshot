import * as React from "react";
import { Button, SplitButton } from "sourcegraph/components";
import { ButtonProps } from "sourcegraph/components/Button";
import { GitHubLogo, Google } from "sourcegraph/components/symbols";
import { typography, whitespace } from "sourcegraph/components/utils";
import { AuthProps, getAuthAction } from "sourcegraph/user/Signup";

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

	const icon = <span style={{ marginRight: whitespace[2] }}>
		{iconType === "github" && <GitHubLogo style={iconSx} />}
		{iconType === "google" && <Google style={iconSx} />}
	</span>;

	const { submit, form } = getAuthAction(authInfo);

	if (secondaryText) {
		return <SplitButton onClick={submit} {...btnProps} secondaryText={secondaryText}>
			{form}
			{icon}
			{children}
		</SplitButton>;
	}

	return <Button onClick={submit} {...btnProps}>
		{form}
		{icon}
		{children}
	</Button>;
}
