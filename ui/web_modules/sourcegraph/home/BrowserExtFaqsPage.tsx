// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Link} from "react-router";
import {privateGitHubOAuthScopes} from "sourcegraph/util/urlTo";
import {Hero, Heading} from "sourcegraph/components";
import * as styles from "sourcegraph/page/Page.css";
import * as base from "sourcegraph/components/styles/_base.css";
import {GitHubAuthButton} from "sourcegraph/components/GitHubAuthButton";
import Helmet from "react-helmet";
import {context} from "sourcegraph/app/context";

interface Props {
	location: any;
}

type State = any;

export class BrowserExtFaqsPage extends React.Component<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		eventLogger: React.PropTypes.object.isRequired,
	};

	_hasPrivateGitHubToken() {
		return context.gitHubToken && context.gitHubToken.scope && context.gitHubToken.scope.includes("repo") && context.gitHubToken.scope.includes("read:org");
	}

	render(): JSX.Element | null {
		let hash = this.props.location.hash;
		return (
			<div>
				<Helmet title="Browser Extension FAQs" />
				<Hero pattern="objects" className={base.pv5}>
					<div className={styles.container}>
						<Heading level="2" color="blue">Sourcegraph Browser Extension FAQs</Heading>
					</div>
				</Hero>

				<div className={styles.faqs_content}>
					<div className={styles.faqs_div}>
						<Heading level="3" underline="blue">How does it work?</Heading>
						<p className={styles.p} max-width="640px">Sourcegraph is how developers discover and understand code. It is a fast, global, semantic code search and cross-reference engine. You can search for any function, type, or package, and see how other developers use it, globally. Sourcegraph is massively scalable, with over 2,000,000,000+ Go and Java nodes in the public code index (and growing).  Sourcegraph's code index powers the Sourcegraph browser extension.</p>
					</div>

					<div className={styles.faqs_div}>
						<Heading level="3" underline="blue">Can I use the extension anonymously?</Heading>
						<div className={styles.p}>Yep - Sourcegraph indexes popular public repositories written in supported languages, and you'll be able to use jump to def on those repositories. Don't see a public repository you'd like indexed? Sign in at <Link to="https://sourcegraph.com">sourcegraph.com</Link>, and the next time you visit that <a id="signin"/>repository on GitHub, we'll start indexing it for you.</div>
					</div>

					<div className={hash === "#signin" ? styles.faqs_div_selected : styles.faqs_div}>
						<Heading level="3" underline="blue">How come I can't see my private code on my private repositories?</Heading>
						<div className={styles.p}> You must be signed in on Sourcegraph.com and link your GitHub account to access your indexed private repositories. Sign in or
						sign up below with GitHub OAuth.</div>
						<a id="enable"/>
						<div className={styles.centered_button}>
							{!context.user && <GitHubAuthButton returnTo="/about/browser-ext-faqs#signin"> Sign up or sign in</GitHubAuthButton>}
						</div>
					</div>

					<div className={hash === "#enable" ? styles.faqs_div_selected : styles.faqs_div}>
						<Heading level="3" underline="blue">How do I use the browser extension with private repositories?</Heading>
						<p className={styles.p}> Sourcegraph indexes repositories that you have enabled. <Link to="https://sourcegraph.com/settings/repos">
						Check your repository settings</Link> if you have not enabled below.</p>
						<div className={styles.centered_button}>
							{!this._hasPrivateGitHubToken() && <GitHubAuthButton scopes={privateGitHubOAuthScopes} returnTo="/about/browser-ext-faqs#enable"> Enable private repositories</GitHubAuthButton>}
						</div>
					</div>

					<div className={styles.faqs_div}>
						<Heading level="3" underline="blue">Where else is Sourcegraph?</Heading>
						<p className={styles.p}> Like the Sourcegraph browser extension? Try global semantic search at<Link to="https://sourcegraph.com"> sourcegraph.com</Link>.</p>
					</div>

					<a id="build"/>
					<div className={hash === "#build" ? styles.faqs_div_selected : styles.faqs_div}>
						<Heading level="3" underline="blue">Troubleshooting</Heading>
						<p className={styles.p}> Sourcegraph runs static analysis on your code. The most common reasons your code might not be analyzed are 1) it's in the build queue, and may take some more time; 2) the file is an unsupported language; 3) the project doesn't build, or we don't support the build system in question. If you think you are seeing a bug, please report it at<a href="mailto:support@sourcegraph.com"> support@sourcegraph.com</a>.</p>
					</div>
				</div>
				<Hero pattern="objects" className={base.pv5}/>
			</div>
		);
	}
}
