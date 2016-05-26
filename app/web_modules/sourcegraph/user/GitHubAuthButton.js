// @flow

import React from "react";
import {urlToGitHubOAuth} from "sourcegraph/util/urlTo";
import {GitHubIcon} from "sourcegraph/components/Icons";
import CSSModules from "react-css-modules";
import {Button} from "sourcegraph/components";
import style from "sourcegraph/user/styles/accountForm.css";

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
	};

	render() {
		return (
			<a href={this.props.url ? this.props.url : urlToGitHubOAuth} styleName="github"
				onClick={() => this.context.eventLogger.logEvent("InitiateGitHubOAuth2Flow")} {...this.props}>
				<Button type="button" outline={this.props.outline} formNoValidate={true} color={this.props.color} block={this.props.block}>
					<GitHubIcon />&nbsp; {this.props.children}
				</Button>
			</a>
		);
	}
}


export default CSSModules(GitHubAuthButton, style);
