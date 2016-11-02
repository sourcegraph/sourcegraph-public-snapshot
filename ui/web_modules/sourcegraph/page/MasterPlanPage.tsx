// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Link} from "react-router";
import {Button, Input, Hero, Heading, Panel} from "sourcegraph/components";
import * as styles from "sourcegraph/page/Page.css";
import * as base from "sourcegraph/components/styles/_base.css";
import Helmet from "react-helmet";
import {whitespace, layout} from "sourcegraph/components/utils";
import {LocationStateToggleLink} from "sourcegraph/components/LocationStateToggleLink";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {media} from "glamor";
import {context} from "sourcegraph/app/context";

function tldr() {
	return (
		<ol>
			<li>Make basic code intelligence (jump-to-def, find-references, etc.) ubiquitous</li>
			<li>Make code review continuous and more intelligent</li>
			<li>Make this all work globally, not just within a single project, to increase the amount/quality of available open-source code and help you avoid reinventing the wheel</li>
		</ol>
	);
}

// emailSubscribeForm returns the email subscription form. tabIndex
// should be an odd number that is unique among all the calls to
// emailSubscribeForm in this file.
function emailSubscribeForm(tabIndex: number, block: boolean) {
	return (
		<div>
			<div className={styles.mailchimp_embed}>
				<form action="//sourcegraph.us8.list-manage.com/subscribe/post?u=81d5d2fe17e49697663f46503&amp;id=32642fc470" method="post" id="mc-embedded-subscribe-form" name="mc-embedded-subscribe-form" className="validate" target="_blank" noValidate>
					<div>
						<div style={{position: "absolute", left: "-5000px"}} aria-hidden="true"><input type="text" name="b_81d5d2fe17e49697663f46503_32642fc470" tabIndex={-1} value="" /></div>
						<Input block={block} type="email" defaultValue="" name="EMAIL" id="mce-EMAIL" placeholder="Email address" style={{marginBottom: whitespace[1]}} className={styles.mailchimp_input} tabIndex={tabIndex} />
						<Button type="submit" block={block} id="mc-embedded-subscribe" color="blue" className={styles.mailchimp_button} tabIndex={tabIndex + 1}>Subscribe to updates</Button>
					</div>
				</form>
			</div>
		</div>
	);
}

function signInButton(block: boolean) {
	return (
		<LocationStateToggleLink
			href="/join"
			modalName="join"
			location={location}
			onToggle={(v) => v && AnalyticsConstants.Events.JoinModal_Initiated.logEvent({page_name: location.pathname})}
			{...media(layout.breakpoints["sm"], { display: "none"})}
			style={{
				paddingTop: whitespace[2],
				paddingBottom: whitespace[2],
			}}
		><Button block={block} color="purple" style={{marginTop: whitespace[1]}}>Start using Sourcegraph</Button>
		</LocationStateToggleLink>
	);
}

