// @flow

import React from "react";
import {urlToGitHubOAuth} from "sourcegraph/util/urlTo";
import {GitHubIcon} from "./Icons";
import type from "./styles/_typography.css";
import {Button} from "sourcegraph/components";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

type Props = {
	scopes: string,
	returnTo?: ?(string | Location),
	color: string,
	outline: bool,
	block: bool,
	size?: string,
	children?: any,
};

class GitHubAuthButton extends React.Component {
	static propTypes = {
		scopes: React.PropTypes.string, // OAuth2 scopes

		// return-to location after GitHub OAuth is complete
		returnTo: React.PropTypes.oneOfType([React.PropTypes.object, React.PropTypes.string]),

		color: React.PropTypes.string,
		outline: React.PropTypes.bool,
		block: React.PropTypes.bool,
		children: React.PropTypes.oneOfType([
			React.PropTypes.arrayOf(React.PropTypes.element),
			React.PropTypes.element,
			React.PropTypes.string,
		]),
		size: React.PropTypes.string,
	};
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
	props: Props;

	render() {
		const {scopes, returnTo, outline, color, block, children, size} = this.props;
		const url = urlToGitHubOAuth(scopes, returnTo);
		return (
			<a href={url}
				onClick={() => this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "InitiateGitHubOAuth2Flow")} {...this.props}>
				<Button type="button" outline={outline} formNoValidate={true} color={color} block={block} size={size}>
					<GitHubIcon className={size === "small" ? type.f5 : type.f4} />&nbsp; {children}
				</Button>
			</a>
		);
	}
}

export default GitHubAuthButton;
