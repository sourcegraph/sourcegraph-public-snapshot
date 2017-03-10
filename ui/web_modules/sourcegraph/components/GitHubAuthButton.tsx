import * as React from "react";

import { RouterLocation } from "sourcegraph/app/router";
import { AuthButton } from "sourcegraph/components/AuthButton";
import { ButtonProps } from "sourcegraph/components/Button";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";

interface Props extends ButtonProps {
	scope: "private" | "public";
	pageName?: string;
	returnTo?: string | RouterLocation;
	newUserReturnTo?: string | RouterLocation;
	secondaryText?: string;
}

export function GitHubAuthButton(props: Props): JSX.Element {
	const {
		color = "blue",
		scope,
		children,
		returnTo,
		newUserReturnTo,
		pageName,
		...transferredProps,
	} = props;

	let scopes = "read:org,user:email";
	if (scope === "private") {
		scopes += ",repo";
	}

	return <AuthButton
		authInfo={{
			scopes,
			pageName,
			returnTo,
			newUserReturnTo,
			provider: "github",
			eventObject: Events.OAuth2FlowGitHub_Initiated,
		}}
		iconType="github"
		color={color}
		{...transferredProps}>
		{children}
	</AuthButton>;
}
