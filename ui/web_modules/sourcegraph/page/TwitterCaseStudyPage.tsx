import { media } from "glamor";
import * as React from "react";
import { Link } from "react-router";

import { context } from "sourcegraph/app/context";
import { Footer } from "sourcegraph/app/Footer";
import { Button, FlexContainer, Heading, Hero } from "sourcegraph/components";
import { GitHubAuthButton } from "sourcegraph/components/GitHubAuthButton";
import { PageTitle } from "sourcegraph/components/PageTitle";
import { colors, layout, whitespace } from "sourcegraph/components/utils";

import * as styles from "sourcegraph/page/Page.css";

export function TwitterCaseStudyPage(): JSX.Element {

	const sprinkleMobileOpacity = 0.1;

	return (
		<div style={{
			position: "relative",
		}}>
			<PageTitle title="Sourcegraph Case Study: Twitter" />
			<Hero style={{
				backgroundColor: colors.coolGray4(),
			}}>
				<FlexContainer justify="center" items="center" direction="top_bottom">
					<img src={`${context.assetsRoot}/img/tw-case/tw-case-logos.svg`} />
					<Heading level={1}>Masters of Scale</Heading>
					<Heading level={4}>How investing in fast, semantic code browsing helps Twitter scale engineering productivity</Heading>
				</FlexContainer>
			</Hero>

			<div className={styles.content}>

				<Heading level={7} align="center">December 2016</Heading>

				<p className={styles.drop_capped}>Few companies have had the impact Twitter has had since it launched in 2006. The social networking service has been described as the “pulse of the planet,” playing a crucial role in just about every culture shift in the last decade. Behind the scenes, Twitter is also an innovator in engineering culture—a fact that becomes even more impressive when you consider that the challenges Twitter faces are formidable.</p>

				<Heading level={6} align="center">Scaling 140 characters to 313 million monthly active users.</Heading>

				<p>Twitter’s codebase is huge and highly complex. It takes a sophisticated engineering organization to build and maintain a product that supports sharing messages, images, video, and more across a global community of 313 million monthly active users (and counting). To scale its products and infrastructure, Twitter uses a combination of languages, including Java, Scala, and others. In many cases, engineering productivity actually goes down with scale as problems with communication and coordination limit the benefits of collaboration. To combat all of this, Twitter has an entire Engineering Effectiveness department focused on investing in people, processes, and tooling to boost the productivity of every Twitter developer.</p>

				<Heading
					level={4}
					underline="purple"
					style={{
						marginTop: whitespace[4],
					}}>
					The Problem:<br />Understanding and reusing existing code.
				</Heading>

				<p>Last year, a small team of engineers in Twitter’s Engineering Effectiveness department got together and discussed high-impact ways to improve developer productivity.</p>

				<p>The problem? <strong>Twitter’s codebase was so large and complex that it was hard to understand how each piece of code affected—or was affected by—everything else.</strong> Moreover, the existing internal code browser simply couldn’t handle the scale of Twitter’s codebase. The net result was that navigating the Twitter codebase was slow. And because engineers could not easily answer code-related questions on their own, they often interrupted their teammates with questions, adding to the communication and coordination overhead.</p>

				<p>Building a solution in-house was going to take too long and require too big an investment, especially given all the other infrastructure and product priorities at the time. That’s when the team, led by veteran engineering director David Keenan, started searching for out-of-the-box solutions.</p>

				<p>Sourcegraph met their requirements.</p>

				<Heading level={4} underline="purple"
					style={{
						marginTop: whitespace[4],
					}}>
					The Solution:<br />Fast, semantic code browsing.
				</Heading>
			</div>

			<Hero color="white">
				<div className={styles.container} style={{
					display: "flex",
					position: "relative",
				}}>
					<p style={{
						textAlign: "left",
						fontFamily: "Courier",
						width: "75%",
					}}>
						Sourcegraph is a <span style={{ color: colors.blue3() }}>fast</span>, <span style={{ color: colors.purple3() }}>semantic code search</span> and <span style={{ color: colors.orange2() }}>cross-reference engine</span>. It allows users to <strong>search for any function, type, or package and see how other developers use it,</strong> <span style={{ color: colors.green2() }}>globally</span>. It’s also massively scalable, with 2,000,000,000+ functions in its public code index (and growing).
					</p>
					<img src={`${context.assetsRoot}/img/tw-case/tw-case-search-example.png`} style={{
						position: "absolute",
						right: 0,
						width: "25%",
					}} />
				</div>
			</Hero>

			<div className={styles.content}>

				<p>Under Keenan’s leadership, Twitter’s team brought Sourcegraph in to boost engineering productivity. They chose Sourcegraph because they believed it could become a go-to resource in their internal suite of developer tools.</p>

				<p>During Hack Week, Keenan’s team built Scala support on the Sourcegraph API, and the tool was deployed to all of Twitter engineering within a week. “Sourcegraph is easy to integrate into your internal ecosystem because all it needs is a Git clone URL,” says Keenan.</p>

				<Heading level={6} align="center">“It works even with a completely homegrown repository hosting system like Twitter’s.”</Heading>

				<p>Sourcegraph indexes the main codebase inside of Twitter and helps developers find the answers they need in seconds, not minutes. <strong>It gives them something no IDE can: the ability to easily explore the entire codebase with all its dependencies and discuss code efficiently by linking to specific functions and types.</strong></p>

			</div>

			<div style={{
				margin: "auto",
				maxWidth: 800,
				width: "100%",
				textAlign: "center",
				overflow: "hidden",
			}}>

				<ul style={{
					listStyle: "none",
					display: "flex",
					justifyContent: "space-between",
					alignItems: "center",
				}}>
					<li>
						<img src={`${context.assetsRoot}/img/tw-case/tw-case-illu-1.png`} width="72" />
					</li>
					<li>
						<img src={`${context.assetsRoot}/img/tw-case/tw-case-illu-2.png`} width="72" />
					</li>
					<li>
						<img src={`${context.assetsRoot}/img/tw-case/tw-case-illu-3.png`} width="54" />
					</li>
					<li>
						<img src={`${context.assetsRoot}/img/tw-case/tw-case-illu-4.png`} width="72" />
					</li>
					<li>
						<img src={`${context.assetsRoot}/img/tw-case/tw-case-illu-0.png`} width="108" />
					</li>
					<li>
						<img src={`${context.assetsRoot}/img/tw-case/tw-case-illu-5.png`} width="72" />
					</li>
					<li>
						<img src={`${context.assetsRoot}/img/tw-case/tw-case-illu-6.png`} width="72" />
					</li>
					<li>
						<img src={`${context.assetsRoot}/img/tw-case/tw-case-illu-7.png`} width="54" />
					</li>
					<li>
						<img src={`${context.assetsRoot}/img/tw-case/tw-case-illu-8.png`} width="72" />
					</li>
				</ul>

			</div>

			<div className={styles.content}>

				<Heading level={4} underline="purple">The Results:<br />Time saved and limitless potential.</Heading>

				<p>Engineers across many different teams now use Sourcegraph multiple times every week, making it a key part of Twitter’s Engineering Effectiveness toolkit. In the words of one senior engineer:</p>

				<Heading level={6} align="center">“It’s very helpful to have functions viewable and clickable in the browser, so you don’t have to lose your place in your code editor.”</Heading>

				<p>The team saves time in three ways:</p>

				<p>First, Sourcegraph supports jump-to-definition across the entire Twitter repository, making it easier for developers to understand how different parts of the codebase relate to one another.</p>

				<p>Second, hover-over usage examples instantly show how existing code calls a function or uses a type.</p>

				<p>And third, Sourcegraph handles the massive scale of Twitter’s codebase, while remaining fast and efficient.</p>

				<p>Sourcegraph’s reception at Twitter has been overwhelmingly positive across the organization. “Sourcegraph is pretty amazing,” says Keenan. “It makes the Scala code so much easier to navigate. We're looking forward to getting this on Java, too.”</p>

				<p>To learn more about how Sourcegraph can help your engineering team, visit us at sourcegraph.com.</p>

				<Link target="_blank" to={`${context.assetsRoot}/img/tw-case/tw-case.pdf`}>
					<Button color="blue" style={{
						display: "flex",
						margin: "auto",
					}}>
						<div style={{
							height: 24,
							marginRight: 8,
						}}>
							<img src={`${context.assetsRoot}/img/tw-case/tw-case-dl-icon.svg`} />
						</div>
						Download Case Study PDF
					</Button>
				</Link>

				{!context.user && <div className={styles.cta}>
					<GitHubAuthButton color="purple">
						<strong>Sign up with GitHub</strong>
					</GitHubAuthButton>
				</div>}

			</div>

			<img src={`${context.assetsRoot}/img/tw-case/tw-case-blue-sprinkle.svg`} style={{
				position: "absolute",
				left: 0,
				top: 400,
				zIndex: -1,
			}}
				{ ...media(layout.breakpoints.sm, { opacity: sprinkleMobileOpacity }) }
				/>
			<img src={`${context.assetsRoot}/img/tw-case/tw-case-purple-sprinkle.svg`} style={{
				position: "absolute",
				top: 720,
				right: 0,
				zIndex: -1,
			}}
				{ ...media(layout.breakpoints.sm, { opacity: sprinkleMobileOpacity }) }
				/>
			<img src={`${context.assetsRoot}/img/tw-case/tw-case-orange-sprinkle.svg`} style={{
				position: "absolute",
				right: 120,
				bottom: 0,
				zIndex: -1,
			}}
				{ ...media(layout.breakpoints.sm, { opacity: sprinkleMobileOpacity }) }
				/>

			<Footer />

		</div>
	);
}
