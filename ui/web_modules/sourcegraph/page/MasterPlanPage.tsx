// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Link} from "react-router";
import {Hero, Heading, Panel, Collapsible} from "sourcegraph/components";
import * as styles from "sourcegraph/page/Page.css";
import * as base from "sourcegraph/components/styles/_base.css";
import Helmet from "react-helmet";
import {whitespace} from "sourcegraph/components/utils";

export function MasterPlanPage(props: {}) {
	return (
		<div>
			<Helmet title="Sourcegraph Master Plan" />
			<Hero pattern="objects" color="blue" className={base.pv4}>
				<div className={styles.container}>
					<Heading level={2} color="white">Sourcegraph Master Plan</Heading>
					<p className={styles.p}>What we're building and why it matters</p>
				</div>
			</Hero>
			<div className={styles.content}>
				<p>Today, Sourcegraph gives you the power of an IDE (jump-to-def, search, and find-references) when reading code on the web, either on <a href="https://sourcegraph.com">Sourcegraph.com</a>, or on GitHub with the <a href="https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack">Sourcegraph Chrome extension</a>. What most people don’t know is that our long-term vision is to make it so everyone, in every community, in every country, and in every industry—not just the ones working at the half-dozen dominant tech companies—can create products using the best technology. We believe this is the only way the world will sustain broad economic growth and build the innovations we need over the next 100 years in transportation, health care, energy, AI, communication, space travel, etc.</p>
				<p>In 1976, just 0.2% of the world’s population had access to a computer. Apple’s vision then was to create a “bicycle for the mind” in the form of a computer, and Microsoft put a computer “on every desk and in every home.” Together, these companies succeeded in bringing computing to billions of people. But these billions of people are still using software applications built by just 0.2% of the world’s population (those who can code).</p>
				<p>The next step is to make it so billions of people, not just 0.2% of the world population, can build software (not just use it). Amazon Web Services and others solve the distribution piece: a tiny team can reach millions of users using the same infrastructure available to the most advanced tech companies. But the process of creating software is stuck in the mainframe era: the “developer experience” of building software is still poor, duplicative, manual, and single-player—and every software project is about integrating components of variable quality developed mostly in isolation, with a high chance of failure.</p>
				<p>At Sourcegraph, we want to fix this and eventually enable everyone to build software. For now, we’re revealing our master plan for phase 1: how we’re going to make it easier and faster for <em>today’s developers</em> to build software.</p>
				<p>The <strong>tl;dr</strong> for phase 1 is:</p>
				<ol>
					<li>Make basic code intelligence (jump-to-def, find-references, etc.) ubiquitous</li>
					<li>Make code review continuous and more intelligent</li>
					<li>Make this all work globally, not just within a single project, so we can increase the amount/quality of available open-source code and help you avoid reinventing the wheel</li>
				</ol>
				<p>When phase 1 is almost done, we’ll reveal phase 2: how we’ll work toward enabling everyone to code. If you think that’s crazy, ask yourself: now that billions of people have access to the Internet, is coding more like reading and writing (which virtually everyone does) or publishing books (which 0.1% of the population does)?</p>

				<br/>
				<Heading level={4} underline="purple" className={styles.h5}>Make basic code intelligence ubiquitous (in every editor and language)</Heading>
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
					<li>Transform the “<strong><em>m</em></strong>-times-<strong><em>n</em></strong>” language-editor tool problem into a more manageable “<strong><em>m</em></strong>-plus-<strong><em>n</em></strong>” problem by using the <a href="https://github.com/Microsoft/language-server-protocol">Language Server Protocol (LSP)</a> open standard</li>
					<li>Create LSP language servers for every language</li>
					<li>Provide the infrastructure for language server developers to measure coverage and accuracy over a large dataset of open-source code</li>
					<li>Create LSP adapter plugins for every editor, code viewer, and code review tool</li>
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
					<Collapsible collapsed={true}>
						<Heading level={5} className="{styles.h6}">
							Does <i>your</i> environment pass The Quinn Test?
							<span style={{color: "gray",float: "right"}}>&#9660;</span>
						</Heading>
						<div>
							<p>Like the <Link to="http://www.joelonsoftware.com/articles/fog0000000043.html">The Joel Test</Link> that helped set the standard for a quality software team, The Quinn Test is a 10-question survey to answer: does my team's environment give us the code intelligence to compete with the giants?</p>
							<ol>
								<li>Do you have jump-to-def in your primary language and editor?</li>
								<li>Does your jump-to-def work across repository boundaries?</li>
								<li>Do you have find-references in your primary language and editor?</li>
								<li>Do you have inline error messages and diagnostics in your primary language and editor?</li>
								<li>Do you have jump-to-def in your primary editor for all of the languages you use?</li>
								<li>Does everyone else on your team satisfy the above 5 questions at least as well as you do?</li>
								<li>Do you have jump-to-def and find-references in your code review tool?</li>
								<li>Do you have jump-to-def and find-references in your code host?</li>
								<li>Is there an automatic notification when a dependency of your project has an important update?</li>
								<li>Do you receive and perform code reviews?</li>
								<li>Does your code review process use any form of checklist (i.e., it’s not completely up to the discretion of the reviewer)?</li>
							</ol>
							<p><a href="https://twitter.com/srcgraph" target="_blank">Let us know</a> how your team did on the Quinn test!</p>
						</div>
					</Collapsible>
				</Panel>


				<br/>
				<Heading level={4} underline="purple" className={styles.h5}>Make code review continuous and more intelligent</Heading>
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
					<li>Add basic code intelligence (jump-to-def, hover, find-references, etc.) to diff views in code review tools (GitHub pull requests, etc.) &nbsp;<a href="https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack">&#x2714; DONE</a></li>
					<li>Apply code intelligence to provide an impact analysis checklist for every change in every code review tool</li>
					<li>Create a way to enable quick sharing of code in your working tree</li>
					<li>Make this all realtime, automatically updated as you make changes in your editor</li>
				</ol>

				<br/>
				<Heading level={4} underline="purple" className={styles.h5}>Build the global code graph</Heading>
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

				<br/><hr/><br/>

				<p>So, in short, the master plan is:</p>
				<ol>
					<li>Make basic code intelligence (jump-to-def, find-references, etc.) ubiquitous</li>
					<li>Make code review continuous and more intelligent</li>
					<li>Make this all work globally, not just within a single project, so we can increase the amount/quality of available open-source code and help you avoid reinventing the wheel</li>
				</ol>

				<p>Tell everyone you know, and:</p>
				<ul>
					<li>Follow along with weekly progress updates ______________ [Subscribe] TODO</li>
					<li><Link to="/">Start using Sourcegraph</Link></li>
				</ul>
			</div>
		</div>
	);
}
