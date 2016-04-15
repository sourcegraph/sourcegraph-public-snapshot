// @flow weak

import React from "react";
import {Button} from "sourcegraph/components";
import {urlToGitHubOAuth} from "sourcegraph/util/urlTo";
import EventLogger from "sourcegraph/util/EventLogger";
import context from "sourcegraph/app/context";
import {GitHubIcon} from "sourcegraph/components/Icons";

import CSSModules from "react-css-modules";
import styles from "sourcegraph/dashboard/styles/Dashboard.css";

class LinkGitHubCTA extends React.Component {
	render() {
		if (context.hasLinkedGitHub) return null;

		return (
			<div styleName="cta">
				<a href={!context.currentUser ? "/join" : urlToGitHubOAuth} onClick={() => context.currentUser && EventLogger.logEvent("SubmitLinkGitHub")}>
					<Button size="normal" outline={true} color="warning">{context.currentUser && <GitHubIcon styleName="github-icon" />}
						{context.currentUser ? "Add my GitHub repositories" : "Add Sourcegraph to my code"}
					</Button>
				</a>
			</div>
		);
	}
}

export default CSSModules(LinkGitHubCTA, styles);
