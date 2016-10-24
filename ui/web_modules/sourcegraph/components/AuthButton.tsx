import * as React from "react";
import {context} from "sourcegraph/app/context";
import {Button} from "sourcegraph/components";
import {GitHubIcon, GoogleIcon} from "sourcegraph/components/Icons";
import {typography, whitespace} from "sourcegraph/components/utils";
import {Location} from "sourcegraph/Location";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";
import {oauthProvider, urlToOAuth} from "sourcegraph/util/urlTo";

interface Props {
	provider: oauthProvider;
	iconType: "github" | "google";
	eventLabel: "InitiateGitHubOAuth2Flow" | "InitiateGoogleOAuth2Flow";

	scopes?: string;
	returnTo?: string | Location;

	color?: string;
	outline?: boolean;
	block?: boolean;
	size?: string;
	className?: string;
	tabIndex?: number;
	pageName?: string;
	img?: boolean;
	style?: React.CSSProperties;
	children?: React.ReactNode[];
}

export function AuthButton(props: Props): JSX.Element {
	const {
		provider,
		iconType,
		eventLabel,

		scopes,
		returnTo,

		color = "blue",
		outline,
		block,
		size,
		className,
		tabIndex,
		pageName,
		img = true,
		style,
		children,
	} = props;

	const url = urlToOAuth(provider, scopes || null, returnTo || null);

	const iconSx = size === "small" ? typography.size[5] : typography.size[4];

	return (
		<form method="POST" action={url}>
			<input type="hidden" name="gorilla.csrf.Token" value={context.csrfToken} />
			<Button
				style={style}
				type="submit"
				outline={outline}
				formNoValidate={true}
				color={color}
				block={block}
				size={size}
				className={className}
				tabIndex={tabIndex}
				onClick={() => {
					EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, eventLabel, {page_name: pageName});
				}}>
				{img &&
					<span style={{marginRight: whitespace[2]}}>
						{iconType === "github" && <GitHubIcon style={iconSx}/>}
						{iconType === "google" && <GoogleIcon style={iconSx}/>}
					</span>
				}
				{children}
			</Button>
		</form>
	);
}
