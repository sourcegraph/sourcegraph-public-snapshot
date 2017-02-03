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
	let authForm: HTMLFormElement | null = null;

	const {
		provider,
		iconType,
		eventObject,
		scopes,
		returnTo,
		newUserReturnTo,
		secondaryText,
		pageName = "",
		children,
		...btnProps
	} = props;

	const url = urlToOAuth(provider, scopes || null, returnTo || null, newUserReturnTo || returnTo || null);
	const iconSx = props.size === "small" ? typography.size[5] : typography.size[4];

	const icon = <span style={{ marginRight: whitespace[2] }}>
		{iconType === "github" && <GitHubLogo style={iconSx} />}
		{iconType === "google" && <Google style={iconSx} />}
	</span>;

	const submitAuthForm = () => {
		if (authForm) {
			authForm.submit();
		}
	};

	return <form method="POST" ref={el => authForm = el} action={url} onSubmit={() => eventObject.logEvent({ page_name: pageName })}>
		<input type="hidden" name="gorilla.csrf.Token" value={context.csrfToken} />
		{secondaryText
			? <SplitButton onClick={submitAuthForm} {...btnProps} secondaryText={secondaryText}>{icon} {children}</SplitButton>
			: <Button onClick={submitAuthForm} {...btnProps}>{icon} {children}</Button>
		}
	</form>;
}
