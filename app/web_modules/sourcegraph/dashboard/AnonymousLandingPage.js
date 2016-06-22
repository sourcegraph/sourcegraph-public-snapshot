import React from "react";
import {Link} from "react-router";
import Component from "sourcegraph/Component";
import CSSModules from "react-css-modules";
import {Logo, Button, Heading, Panel} from "sourcegraph/components";
import styles from "./styles/Home.css";
import base from "sourcegraph/components/styles/_base.css";
import GitHubAuthButton from "sourcegraph/components/GitHubAuthButton";
import GlobalSearch from "sourcegraph/search/GlobalSearch";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

class AnonymousLandingPage extends Component {
	static propTypes = {
		location: React.PropTypes.object.isRequired,
	}

	static contextTypes = {
		signedIn: React.PropTypes.bool.isRequired,
		siteConfig: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
	}

	reconcileState(state, props) {
		state.location = props.location;
	}

	render() {
		const {siteConfig, eventLogger} = this.context;
		return (
			<div styleName="flex-fill" style={{marginTop: "-2.3rem"}}>
				<div styleName="box-purple-gradient" className={base.pt5}>
					<div styleName="search-container" className={base.pt3}>
						<div styleName="row tc">
							<Heading level="1" color="white">
								Global code search &amp; cross&#8209;references
							</Heading>
						</div>
						<div styleName="row">
							<Panel hoverLevel="low" className={`${base.mv4} ${base.pb4} ${base.ph4} ${base.pt3}`}>
								<GlobalSearch query={this.props.location.query.q || ""} location={this.props.location} />
							</Panel>
						</div>
						<div styleName="row tc search-examples">
							<div styleName="examples-label">
								<span styleName="white">Try some common searches </span>
								<span styleName="examples-brace">{"{"}</span>
							</div>
							<div className={base.pl2}>
								<table>
									<tbody>
										<tr>
											<td>Go:</td>
											<td styleName="examples-category">
												<Link to="/?q=golang+http.Get">
													<code styleName="search-example">http.Get</code>
												</Link>
												<Link to="/?q=golang+Sprintf">
													<code styleName="search-example">Sprintf</code>
												</Link>
												<Link to="/?q=func+Decode">
													<code styleName="search-example">func Decode</code>
												</Link>
											</td>
										</tr>
										<tr>
											<td>Java:</td>
											<td styleName="examples-category">
												<Link to="/?q=java+sql.ResultSet">
													<code styleName="search-example">sql.ResultSet</code>
												</Link>
												<Link to="/?q=java+DateTime">
													<code styleName="search-example">DateTime</code>
												</Link>
												<Link to="/?q=java+junit+assertEquals">
													<code styleName="search-example">junit assertEquals</code>
												</Link>
											</td>
										</tr>
									</tbody>
								</table>
							</div>
						</div>
						<div styleName="row tc">
							<p styleName="white ma0-sm" className={`${base.mb4}`}>
								<GitHubAuthButton outline={true} color="purple" className={base.mr3}>
									<strong>Sign up with GitHub</strong>
								</GitHubAuthButton>
							</p>
						</div>
					</div>
				</div>
				<div styleName="container-lg">
					<div styleName="content-block">
						<div styleName="img-left">
							<Link to="/github.com/golang/go@0cc710dca63b79ed2dd6ce9375502e76e5fc0484/-/tree/src/testing?q=testing" onClick={(v) => v && eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_DEF_INFO, AnalyticsConstants.ACTION_CLICK, "ClickedExplorePublicRepo", {page_name: AnalyticsConstants.PAGE_HOME})}>
								<img src={`${siteConfig.assetsRoot}/img/Homepage/screenshot-sourcegraph.png`} styleName="img" width="460" />
							</Link>
						</div>
						<div styleName="content-right">
							<div styleName="content">
								<Logo width="32px" className={base.mt4} />
								<h3 styleName="h3">Search by function, type, or package – globally</h3>
								<p>Connect your GitHub account to Sourcegraph to start searching, browsing, and cross-referencing your code, with IDE-like capabilities in your browser. Free for public and private projects.</p>
							</div>
							<GitHubAuthButton className={base.mr3}><strong>Continue with GitHub</strong></GitHubAuthButton>
							<Link to="/github.com/aws/aws-sdk-go/-/info/GoPackage/github.com/aws/aws-sdk-go/aws/credentials/-/NewStaticCredentials" onClick={(v) => v && eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_DEF_INFO, AnalyticsConstants.ACTION_CLICK, "ClickedExplorePublicRepo", {page_name: AnalyticsConstants.PAGE_HOME})} styleName="block-sm mv4-sm">Or try it on open-source code &nbsp;&#x276f;</Link>
						</div>
					</div>

