import * as React from "react";

import { context } from "sourcegraph/app/context";
import { LocationProps } from "sourcegraph/app/router";
import { Button, FlexContainer, Heading, Hero, User } from "sourcegraph/components";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { PageTitle } from "sourcegraph/components/PageTitle";
import * as base from "sourcegraph/components/styles/_base.css";
import { whitespace } from "sourcegraph/components/utils";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";

export function AboutPage({ location }: LocationProps): JSX.Element {
	const boardMemberSx = { marginBottom: whitespace[4], marginRight: whitespace[4] };
	return (
		<div>
			<PageTitle title="About" />
			<Hero pattern="objects" color="blue" className={base.pv5}>
				<FlexContainer style={{ margin: "auto", maxWidth: 640 }}>
					<Heading level={3} color="white">The pace at which humans can write code is the only thing that stands between us and flying cars, a habitat on Mars and a cure for cancer.</Heading>
				</FlexContainer>
			</Hero>
			<FlexContainer direction="top_bottom" style={{
				maxWidth: 640,
				margin: "auto",
			}}>
				<Heading level={4} underline="purple" style={{ marginTop: whitespace[4] }}>Master Plan</Heading>
				<p>We believe code intelligence can help bring <strong>the future sooner.</strong> Our long-term vision is to make it so everyone, in every community, in every country, and in every industry can create products using the best technology. Here is what we are working on now to help this happen:</p>
				<ul>
					<li>Make basic code intelligence ubiquitous</li>
					<li>Make code review continuous and intelligent</li>
					<li>Increase the amount and quality of open-source code</li>
				</ul>

				<p>Read more at <a href="/plan">sourcegraph.com/plan</a>.</p>
				<br />

				<Heading level={4} underline="purple">Values</Heading>

				<FlexContainer>
					<ValueCol img="img/about/about-hash-people.png">
						<p><strong>#people</strong> come first.<br />Together we are advancing technology. We will attract, hire and retain the best teammates in the world and treat everyone in a first-class manner.</p>
					</ValueCol>
					<ValueCol img="img/about/about-hash-journey.png">
						<p><strong>#journey</strong> is the collection of moments, experiences, and memories that the team shares as we make #progress: the light moments, the laughter, the team coming together to solve a problem, etc.</p>
					</ValueCol>
					<ValueCol img="img/about/about-hash-progress.png">
						<p><strong>#progress</strong> is the continuous march toward achieving our mission as a team: the milestones, the successes, the breakthroughs, etc.</p>
					</ValueCol>
				</FlexContainer>

				<Heading level={4} underline="purple">Founders</Heading>
				<img width="100%" src={`${context.assetsRoot}/img/about/about-founders.png`} />

				<FlexContainer>
					<div style={{
						paddingRight: whitespace[3],
					}}>
						<Heading level={5}>Quinn Slack, CEO</Heading>
						<div style={{
							marginBottom: whitespace[2],
						}}>
							<a target="_blank" href="https://www.linkedin.com/in/quinnslack">
								<img src={`${context.assetsRoot}/img/about/about-li-icon.svg`} style={{
									marginRight: whitespace[3],
								}} />
							</a>
							<a target="_blank" href="https://github.com/sqs">
								<img src={`${context.assetsRoot}/img/about/about-gh-icon.svg`} style={{
									marginRight: whitespace[3],
								}} />
							</a>
							<a target="_blank" href="https://twitter.com/sqs">
								<img src={`${context.assetsRoot}/img/about/about-tw-icon.svg`} />
							</a>
						</div>
						Quinn Slack is CEO and co-founder of Sourcegraph. Prior to Sourcegraph, Quinn co-founded Blend Labs, an enterprise technology company with over 100 employees dedicated to improving home lending. At Palantir Technologies he created a technology platform to help two of the top five U.S. banks recover from the housing crisis. He was the first employee and developer at Bleacher Report after graduating from high school. Quinn graduated with a BS in Computer Science from Stanford.
					</div>
					<div style={{
						paddingLeft: whitespace[3],
					}}>
						<Heading level={5}>Beyang Liu, CTO</Heading>
						<div style={{
							marginBottom: whitespace[2],
						}}>
							<a target="_blank" href="https://www.linkedin.com/in/beyang-liu-07651227">
								<img src={`${context.assetsRoot}/img/about/about-li-icon.svg`} style={{
									marginRight: whitespace[3],
								}} />
							</a>
							<a target="_blank" href="https://github.com/beyang">
								<img src={`${context.assetsRoot}/img/about/about-gh-icon.svg`} style={{
									marginRight: whitespace[3],
								}} />
							</a>
							<a target="_blank" href="https://twitter.com/beyang">
								<img src={`${context.assetsRoot}/img/about/about-tw-icon.svg`} />
							</a>
						</div>
						Beyang Liu is CTO and co-founder of Sourcegraph. Previous to Sourcegraph, Beyang worked as a software engineer at Palantir Technologies where he developed new products on a small, customer-facing team working with Fortune 100 clients. Beyang is a patent holder in machine learning and has contributed to many open-source projects. Beyang graduated from Stanford with a BS in Computer Science where he was a published research assistant and software development intern at Google.
					</div>
				</FlexContainer>

				<Heading level={4} underline="purple" style={{ marginTop: whitespace[4] }}>
					Investors
				</Heading>
				<img width="100%" src={`${context.assetsRoot}/img/about/about-investors.png`} />

				<Heading level={4} underline="purple" style={{ marginTop: whitespace[4] }}>
					Board of Directors
				</Heading>

				<FlexContainer wrap={true}>

					<FlexContainer wrap={true} direction="top_bottom">
						<User nickname="Scott Raney" email="Redpoint Ventures" style={boardMemberSx} avatar={`${context.assetsRoot}/img/about/about-board-scott.png`} />
						<User nickname="Daniel Friedland" email="Goldcrest Capital" style={boardMemberSx} avatar={`${context.assetsRoot}/img/about/about-board-daniel.png`} />
					</FlexContainer>

					<FlexContainer wrap={true} direction="top_bottom">
						<User nickname="Quinn Slack" email="Sourcegraph" style={boardMemberSx} avatar={`${context.assetsRoot}/img/about/about-board-quinn.png`} />
						<User nickname="Beyang Liu" email="Sourcegraph" style={boardMemberSx} avatar={`${context.assetsRoot}/img/about/about-board-beyang.png`} />
					</FlexContainer>

				</FlexContainer>

				<Heading level={4} underline="purple" style={{ marginTop: whitespace[4] }}>
					Advisors and Supporters
					</Heading>

				<FlexContainer wrap={true}>

					<FlexContainer wrap={true} direction="top_bottom">
						<User nickname="Adam Ross" email="Goldcrest Capital" style={boardMemberSx} avatar={`${context.assetsRoot}/img/about/about-advisor-adam.png`} />
						<User nickname="Bear Douglas" email="Developer Evangelist" style={boardMemberSx} avatar={`${context.assetsRoot}/img/about/about-advisor-bear.jpg`} />
						<User nickname="Jens Pillgram-Larsen" email="LinkedIn" style={boardMemberSx} avatar={`${context.assetsRoot}/img/about/about-advisor-jens.jpg`} />
						<User nickname="James Waldrop" email="Fitbit" style={boardMemberSx} avatar={`${context.assetsRoot}/img/about/about-advisor-james.jpg`} />
						<User nickname="Mark Striebeck" email="Google" style={boardMemberSx} avatar={`${context.assetsRoot}/img/about/about-advisor-mark.jpg`} />
					</FlexContainer>

					<FlexContainer wrap={true} direction="top_bottom">
						<User nickname="Audrey Maclean" email="Entrepreneur" style={boardMemberSx} avatar={`${context.assetsRoot}/img/about/about-advisor-audrey.jpg`} />
						<User nickname="Bill Malloy" email="Technologist & Financier" style={boardMemberSx} avatar={`${context.assetsRoot}/img/about/about-advisor-bill.jpg`} />
						<User nickname="John Vrionis" email="Lightspeed Venture Partners" style={boardMemberSx} avatar={`${context.assetsRoot}/img/about/about-advisor-john.jpg`} />
						<User nickname="Kevin Mahaffey" email="Lookout" style={boardMemberSx} avatar={`${context.assetsRoot}/img/about/about-advisor-kevin.jpg`} />
						<User nickname="Peter Hubert" email="Lux Capital" style={boardMemberSx} avatar={`${context.assetsRoot}/img/about/about-advisor-peter.jpg`} />
					</FlexContainer>

				</FlexContainer>

				{!context.user && <div style={{
					textAlign: "center",
					marginTop: whitespace[5],
					marginBottom: whitespace[5],
				}}>
					<LocationStateToggleLink href="/join" modalName="join" location={location} onToggle={(v) => v && Events.JoinModal_Initiated.logEvent({ page_name: location.pathname, location_on_page: "Footer" })}>
						<Button color="purple">
							Sign up for free
						</Button>
					</LocationStateToggleLink>
				</div>}

			</FlexContainer>

		</div>
	);
};

export function ValueCol({ img, children }: {
	img: string,
	children?: React.ReactNode[],
}): JSX.Element {
	return <div style={{ flex: 1, padding: whitespace[3] }}>
		<img width="100%" src={`${context.assetsRoot}/${img}`} />
		{children}
	</div>;
}
