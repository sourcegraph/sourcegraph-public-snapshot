import * as omit from "lodash/omit";
import * as React from "react";

import { RouterLocation } from "sourcegraph/app/router";
import { AuthButton } from "sourcegraph/components/AuthButton";
import { ButtonProps } from "sourcegraph/components/Button";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

interface Props extends ButtonProps {
	scopes?: string;
	returnTo?: string | RouterLocation;
	newUserReturnTo?: string | RouterLocation;
	pageName?: string;
	secondaryText?: string;
}

export function GitHubAuthButton(props: Props): JSX.Element {
	const transferredProps = omit(props, ["scopes", "provider", "iconType", "eventObject"]);
	const scopes = props.scopes || "read:org,repo,user:email";

	return <AuthButton
		provider="github"
		iconType="github"
		eventObject={AnalyticsConstants.Events.OAuth2FlowGitHub_Initiated}
		scopes={scopes}
		{...transferredProps}>
		{props.children}
	</AuthButton>;
}