					<div styleName="content-block">
						<div styleName="img-right">
							<a href="https://chrome.google.com/webstore/detail/sourcegraph-chrome-extens/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en" target="new">
								<img src={`${siteConfig.assetsRoot}/img/Homepage/screenshot-github.png`} styleName="img" width="460" />
							</a>
						</div>
						<div styleName="content-left">
							<div styleName="content">
								<img src={`${siteConfig.assetsRoot}/img/symbols/branch.svg`} className={base.mt3} />
								<h3 styleName="h3">Chrome extension for GitHub</h3>
								<p>Browse GitHub like an IDE, with jump-to-definition links, semantic code search, and documentation tooltips. <em>Support for other browsers is coming soon.</em></p>
							</div>
							<a href="https://chrome.google.com/webstore/detail/sourcegraph-chrome-extens/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en" target="new">
								<Button color="blue" onClick={(v) => v && eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_DEF_INFO, AnalyticsConstants.ACTION_CLICK, "ClickedInstallChromeExt", {page_name: AnalyticsConstants.PAGE_HOME})}>
									Install the Chrome extension
								</Button>
							</a>
						</div>
					</div>

					<div styleName="content-block">
						<div styleName="img-left">
							<Link to="/tools/editor" onClick={(v) => v && eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_TOOLS, AnalyticsConstants.ACTION_CLICK, "ClickedInstallEditorExt", {page_name: AnalyticsConstants.PAGE_HOME})}>
								<img src={`${siteConfig.assetsRoot}/img/Homepage/screenshot-editor.png`} styleName="img" width="460" />
							</Link>
						</div>
						<div styleName="content-right">
							<div styleName="content">
								<Logo width="32px" className={base.mt4} />
								<h3 styleName="h3">Sourcegraph for your editor</h3>
								<p>See usage examples for code instantly, as you type. It's like pair programming with the smartest developer in the world.</p>
								<div styleName="flex">
									<div styleName="sfye-flex">
										<a styleName="" href="/tools/editor">
											<Button color="blue" onClick={(v) => v && eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_TOOLS, AnalyticsConstants.ACTION_CLICK, "ClickedInstallEditorExt", {page_name: AnalyticsConstants.PAGE_HOME})}>
												Install for your editor
											</Button>
										</a>
									</div>
									<div styleName="sfye_beta"><em>Beta for Go in Sublime Text and Vim. <a href="/tools/editor?expanded=true">Get notified when Sourcegraph is available for your editor.</a></em></div>
								</div>
							</div>
						</div>
					</div>

					<div styleName="box-white">
						<div styleName="language-container">
							<Heading level="1" underline="blue" align="center">Language support</Heading>
							<p styleName="lead tc">Powered by <a href="https://srclib.org/" target="new">srclib</a>, a hackable code analysis library.</p>

							<div styleName="language" className={base.mt5}>
								Go
								<span styleName="label-blue">75,311 projects</span>
							</div>
							<div styleName="row" className={base.mt4}>
								<div styleName="featured-project">
									<Link to="/github.com/golang/go">
										<img src={`${siteConfig.assetsRoot}/img/symbols/folder.svg`} className={`${base.mt1} ${base.mr2}`} width="16px" />
										<strong>golang/go</strong>
									</Link>
									<p styleName="project-desc">
										Used in 45,328 repositories
									</p>
								</div>
								<div styleName="featured-project">
									<Link to="/github.com/gorilla/mux">
										<img src={`${siteConfig.assetsRoot}/img/symbols/folder.svg`} className={`${base.mt1} ${base.mr2}`} width="16px" />
										<strong>gorilla/mux</strong>
									</Link>
									<p styleName="project-desc">
										Used in 1,843 repositories
									</p>
								</div>
								<div styleName="featured-project">
									<Link to="/github.com/aws/aws-sdk-go">
										<img src={`${siteConfig.assetsRoot}/img/symbols/folder.svg`} className={`${base.mt1} ${base.mr2}`} width="16px" />
										<strong>aws-sdk-go</strong>
									</Link>
									<p styleName="project-desc">
										Used in 171 repositories
									</p>
								</div>
							</div>
							<div styleName="language" className={base.mt5}>
								Java
								<span styleName="label-blue">15,782 projects</span>
							</div>
							<div styleName="row" className={base.mt4}>
								<div styleName="featured-project">
									<Link to="/github.com/google/guava">
										<img src={`${siteConfig.assetsRoot}/img/symbols/folder.svg`} className={`${base.mt1} ${base.mr2}`} width="16px" />
										<strong>guava</strong>
									</Link>
									<p styleName="project-desc">
										Used in 6,256 repositories
									</p>
								</div>
								<div styleName="featured-project">
									<Link to="/github.com/JodaOrg/joda-time">
										<img src={`${siteConfig.assetsRoot}/img/symbols/folder.svg`} className={`${base.mt1} ${base.mr2}`} width="16px" />
										<strong>joda-time</strong>
									</Link>
									<p styleName="project-desc">
										Used in 1,113 repositories
									</p>
								</div>
								<div styleName="featured-project">
									<Link to="/github.com/dropwizard/metrics">
										<img src={`${siteConfig.assetsRoot}/img/symbols/folder.svg`} className={`${base.mt1} ${base.mr2}`} width="16px" />
										<strong>dropwizard/metrics</strong>
									</Link>
									<p styleName="project-desc">
										Used in 87 repositories
									</p>
								</div>
							</div>

							<h5 styleName="header-5" className={base.mt6}>Coming soon</h5>

							<div styleName="row">
								<div styleName="language-2">
									C#
								</div>
								<div styleName="language-5">
									JavaScript
								</div>
								<div styleName="language-2">
									Python
								</div>
							</div>
						</div>
					</div>

				</div>

				<div styleName="box-purple-gradient" className={`${base.pt6} ${base.pb5}`}>
					<div styleName="bottom-container">
						<Heading level="3" color="white" className={base.mb3}>
							We built Sourcegraph to keep you in flow while coding
						</Heading>
						<p styleName="lead white">
							Start saving time and sharpening your skills. Join tons of other developers who use Sourcegraph, around the world and in large, well-known companies.
						</p>
						<p className={base.mt4}>
							<GitHubAuthButton color="purple" outline={true} className={base.mr3}>
								<strong>Sign up with GitHub</strong>
							</GitHubAuthButton>
							<a target="_blank" styleName="block-sm mv4-sm white"
								href="https://chrome.google.com/webstore/detail/sourcegraph-chrome-extens/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en"
								onClick={(v) => v && eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_TOOLS, AnalyticsConstants.ACTION_CLICK, "ClickedInstallChromeExt", {page_name: AnalyticsConstants.PAGE_HOME})}>
								Just install the Chrome extension &nbsp;&#x276f;
							</a>
						</p>
					</div>
				</div>

			</div>
		);
	}
}

export default CSSModules(AnonymousLandingPage, styles, {allowMultiple: true});
