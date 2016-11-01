import * as React from "react";
import Helmet from "react-helmet";
import {Heading, Hero} from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";
import * as styles from "sourcegraph/page/Page.css";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

export function DocsPage(): JSX.Element {
	return (
		<div>
			<Helmet title="Docs" />
			<Hero pattern="objects" color="blue" className={base.pv5}>
				<div className={styles.container}>
					<Heading level={2} color="white">Sourcegraph Docs</Heading>
				</div>
			</Hero>
			<div className={styles.content}>
				<Heading level={3}>Overview</Heading>
				<p className={styles.p}>Sourcegraph is a tool that helps developers explore and understand code. These docs describe how to use Sourcegraph in your development workflow. If you have any problems or requests, please contact <a href="#contact_us" onClick={(e) => e && AnalyticsConstants.Events.DocsContactSupportCTA_Clicked.logEvent()}>support</a>.</p>
				<ul>
					<li>
						<a href="#code_intelligence">Code Intelligence</a>
						<ul>
							<li><a href="#find_local_references">Find Local References</a></li>
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
					<li>
						<a href="#chrome_extension">Chrome Extension</a>
						<ul>
							<li><a href="#github_hover_over_documentation">Hover over Documentation</a></li>
							<li><a href="#github_go_to_definition">Go to Definition</a></li>
							<li><a href="#github_keyboard_shortcuts">Keyboard Shortcuts</a></li>
						</ul>
					</li>
					<li>
						<a href="#languages_supported">Languages Supported</a>
					</li>
				</ul>
				<a id="sourcegraph"></a>
				<a id="code_intelligence"></a>
				<Heading level={4} className={styles.h5}>Code Intelligence</Heading>
				<p className={styles.p}>Sourcegraph’s code intelligence provides you with full context of the code you are reading, without leaving the page.</p>
				<br/>

				<a id="find_local_references"></a>
				<Heading level={5} className={styles.h5}>Find Local References</Heading>
				<p className={styles.p}>Right click on a symbol and select “Find Local References” to find all the places it's referenced in the current repository.</p>
				<a href="https://sourcegraph.com/github.com/gorilla/mux/-/blob/mux.go" target="_blank"><img src="https://storage.googleapis.com/sourcegraph-assets/find_local_references_1.png" width="100%"/></a>
				<p className={styles.p}>A pop-up will appear, under the symbol, showing how it's being used in specific examples.</p>
				<a href="https://sourcegraph.com/github.com/gorilla/mux/-/blob/mux.go" target="_blank"><img src="https://storage.googleapis.com/sourcegraph-assets/find_local_references_2.png" width="100%"/></a>
				<br/>
				<br/>
				<br/>

				<a id="go_to_definition"></a>
				<Heading level={5} className={styles.h5}>Go to Definition</Heading>
				<p className={styles.p}>Click on a symbol to jump to its definition. Alternatively, you can right click on a symbol and select “Go to Definition” to do the same thing.</p>
				<a href="https://sourcegraph.com/github.com/gorilla/mux/-/blob/mux.go" target="_blank"><img src="https://storage.googleapis.com/sourcegraph-assets/go_to_definition.png" width="100%"/></a>
				<br/>
				<br/>
				<br/>

				<a id="find_external_references"></a>
				<Heading level={5} className={styles.h5}>Find External References</Heading>
				<p className={styles.p}>Right click on a symbol and select “Find External References” to find all the places it's referenced across all publicly viewable code on GitHub.</p>
				<a href="https://sourcegraph.com/github.com/gorilla/mux/-/blob/mux.go" target="_blank"><img src="https://storage.googleapis.com/sourcegraph-assets/find_external_references.png" width="100%"/></a>
				<br/>
				<br/>

				<br/>

				<a id="search"></a>
				<Heading level={4} className={styles.h5}>Search</Heading>
				<p className={styles.p}>Sourcegraph allows you to quickly jump between code definitions, files, and repositories through our snappy search interface. Bring up the search bar by:</p>
				<ul>
					<li className={styles.p}>hitting "/" from anywhere on sourcegraph.com, or</li>
					<li className={styles.p}>clicking the search icon in the nav bar</li>
				</ul>
				<br/>

				<a id="repository_search"></a>
				<Heading level={5} className={styles.h5}>Repository Search</Heading>
				<p className={styles.p}>Jump to any publicly viewable GitHub repository and also any private repositories you've authenticated.</p>
				<a href="https://sourcegraph.com/github.com/docker/docker/-/blob/api/errors/errors.go" target="_blank"><img src="https://storage.googleapis.com/sourcegraph-assets/repo_search.png" width="100%" /></a>
				<br/>
				<br/>

				<a id="file_search"></a>
				<Heading level={5} className={styles.h5}>File Search</Heading>
				<p className={styles.p}>Once you are within a repository, you can jump to any file within the repository.</p>
				<a href="https://sourcegraph.com/github.com/gorilla/mux" target="_blank"><img src="https://storage.googleapis.com/sourcegraph-assets/file_search.png" width="100%" /></a>
				<br/>
				<br/>

				<a id="definition_search"></a>
				<Heading level={5} className={styles.h5}>Definition Search</Heading>
				<p className={styles.p}>Once you are within a repository, you can jump to any definition. A definition can be any function, method, struct, type, variable, or package.</p>
				<a href="https://sourcegraph.com/github.com/gorilla/mux/-/blob/mux.go" target="_blank"><img src="https://storage.googleapis.com/sourcegraph-assets/definition_search.png" width="100%" /></a>
				<br/>
				<br/>

				<br/>

				<a id="chrome_extension"></a>
				<Heading level={3} underline="blue">Chrome Extension</Heading>
				<p className={styles.p}>Sourcegraph's Chrome extension allows you to browse GitHub with IDE-like functionality. <a href="https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack" target="_blank" onClick={(e) => e && AnalyticsConstants.Events.DocsInstallExtensionCTA_Clicked.logEvent()}>Install our Chrome extension.</a></p>

				<a id="github_hover_over_documentation"></a>
				<Heading level={5} className={styles.h5}>Hover over Documentation</Heading>
				<p className={styles.p}>Hover over any symbol on GitHub to get its type information and documentation.</p>
				<a href="https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack" target="_blank"><img src="https://storage.googleapis.com/sourcegraph-assets/github_hover_over_documentation.png" width="100%" /></a>
				<br/>
				<br/>

				<a id="github_go_to_definition"></a>
				<Heading level={5} className={styles.h5}>Go to Definition</Heading>
				<p className={styles.p}>Click on a symbol on GitHub to go to its definition.</p>
				<a href="https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack" target="_blank"><img src="https://storage.googleapis.com/sourcegraph-assets/github_jump_to_definition.png" width="100%" /></a>
				<br/>
				<br/>

				<a id="github_keyboard_shortcuts"></a>
				<Heading level={5} className={styles.h5}>Keyboard shortcuts</Heading>
				<p className={styles.p}>Press `u` when viewing code on GitHub to navigate to the same code on Sourcegraph.com.</p>

				<br/>

				<a id="languages_supported"></a>
				<Heading level={3} underline="blue">Languages Supported</Heading>
				<p className={styles.p}>Sourcegraph currently supports:</p>
				<ul>
					<li className={styles.p}>Go</li>
				</ul>
				<p className={styles.p}>In Beta:</p>
				<ul>
					<li className={styles.p}>TypeScript</li>
					<li className={styles.p}>JavaScript</li>
					<li className={styles.p}>C</li>
					<li className={styles.p}><a href="https://sourcegraph.com/beta" target="_blank">Don't see your language?</a></li>
				</ul>
				<br/>

				<p id="contact_us" className={styles.p}>Want to request a feature or find a bug? Email support@sourcegraph.com.</p>

			</div>
		</div>
	);
}
