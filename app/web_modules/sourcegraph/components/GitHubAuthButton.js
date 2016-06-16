// @flow

import React from "react";
import {urlToGitHubOAuth} from "sourcegraph/util/urlTo";
import {GitHubIcon} from "./Icons";
import type from "./styles/_typography.css";
import {Button} from "sourcegraph/components";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

class GitHubAuthButton extends React.Component {
	static propTypes = {
		url: React.PropTypes.string,
		color: React.PropTypes.string,
		outline: React.PropTypes.bool,
		block: React.PropTypes.bool,
		children: React.PropTypes.oneOfType([
			React.PropTypes.arrayOf(React.PropTypes.element),
			React.PropTypes.element,
			React.PropTypes.string,
		]),
	};
	static contextTypes = {
		eventLogger: React.PropTypes.object.isRequired,
	};
	static defaultProps = {
		color: "blue",
		outline: false,
		block: false,
		url: urlToGitHubOAuth,
	};

	render() {
		const {url, outline, color, block, children} = this.props;
		return (
			<a href={url}
				onClick={() => this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "InitiateGitHubOAuth2Flow")} {...this.props}>
				<Button type="button" outline={outline} formNoValidate={true} color={color} block={block}>
					<GitHubIcon className={type.f4} />&nbsp; {children}
				</Button>
			</a>
		);
	}
}

export default GitHubAuthButton;