export function MasterPlanPage(props: {}) {
	return (
		<div className={styles.mp_body}>
			<Helmet title="Sourcegraph Master Plan" />
			<Hero pattern="objects" color="blue" className={base.pv4}>
				<div className={styles.container_wide}>
					<Heading level={2} color="white">Sourcegraph Master Plan</Heading>
					<p className={styles.p}>What we're building and why it matters</p>
				</div>
			</Hero>
			<div className={styles.content_wide}>
				<Panel hoverLevel="high" hover={false} className={styles.sidebox} style={{
					margin: whitespace[4],
					padding: whitespace[3],
				}}>
					<Heading level={5} className="{styles.h6}">
						Sourcegraph Master Plan, in short
					</Heading>
					<div>
						{tldr()}
						{emailSubscribeForm(5, true)}
						{!context.user && <div><hr className={styles.button_sep} />{signInButton(true)}</div>}
					</div>
				</Panel>

				<p>Today, Sourcegraph gives you the power of an IDE (jump-to-def, search, and find-references) when reading code on the web, either on <a href="https://sourcegraph.com">Sourcegraph.com</a>, or on GitHub with the <a href="https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack">Sourcegraph Chrome extension</a>. What most people don’t know is that our long-term vision is to make it so everyone, in every community, in every country, and in every industry—not just the ones working at the half-dozen dominant tech companies—can create products using the best technology. We believe this is the only way the world will sustain broad economic growth and build the innovations we need over the next 100 years in transportation, health care, energy, AI, communication, space travel, etc.</p>
				<p>In 1976, just 0.2% of the world’s population had access to a computer. Apple’s vision then was to create a “bicycle for the mind” in the form of a computer, and Microsoft put a computer “on every desk and in every home.” Together, these companies succeeded in bringing computing to billions of people. But these billions of people are still using software applications built by just 0.2% of the world’s population (those who can code).</p>
				<p>The next step is to make it so billions of people, not just 0.2% of the world population, can build software (not just use it). Amazon Web Services and others solve the distribution piece: a tiny team can reach millions of users using the same infrastructure available to the most advanced tech companies. But the process of creating software is stuck in the mainframe era: the “developer experience” of building software is still poor, duplicative, manual, and single-player—and every software project is about integrating components of variable quality developed mostly in isolation, with a high chance of failure.</p>
				<p>At Sourcegraph, we want to fix this and eventually enable everyone to build software. For now, we’re revealing our master plan for phase 1: how we’re going to make it easier and faster for <em>today’s developers</em> to build software.</p>
				<p>In short, phase 1 is:</p>
				{tldr()}
				<p>When phase 1 is almost done, we’ll reveal phase 2: how we’ll work toward enabling everyone to code. If you think that’s crazy, ask yourself: now that billions of people have access to the Internet, is coding more like reading and writing (which virtually everyone does) or publishing books (which 0.1% of the population does)?</p>

				<br/><br/>
				<Heading level={4} underline="purple" className={styles.h5}><a id="code-intelligence" style={{color: "inherit"}}>Make basic code intelligence ubiquitous (in every editor and language)</a></Heading>
				<Panel hoverLevel="low" hover={false} className={styles.sidebox} style={{
					margin: whitespace[4],
					padding: whitespace[3],
				}}>
					<Heading level={5} className="{styles.h6}">
						Try code intelligence on Sourcegraph
					</Heading>
					<div>
						<p><Link to="/">Sourcegraph</Link> currently supports 4 languages on the web and <a href="https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack" target="_blank">Chrome extension</a>, with many more planned by the end of 2016.</p>
						<ul>
							<li>Go &mdash; <Link target="_blank" to="/github.com/gorilla/websocket/-/blob/client.go">try it</Link></li>
							<li>TypeScript (beta) &mdash; <Link target="_blank" to="/github.com/Microsoft/vscode/-/blob/src/vs/editor/contrib/hover/browser/hoverOperation.ts#L51-52">try it</Link></li>
							<li>JavaScript (beta) &mdash; <Link target="_blank" to="/github.com/swimlane/angular-data-table@master/-/blob/src/components/DataTableController.js#L33">try it</Link></li>
							<li>C (beta) &mdash; <Link target="_blank" to="/github.com/jgamblin/Mirai-Source-Code/-/blob/mirai/bot/resolv.c">try it</Link></li>
						</ul>
						<p><Link to="/beta">Sign up for early access</Link> to other languages.</p>
					</div>
				</Panel>

				<p>Every developer deserves to have all these features work 100% of the time:</p>
				<ul>
					<li>Jump to definition</li>
					<li>Hover tooltips (with type info and docs)</li>
					<li>Semantic code search</li>
					<li>Find references (local and cross-repository)</li>
					<li>Autocomplete</li>
					<li>Automatic formatting</li>
					<li>Inline errors, linting, and diagnostics</li>
				</ul>
				<p>The above features should be expected:</p>
				<ul>
					<li>For every language (especially dynamic languages, such as JavaScript, Python, and Ruby)</li>
					<li>In every IDE and every text editor (so you can use your editor of choice)</li>
					<li>Everywhere you read code (on GitHub, Bitbucket, GitLab, Visual Studio Team Services, Stack Overflow, Phabricator, Sourcegraph, etc.)</li>
					<li>Everywhere you review code (in pull requests on GitHub, etc.)</li>
					<li>For both your own code and the code of all of your dependencies</li>
					<li>With zero configuration</li>
				</ul>
				<p>Getting this basic code intelligence everywhere is an obvious win. Unfortunately, it’s far too difficult to install and configure it today, so most developers are missing these benefits for a large portion of their work.</p>
				<p>The current approach is broken because it’s an “<strong><em>m</em></strong>-times-<strong><em>n</em></strong>” solution: one tool for each combination of <strong><em>m</em></strong> applications (Vim, Emacs, Visual Studio, Sublime, IntelliJ, Eclipse, GitHub’s code file viewer, Codenvy, etc.) and <strong><em>n</em></strong> languages (JavaScript, C++, Java, C#, Python, etc.). This means we’d need thousands of individual tools, all maintained independently, to get complete coverage.</p>
				<p>Here’s how to fix it and bring basic code intelligence to every developer, everywhere:</p>
				<ol>
					<li>Transform the “<strong><em>m</em></strong>-times-<strong><em>n</em></strong>” language-editor tool problem into a more manageable “<strong><em>m</em></strong>-plus-<strong><em>n</em></strong>” problem by using the <a href="https://github.com/Microsoft/language-server-protocol">Language Server Protocol (LSP)</a> open standard
						<ul>
							<li>Create open-source LSP language servers for every language &mdash; <strong><a href="http://langserver.org/" target="_blank">in progress</a></strong></li>
							<li>Create open-source LSP adapter plugins for every editor, code viewer, and code review tool &mdash; <strong><a href="http://langserver.org/" target="_blank">in progress</a></strong></li>
							<li>Provide the infrastructure for language server developers to measure coverage and accuracy over a large dataset of open-source code &mdash; <strong>in progress</strong></li>
						</ul>
					</li>
					<li>Make it easy for projects to supply the necessary configuration (if any) so that everyone gets code intelligence on the project’s code</li>
					<li>Make it quick and easy to add/install code intelligence for any language in your tools of choice</li>
				</ol>
				<p>The end result is that anytime you look at code, you have the full power of a perfectly configured IDE.</p>

				<br/>
				<Panel hoverLevel="high" hover={false} style={{
					margin: "auto",
					maxWidth: 960,
					padding: whitespace[4],
				}}>
						<Heading level={5} className="{styles.h6}">
							Take <a href="https://text.sourcegraph.com/the-sourcegraph-test-e5c281850c" target="_blank">The Sourcegraph Test</a> to see how good your team's dev tools are
						</Heading>
				</Panel>

				<br/><br/><br/>
				<Heading level={4} underline="purple" className={styles.h5}><a id="code-review" style={{color: "inherit"}}>Make code review continuous and more intelligent</a></Heading>
				<p>Code review is supposed to improve quality and share knowledge. But few teams feel their code review process (if any) is effective, because today’s tools make code review a manual, error-prone process performed (far too often) at the very end of the development cycle.</p>
				<p>Toyota long ago showed that high-quality production processes should be the opposite: continuous (to find defects immediately, not after the car is fully assembled) and systematic (based on checklists compiled from experience). Medicine and aviation also recognize the value of this approach. We’ll apply these principles to make code review continuous and more intelligent, so you can:</p>
				<ul>
					<li>See realtime impact analysis of any changes, in the form of a checklist of possible impacts/defects drawn from:
						<ul>
							<li>Code intelligence (call graph/dependencies)</li>
							<li>Repository data (merge conflicts, blame/authorship)</li>
							<li>Heuristics from past code reviews</li>
						</ul>
					</li>
					<li>Likewise, see your teammates’ work-in-progress changes that affect your current work</li>
					<li>Quickly share code with teammates to get help or informal reviews instead of waiting until the end</li>
					<li>Automatically and always have code reviewed by the right teammates</li>
					<li>Easily run the modified codebase to inspect the actual product (not just the code)</li>
				</ul>
				<p>Current code review tools aren’t able to provide these things because they lack code intelligence and a way to give realtime feedback on your local work-in-progress changes. The previous step (bringing basic code intelligence to everyone in all the tools they use) fixes this: it provides the underlying analysis to automatically enumerate possible impacts/defects—and the UI (in their editor and other existing tools) to collect and present this information as needed.</p>
				<p>Here’s how we’ll bring continuous, intelligent code review (as described above) to every team:</p>
				<ol>
					<li>Add basic code intelligence (jump-to-def, hover, find-references, etc.) to diff views in code review tools (GitHub pull requests, etc.) &mdash; <strong><a href="https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack">&#x2714; done</a></strong></li>
					<li>Apply code intelligence to provide an impact analysis checklist for every change in every code review tool</li>
					<li>Create a way to enable quick sharing of code in your working tree</li>
					<li>Make this all realtime, automatically updated as you make changes in your editor</li>
				</ol>

				<Panel hoverLevel="high" hover={false} style={{
					textAlign: "center",
					padding: whitespace[4],
				}}>
					<Heading level={5} className="{styles.h6}">
						Join the Sourcegraph code review beta
					</Heading>
					<div>
						<a href="https://docs.google.com/a/sourcegraph.com/forms/d/e/1FAIpQLSf9BYHsv2P1hWt4oRY0_PkteXsb5bxQq3ZXjEJZYf2y6vajFg/viewform" target="_blank"><Button color="blue">Apply for beta access</Button></a>
					</div>
				</Panel>

				<br/><br/><br/>
				<Heading level={4} underline="purple" className={styles.h5}><a id="global-code-graph" style={{color: "inherit"}}>Build the global code graph</a></Heading>
				<p>The fundamental problem of software development is that most developers spend most of their time doing things that aren’t core to solving their actual problem. Of all the code you write, only a tiny fraction is core to your particular business or application. Likewise for the bugs you spend time fixing.</p>

				<p>We will make it much easier to create and reuse public, open-source code by giving everyone access to the global code graph. The global code graph is the collection of all the code in the world, stored in a system that understands the dependency and call graph relationships across tens of millions of codebases. It's what powers the features in the previous steps.</p>

				<p>This will increase the amount and reusability of available code by 10-100x because the current tools for creating and using open-source code are very limited. For one: creators and maintainers of open-source projects have no data about who’s using their project and how, except from bug reports. Imagine running and stocking a supermarket if you only knew what items customers returned, not what they bought.</p>

				<p>The global code graph will make it easier and more rewarding to create and maintain open-source code:</p>
				<ul>
					<li>Users can opt in to share aggregate data about how they’re using open-source projects (what APIs, patterns, etc., they use), determined automatically from the users’ own code. Every project’s community will grow because every user becomes a contributor.</li>
					<li>Projects can use this information to prioritize enhancements, bug fixes, documentation, etc.</li>
					<li>Creators and maintainers can see and understand how they are helping hundreds or thousands of people all around the world (instead of just seeing bug reports and stars).</li>
				</ul>

				<p>The global code graph will also make it easier for you to find and reuse high-quality open-source code:</p>
				<ul>
					<li>When coding, you can see contextual usage examples/patterns and discussions based on everyone else’s similar (opted-in) code—on the web or in your editor.</li>
					<li>If you make a common mistake that other users have encountered and flagged, you’ll be notified immediately. When a library releases an update, you’ll be notified about the impact it has on your own code, and you can see information from everyone else who has upgraded.</li>
					<li>When evaluating a library, you can see how many people are actively using it, what APIs they’re calling, what other libraries they’re using alongside it, etc., to make the best decision about which library to use.</li>
				</ul>

				<p>We’ll build open-source tools and open APIs to make these data and features accessible to every developer, in every environment, in every workflow. Code hosts, monitoring tools, cloud providers, etc., will also be able to enhance their own products by using and adding to the global graph graph.</p>

				<p>The global code graph is inevitable and universally beneficial, and there are many important things to get right:</p>
				<ul>
					<li>How do we get developers to opt in to contribute to the graph, not just consume it (i.e., avoid the tragedy of the commons)? Developers, ourselves included, are very privacy conscious. But developers also love to help each other and advance technology—no other profession shares as much and as openly as developers. Open source itself is evidence of this.</li>
					<li>For developers who opt in to sharing with the graph, how do we extract useful signals from code without leaking sensitive information?</li>
					<li>And many more things. We released this master plan early so we can start the community discussion and do it right, together.</li>
				</ul>

				<p>Getting these right and building the global code graph means you’ll be able to find and use more existing, high-quality open-source components for the common parts of your application, so you can focus on solving the problems that are unique to your business or project.</p>

				<br/><br/>

				<Panel hoverLevel="high" hover={false} style={{
					margin: "auto",
					maxWidth: 960,
					padding: whitespace[4],
				}}>
					<Heading level={5} className="{styles.h6}">
						Sourcegraph Master Plan, in short
					</Heading>
					<div>
						{tldr()}
						{!context.user && signInButton(false)}
						<br/><br/>
						<Heading level={5} className="{styles.h6}">Follow our progress</Heading>
						{emailSubscribeForm(7, false)}
					</div>
				</Panel>
			</div>
		</div>
	);
}
