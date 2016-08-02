// @flow

import React from "react";
import {Link} from "react-router";
import {privateGitHubOAuthScopes} from "sourcegraph/util/urlTo";
import {Hero, Heading} from "sourcegraph/components";
import styles from "sourcegraph/page/Page.css";
import base from "sourcegraph/components/styles/_base.css";
import CSSModules from "react-css-modules";
import GitHubAuthButton from "sourcegraph/components/GitHubAuthButton";
import Helmet from "react-helmet";

class BrowserExtFaqsPage extends React.Component {
	static propTypes = {
		location: React.PropTypes.object.isRequired,
	};

	static contextTypes = {
		signedIn: React.PropTypes.bool.isRequired,
		githubToken: React.PropTypes.object,
		eventLogger: React.PropTypes.object.isRequired,
	};

	_hasPrivateGitHubToken() {
		return this.context.githubToken && this.context.githubToken.scope && this.context.githubToken.scope.includes("repo") && this.context.githubToken.scope.includes("read:org");
	}

	render() {
		let hash = this.props.location.hash;
		return (
			<div>
				<Helmet title="Browser Extension FAQs" />
				<Hero pattern="objects" className={base.pv5}>
					<div styleName="container">
						<Heading level="2" color="blue">Sourcegraph Browser Extension FAQs</Heading>
					</div>
				</Hero>

				<div styleName="faqs_content">
					<div styleName="faqs_div">
						<Heading level="3" underline="blue">How does it work?</Heading>
						<p styleName="p" max-width="640px">Sourcegraph is how developers discover and understand code. It is a fast, global, semantic code search and cross-reference engine. You can search for any function, type, or package, and see how other developers use it, globally. Sourcegraph is massively scalable, with over 2,000,000,000+ Go and Java nodes in the public code index (and growing).  Sourcegraph's code index powers the Sourcegraph browser extension.</p>
					</div>

					<div styleName="faqs_div">
						<Heading level="3" underline="blue">Can I use the extension anonymously?</Heading>
						<div styleName="p">Yep - Sourcegraph indexes popular public repositories written in supported languages, and you'll be able to use jump to def on those repositories. Don't see a public repository you'd like indexed? Sign in at <Link to="https://sourcegraph.com">sourcegraph.com</Link>, and the next time you visit that <a id="signin"/>repository on GitHub, we'll start indexing it for you.</div>
					</div>

					<div styleName={hash === "#signin" ? "faqs_div_selected" : "faqs_div"}>
						<Heading level="3" underline="blue">How come I can't see my private code on my private repositories?</Heading>
						<div styleName="p"> You must be signed in on Sourcegraph.com and link your GitHub account to access your indexed private repositories. Sign in or
						sign up below with GitHub OAuth.</div>
						<a id="enable"/>
						<div styleName="centered_button">
							{!this.context.signedIn && <GitHubAuthButton returnTo="/about/browser-ext-faqs#signin"> Sign up or sign in</GitHubAuthButton>}
						</div>
					</div>

					<div styleName={hash === "#enable" ? "faqs_div_selected" : "faqs_div"}>
						<Heading level="3" underline="blue">How do I use the browser extension with private repositories?</Heading>
						<p styleName="p"> Sourcegraph indexes repositories that you have enabled. <Link to="https://sourcegraph.com/settings/repos">
						Check your repository settings</Link> if you have not enabled below.</p>
						<div styleName="centered_button">
							{!this._hasPrivateGitHubToken() && <GitHubAuthButton scopes={privateGitHubOAuthScopes} returnTo="/about/browser-ext-faqs#enable"> Enable private repositories</GitHubAuthButton>}
						</div>
					</div>

					<div styleName="faqs_div">
						<Heading level="3" underline="blue">Where else is Sourcegraph?</Heading>
						<p styleName="p"> Like the Sourcegraph browser extension? Try global semantic search at<Link to="https://sourcegraph.com"> sourcegraph.com</Link>.</p>
					</div>

					<a id="build"/>
					<div styleName={hash === "#build" ? "faqs_div_selected" : "faqs_div"}>
						<Heading level="3" underline="blue">Troubleshooting</Heading>
						<p styleName="p"> Sourcegraph runs static analysis on your code. The most common reasons your code might not be analyzed are 1) it's in the build queue, and may take some more time; 2) the file is an unsupported language; 3) the project doesn't build, or we don't support the build system in question. If you think you are seeing a bug, please report it at<a href="mailto:support@sourcegraph.com"> support@sourcegraph.com</a>.</p>
					</div>
				</div>
				<Hero pattern="objects" className={base.pv5}/>
			</div>
		);
	}
}

export default CSSModules(BrowserExtFaqsPage, styles);
