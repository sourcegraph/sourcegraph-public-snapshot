import * as React from "react";
import { context } from "sourcegraph/app/context";
import { RouterLocation } from "sourcegraph/app/router";
import { Button, SplitButton } from "sourcegraph/components";
import { ButtonProps } from "sourcegraph/components/Button";
import { GitHubLogo, Google } from "sourcegraph/components/symbols";
import { typography, whitespace } from "sourcegraph/components/utils";
import { LoggableEvent } from "sourcegraph/tracking/constants/AnalyticsConstants";
import { oauthProvider, urlToOAuth } from "sourcegraph/util/urlTo";

interface Props extends ButtonProps {
	iconType: "github" | "google";
	secondaryText?: string;
	authInfo: ActionProps;
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

	const { submit, form } = GetAuthzAction(authInfo);

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

interface ActionProps {
	eventObject: LoggableEvent;
	pageName?: string;

	provider: oauthProvider;
	scopes: string;
	returnTo?: string | RouterLocation;
	newUserReturnTo?: string | RouterLocation;
}

/**
 * Get an authorization action and form. The invisible form must be included in
 * the page for the action to work.
 */
export function GetAuthzAction(props: ActionProps): { submit: () => void, form: JSX.Element } {
	let url = urlToOAuth(
		props.provider,
		props.scopes,
		props.returnTo || null,
		props.newUserReturnTo || null,
	);

	let authForm: HTMLFormElement | null = null;
	const submitAuthForm = () => {
		if (authForm) {
			authForm.submit();
		}
	};
	const logEvent = () => {
		props.eventObject.logEvent({ page_name: props.pageName || "" });
	};

	return {
		submit: submitAuthForm,
		form: <form
			action={url}
			method="POST"
			onSubmit={logEvent}
			ref={el => authForm = el}
			style={{ display: "none" }} >
			<input type="hidden" name="gorilla.csrf.Token" value={context.csrfToken} />
		</form>
	};
}
