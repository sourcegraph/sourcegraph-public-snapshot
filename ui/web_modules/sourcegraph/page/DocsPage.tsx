import * as React from "react";
import Helmet from "react-helmet";
import {context} from "sourcegraph/app/context";
import {Heading, Hero} from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";
import * as styles from "sourcegraph/page/Page.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";


export function DocsPage(): JSX.Element {
	return (
		<div>
			<Helmet title="Docs" />
			<Hero pattern="objects" className={base.pv5}>
				<div className={styles.container}>
					<Heading level="2" color="blue">Sourcegraph docs</Heading>
				</div>
			</Hero>
			<div className={styles.content}>
				<Heading level="3">Overview</Heading>
				<p className={styles.p}>Sourcegraph is a tool that helps developers explore and understand code. These docs describe how to leverage Sourcegraph in your development workflow. If you have any problems or requests, please contact <a href="#contact_us" onClick={(e) => e && EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_DOCS, AnalyticsConstants.ACTION_CLICK, "clickedContactSupportFromDocs")}>support</a>.</p>
				<ul>
					<li>
						<a href="#sourcegraph">Sourcegraph</a>
						<ul>
							<li>
								<a href="#code_intelligence">Code Intelligence</a>
								<ul>
									<li><a href="#find_local_references">Find Local References</a></li>
									<li><a href="#peek_definition">Peek Definition</a></li>
									<li><a href="#go_to_definition">Go to Definition</a></li>
									<li><a href="#find_external_references">Find External References</a></li>
								</ul>
							</li>
							<li>
								<a href="#search">Search</a>
								<ul>
									<li><a href="#repository_search">Repository Search</a></li>
									<li><a href="#file_search">File Search</a></li>
									<li><a href="#definition_search">Definition Search</a></li>
								</ul>
							</li>
						</ul>
					</li>
					<li>
						<a href="#github_extension">GitHub Extension</a>
						<ul>
							<li><a href="#github_hover_over_documentation">Hover over Documentation</a></li>
							<li><a href="#github_jump_to_definition">Jump to Definition</a></li>
						</ul>
					</li>
					<li>
						<a href="#auth_private_repos">Sourcegraph for your Private Code</a>
					</li>
					<li>
						<a href="#languages_supported">Languages Supported</a>
					</li>
				</ul>
				<a id="sourcegraph"></a>
				<Heading level="3" underline="blue">Sourcegraph</Heading>
				<a id="code_intelligence"></a>
				<Heading level="4" className={styles.h5}>Code Intelligence</Heading>
				<p className={styles.p}>Sourcegraph’s code intelligence provides you with full context of the code you are reading, without leaving the page.</p>
				<br/>

				<a id="find_local_references"></a>
				<Heading level="5" className={styles.h5}>Find Local References</Heading>
				<p className={styles.p}>Right click on a symbol and select “Find Local References” to find all the places it's referenced in the current repository.</p>
				<img src={`${context.assetsRoot}/img/DocsPage/find_local_references_1.png`} width="100%" />
				<p className={styles.p}>A pop-up will appear, under the symbol, showing how it's being used in specific examples.</p>
				<img src={`${context.assetsRoot}/img/DocsPage/find_local_references_2.png`} width="100%" />
				<br/>
				<br/>
				<br/>

				<a id="peek_definition"></a>
				<Heading level="5" className={styles.h5}>Peek Definition</Heading>
				<p className={styles.p}>Right click on a symbol and select “Peek Definition” to view how it's defined without leaving the page.</p>
				<img src={`${context.assetsRoot}/img/DocsPage/peek_definition_1.png`} width="100%" />
				<img src={`${context.assetsRoot}/img/DocsPage/peek_definition_2.png`} width="100%" />
				<br/>
				<br/>
				<br/>

				<a id="go_to_definition"></a>
				<Heading level="5" className={styles.h5}>Go to Definition</Heading>
				<p className={styles.p}>Click on a symbol to jump to its definition. Alternatively, you can right click on a symbol and select “Go to Definition” to do the same thing.</p>
				<img src={`${context.assetsRoot}/img/DocsPage/go_to_definition.png`} width="100%" />
				<br/>
				<br/>
				<br/>

				<a id="find_external_references"></a>
				<Heading level="5" className={styles.h5}>Find External References</Heading>
				<p className={styles.p}>Right click on a symbol and select “Find External References” to find all the places it's referenced across all publicly viewable code on GitHub.</p>
				<img src={`${context.assetsRoot}/img/DocsPage/find_external_references_1.png`} width="100%" />
				<img src={`${context.assetsRoot}/img/DocsPage/find_external_references_2.png`} width="100%" />
				<br/>
				<br/>

				<br/>

				<a id="search"></a>
				<Heading level="4" className={styles.h5}>Search</Heading>
				<p className={styles.p}>Sourcegraph allows you to quickly jump between code definitions, files, and repositories through our snappy search interface. Bring up the search bar by:</p>
				<ul>
					<li className={styles.p}>hitting "/" from anywhere on sourcegraph.com, or</li>
					<li className={styles.p}>clicking the search icon in the nav bar</li>
				</ul>
				<br/>

				<a id="repository_search"></a>
				<Heading level="5" className={styles.h5}>Repository Search</Heading>
				<p className={styles.p}>Jump to any publicly viewable GitHub repository and also any private repositories you've authenticated.</p>
				<img src={`${context.assetsRoot}/img/DocsPage/repo_search.png`} width="100%" />
				<br/>
				<br/>

				<a id="file_search"></a>
				<Heading level="5" className={styles.h5}>File Search</Heading>
				<p className={styles.p}>Once you are within a repository, you can jump to any file within the repository.</p>
				<img src={`${context.assetsRoot}/img/DocsPage/file_search.png`} width="100%" />
				<br/>
				<br/>

				<a id="definition_search"></a>
				<Heading level="5" className={styles.h5}>Definition Search</Heading>
				<p className={styles.p}>Once you are within a repository, you can jump to any definition. A definition can be any function, method, struct, type, variable, or package.</p>
				<img src={`${context.assetsRoot}/img/DocsPage/definition_search.png`} width="100%" />
				<br/>
				<br/>

				<br/>

				<a id="github_extension"></a>
				<Heading level="3" underline="blue">GitHub Extension</Heading>
				<p className={styles.p}>Sourcegraph's GitHub extension allows you to browse GitHub with IDE-like functionality. Click <a href="https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack" onClick={(e) => e && EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_DOCS, AnalyticsConstants.ACTION_CLICK, "clickedInstallBrowserExtFromDocs")}>here</a> to install.</p>

				<a id="github_hover_over_documentation"></a>
				<Heading level="5" className={styles.h5}>Hover over Documentation</Heading>
				<p className={styles.p}>Hover over any symbol on GitHub to get its type information and documentation.</p>
				<img src={`${context.assetsRoot}/img/DocsPage/github_hover_over_documentation.png`} width="100%" />
				<br/>
				<br/>

				<a id="github_jump_to_definition"></a>
				<Heading level="5" className={styles.h5}>Jump to Definition</Heading>
				<p className={styles.p}>Click on a symbol on GitHub to jump to its definition.</p>
				<img src={`${context.assetsRoot}/img/DocsPage/github_jump_to_definition.png`} width="100%" />
				<br/>
				<br/>

				<br/>

				<a id="auth_private_repos"></a>
				<Heading level="3" underline="blue">Sourcegraph for your Private Code</Heading>
				<p className={styles.p}>Want code intelligence and search for your private repositories? Click <a href="https://sourcegraph.com/?ob=github" onClick={(e) => e && EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_DOCS, AnalyticsConstants.ACTION_CLICK, "clickedAuthPrivateReposFromDocs")}>here</a> to authenticate.</p>

				<br/>

				<a id="languages_supported"></a>
				<Heading level="3" underline="blue">Languages Supported</Heading>
				<p className={styles.p}>Sourcegraph currently supports:</p>
				<ul>
					<li className={styles.p}>Go</li>
				</ul>
				<p className={styles.p}>Coming soon:</p>
				<ul>
					<li className={styles.p}>Java</li>
					<li className={styles.p}>Ruby</li>
					<li className={styles.p}>JavaScript</li>
					<li className={styles.p}>TypeScript</li>
				</ul>
				<br/>

				<p id="contact_us" className={styles.p}>Find a bug or want to request a feature? Email support@sourcegraph!</p>

			</div>
		</div>
	);
}
