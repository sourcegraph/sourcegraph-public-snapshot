import * as React from "react";

import { RouterLocation } from "sourcegraph/app/router";
import { AuthButton } from "sourcegraph/components/AuthButton";
import { ButtonProps } from "sourcegraph/components/Button";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";

interface Props extends ButtonProps {
	scope: "private" | "public";
	returnTo?: string | RouterLocation;
	newUserReturnTo?: string | RouterLocation;
	pageName?: string;
	secondaryText?: string;
}

export function GitHubAuthButton(props: Props): JSX.Element {
	const {
		scope,
		children,
		...transferredProps,
	} = props;

	const scopes = scope === "private" ?
		"read:org,user:email,repo" :
		"read:org,user:email";

	return <AuthButton
		provider="github"
		iconType="github"
		eventObject={Events.OAuth2FlowGitHub_Initiated}
		scopes={scopes}
		{...transferredProps}>
		{children}
	</AuthButton>;
}
