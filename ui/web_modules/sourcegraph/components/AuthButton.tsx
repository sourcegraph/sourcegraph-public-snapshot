import * as omit from "lodash/omit";
import * as React from "react";
import { context } from "sourcegraph/app/context";
import { RouterLocation } from "sourcegraph/app/router";
import { Button, SplitButton } from "sourcegraph/components";
import { ButtonProps } from "sourcegraph/components/Button";
import { GitHubLogo, Google } from "sourcegraph/components/symbols";
import { typography, whitespace } from "sourcegraph/components/utils";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { oauthProvider, urlToOAuth } from "sourcegraph/util/urlTo";

export interface Props extends ButtonProps {
	provider: oauthProvider;
	iconType: "github" | "google";
	eventObject: AnalyticsConstants.LoggableEvent;

	scopes?: string;
	returnTo?: string | RouterLocation;
	newUserReturnTo?: string | RouterLocation;
	pageName?: string;

	secondaryText?: string;
}

export function AuthButton(props: Props): JSX.Element {
	const {
		provider,
		iconType,
		eventObject,
		scopes,
		returnTo,
		newUserReturnTo,
		secondaryText,
		size,
		pageName = "",
		children,
	} = props;

	const url = urlToOAuth(provider, scopes || null, returnTo || null, newUserReturnTo || returnTo || null);
	const iconSx = size === "small" ? typography.size[5] : typography.size[4];

	const btnProps = omit(props, [
		"provider",
		"iconType",
		"eventObject",
		"scopes",
		"returnTo",
		"newUserReturnTo",
		"secondaryText",
		"pageName",
	]);

	const icon = <span style={{ marginRight: whitespace[2] }}>
		{iconType === "github" && <GitHubLogo style={iconSx} />}
		{iconType === "google" && <Google style={iconSx} />}
	</span>;

	return (
		<form method="POST" action={url} onSubmit={() => eventObject.logEvent({ page_name: pageName })}>
			<input type="hidden" name="gorilla.csrf.Token" value={context.csrfToken} />
			{secondaryText ? <SplitButton type="submit" {...btnProps} secondaryText={secondaryText}>
				{icon} {children}
			</SplitButton> : <Button type="submit" {...btnProps}>	{icon} {children}</Button>}
		</form>
	);
}
