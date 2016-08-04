// tslint:disable

import * as React from "react";
import {urlToGitHubOAuth} from "sourcegraph/util/urlTo";
import {GitHubIcon} from "./Icons";
import typography from "./styles/_typography.css";
import {Button} from "sourcegraph/components/index";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

type Props = {
	scopes?: string,
	returnTo?: string | Location | null,
	color?: string,
	outline?: boolean,
	block?: boolean,
	size?: string,
	children?: any,
	className?: string,
	onClick?: () => void,
	tabIndex?: string,
};

class GitHubAuthButton extends React.Component<any, any> {
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

export default GitHubAuthButton;
