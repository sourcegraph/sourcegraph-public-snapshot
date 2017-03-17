import * as React from "react";

import { AuthButton } from "sourcegraph/components/AuthButton";
import { ButtonProps } from "sourcegraph/components/Button";

interface Props extends ButtonProps {
	privateCode: boolean;
	pageName?: string;
	secondaryText?: string;
}

export function GitHubAuthButton(props: Props): JSX.Element {
	const {
		color = "blue",
		children,
		pageName,
		...transferredProps,
	} = props;

	return <AuthButton
		color={color}
		{...transferredProps}>
		{children}
	</AuthButton>;
}
