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

export function GoogleAuthButton(props: Props): JSX.Element {
	const transferredProps = omit(props, ["provider", "iconType", "eventObject"]);
	return <AuthButton
		provider="google"
		iconType="google"
		eventObject={AnalyticsConstants.Events.OAuth2FlowGCP_Initiated}
		{...transferredProps}>
		{props.children}
	</AuthButton>;
}
