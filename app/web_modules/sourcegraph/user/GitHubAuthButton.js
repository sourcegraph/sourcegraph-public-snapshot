// @flow

import React from "react";
import {urlToGitHubOAuth} from "sourcegraph/util/urlTo";
import {GitHubIcon} from "sourcegraph/components/Icons";
import CSSModules from "react-css-modules";
import {Button} from "sourcegraph/components";
import style from "sourcegraph/user/styles/accountForm.css";

function GitHubAuthButton(props, {eventLogger}) {
	return (
		<a href={props.url ? props.url : urlToGitHubOAuth} styleName="github"
			onClick={() => eventLogger.logEvent("InitiateGitHubOAuth2Flow")} {...props}>
			<Button type="button" outline={Boolean(props.outline)} formNoValidate={true} color={props.color || "blue"} block={true}>
				<GitHubIcon />&nbsp; {props.children}
			</Button>
		</a>
	);
}
GitHubAuthButton.propTypes = {
	url: React.PropTypes.string,
	color: React.PropTypes.string,
	outline: React.PropTypes.bool,
	children: React.PropTypes.oneOfType([
		React.PropTypes.arrayOf(React.PropTypes.element),
		React.PropTypes.element,
		React.PropTypes.string,
	]),
};
GitHubAuthButton.contextTypes = {
	eventLogger: React.PropTypes.object.isRequired,
};

export default CSSModules(GitHubAuthButton, style);
