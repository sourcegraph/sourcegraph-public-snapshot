import * as React from "react";

import { RouterLocation } from "sourcegraph/app/router";
import { AuthButton } from "sourcegraph/components/AuthButton";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

interface Props {
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
	style?: React.CSSProperties;
	children?: React.ReactNode[];
}

export function GitHubAuthButton(props: Props): JSX.Element {
	const scopes = props.scopes || "read:org,repo,user:email";

	return (
		<AuthButton
			provider="github"
			iconType="github"
			eventObject={AnalyticsConstants.Events.OAuth2FlowGitHub_Initiated}
			scopes={scopes}
			returnTo={props.returnTo}
			newUserReturnTo={props.newUserReturnTo}
			color={props.color}
			outline={props.outline}
			block={props.block}
			size={props.size}
			className={props.className}
			tabIndex={props.tabIndex}
			pageName={props.pageName}
			img={props.img}
			style={props.style}>
			{props.children}
		</AuthButton>
	);
}
