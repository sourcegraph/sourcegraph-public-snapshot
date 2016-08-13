// tslint:disable: typedef ordered-imports

import * as React from "react";
import {urlToGitHubOAuth} from "sourcegraph/util/urlTo";
import {GitHubIcon} from "./Icons";
import * as typography from "./styles/_typography.css";
import {Button} from "sourcegraph/components/index";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

interface Props {
	scopes?: string;
	returnTo?: string | Location | null;
	color?: string;
	outline?: boolean;
	block?: boolean;
	size?: string;
	children?: any;
	className?: string;
	onClick?: () => void;
	tabIndex?: string;
}

type State = any;

export class GitHubAuthButton extends React.Component<Props, State> {
	static contextTypes = {
		eventLogger: React.PropTypes.object.isRequired,
	};
	static defaultProps: Props = {
		scopes: "",
		returnTo: null,
		color: "blue",
		outline: false,
		block: false,
	};

	render(): JSX.Element | null {
		const {scopes, returnTo, outline, color, block, children, size} = this.props;
		const url = urlToGitHubOAuth(scopes || null, returnTo || null);
		return (
			<a href={url}
				onClick={() => (this.context as any).eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "InitiateGitHubOAuth2Flow")} {...this.props as any}>
				<Button type="button" outline={outline} formNoValidate={true} color={color} block={block} size={size}>
					<GitHubIcon className={size === "small" ? typography.f5 : typography.f4} />&nbsp; {children}
				</Button>
			</a>
		);
	}
}
