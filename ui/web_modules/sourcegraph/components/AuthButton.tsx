import * as React from "react";
import { context } from "sourcegraph/app/context";
import { RouterLocation } from "sourcegraph/app/router";
import { Button } from "sourcegraph/components";
import { GitHubLogo, Google } from "sourcegraph/components/symbols";
import { typography, whitespace } from "sourcegraph/components/utils";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { oauthProvider, urlToOAuth } from "sourcegraph/util/urlTo";

interface Props {
	provider: oauthProvider;
	iconType: "github" | "google";
	eventObject: AnalyticsConstants.LoggableEvent;

	scopes?: string;
	returnTo?: string | RouterLocation;
	newUserReturnTo?: string | RouterLocation;
	color?: "blue" | "green" | "orange" | "purple" | "red" | "blueGray";
	outline?: boolean;
	block?: boolean;
	size?: "small" | "large";
	className?: string;
	tabIndex?: number;
	pageName?: string;
	img?: boolean;
	tintLabel?: string;
	style?: React.CSSProperties;
	children?: React.ReactNode[];
}

export function AuthButton(props: Props): JSX.Element {
	const {
		provider,
		iconType,
		eventObject,

		scopes,
		returnTo,
		newUserReturnTo,
		color = "blue",
		outline,
		block,
		size,
		className,
		tabIndex,
		pageName = "",
		img = true,
		tintLabel,
		style,
		children,
	} = props;

	const url = urlToOAuth(provider, scopes || null, returnTo || null, newUserReturnTo || null);
	const iconSx = size === "small" ? typography.size[5] : typography.size[4];

	return (
		<form method="POST" action={url} onSubmit={() => eventObject.logEvent({ page_name: pageName })}>
			<input type="hidden" name="gorilla.csrf.Token" value={context.csrfToken} />
			<Button
				style={style}
				type="submit"
				outline={outline}
				formNoValidate={true}
				color={color}
				block={block}
				size={size}
				tintLabel={tintLabel}
				className={className}
				tabIndex={tabIndex}>
				{img &&
					<span style={{ marginRight: whitespace[2] }}>
						{iconType === "github" && <GitHubLogo style={iconSx} />}
						{iconType === "google" && <Google style={iconSx} />}
					</span>
				}
				{children}
			</Button>
		</form>
	);
}
