import { media, style } from "glamor";
import * as React from "react";
import { Link } from "react-router";

import { context } from "sourcegraph/app/context";
import { Footer } from "sourcegraph/app/Footer";
import { Router, RouterLocation } from "sourcegraph/app/router";
import { BGContainer, Button, FlexContainer, Heading, Panel } from "sourcegraph/components";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { LocationStateModal, dismissModal } from "sourcegraph/components/Modal";
import { ChevronRight } from "sourcegraph/components/symbols/Primaries";
import { colors, layout, whitespace } from "sourcegraph/components/utils";
import { BetaInterestForm } from "sourcegraph/home/BetaInterestForm";
import { FeatureCarousel } from "sourcegraph/home/FeatureCarousel";
import { Nav } from "sourcegraph/home/Nav";
import { Events, PAGE_DASHBOARD } from "sourcegraph/tracking/constants/AnalyticsConstants";

interface HomeProps { location: RouterLocation; }

export class Home extends React.Component<HomeProps, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };

	constructor(props: HomeProps) {
		super(props);
	}

	componentDidMount(): void {
		/* The twitter-wjs module loaded here is only used for this page
		   That's why it's in this file and not in app/templates/scripts.html */
		let script = document.createElement("script");
		script.id = "twitter-wjs";
		script.src = "//platform.twitter.com/widgets.js";
		script.charset = "utf-8";
		document.body.appendChild(script);
	}

	render(): JSX.Element | null {
		return <div style={{
			backgroundColor: "white",
			overflowX: "hidden",
			WebkitOverflowScrolling: "touch",
		}}>

			<BGContainer
				img={`${context.assetsRoot}/img/Homepage/bg-circuit.svg`}
				style={{ boxShadow: "inset 0 -30px 100px white" }}>

				<Nav location={this.props.location} style={{ padding: whitespace[5] }} context={this.context} />

				<div style={layout.container.lg}>

					<Heading
						align="center" level={1} style={Object.assign({},
							layout.container.lg,
							{
								maxWidth: 680,
								marginBottom: whitespace[8],
								marginTop: whitespace[3],
								padding: whitespace[3],
							}
						)}>The global code graph</Heading>

					<p style={{ textAlign: "center" }}>
						<LocationStateToggleLink href="/join" modalName="join" location={this.props.location} onToggle={(v) => v && Events.JoinModal_Initiated.logEvent({ page_name: location.pathname, location_on_page: "Header" })}>
							<Button color="orange" style={{
								margin: whitespace[3],
								paddingLeft: whitespace[3],
								paddingRight: whitespace[3],
							}}>
								Sign up
							</Button>
						</LocationStateToggleLink> or
						<Link to="/plan" style={{ margin: whitespace[2] }}><strong>learn more</strong></Link>
					</p>
				</div>
			</BGContainer>

			<div style={{ marginBottom: whitespace[8], marginTop: whitespace[9] }}>
				<Heading level={3} align="center" style={{ fontWeight: "normal" }}>
					Code Intelligence for public and private code.
				</Heading>
				<FeatureCarousel assetsURL={context.assetsRoot} />
			</div>

			<BGContainer img={`${context.assetsRoot}/img/Homepage/bg-sourcesprinkles.svg`} position="center top" size="cover">

				<Heading level={4} align="center" style={{ fontWeight: "normal" }}>
					Language support:
				</Heading>

				<div style={{
					margin: "auto",
					marginBottom: whitespace[2],
					maxWidth: 640,
					textAlign: "center",
				}}>

					<FlexContainer justify="center">

						<Heading level={5} style={{
							paddingTop: whitespace[1],
							paddingRight: whitespace[3],
							paddingBottom: whitespace[1],
							paddingLeft: whitespace[3],
							backgroundColor: colors.greenL2(),
							color: colors.greenD1(),
							borderRadius: 20,
							marginRight: whitespace[3],
						}}>
							Go
						</Heading>
						<Heading level={5} style={{
							paddingTop: whitespace[1],
							paddingRight: whitespace[3],
							paddingBottom: whitespace[1],
							paddingLeft: whitespace[3],
							backgroundColor: colors.orangeL2(),
							color: colors.orangeD1(),
							borderRadius: 20,
							marginRight: whitespace[3],
						}}>
							<div>Java</div>
						</Heading>
						<Heading level={5} style={{
							paddingTop: whitespace[1],
							paddingRight: whitespace[3],
							paddingBottom: whitespace[1],
							paddingLeft: whitespace[3],
							backgroundColor: colors.purpleL2(),
							color: colors.purpleD1(),
							borderRadius: 20,
							marginRight: whitespace[3],
						}}>
							<div>Android</div>
						</Heading>
						<Heading level={5} style={{
							paddingTop: whitespace[1],
							paddingRight: whitespace[3],
							paddingBottom: whitespace[1],
							paddingLeft: whitespace[3],
							backgroundColor: colors.blueL2(),
							color: colors.blueD1(),
							borderRadius: 20,
							marginRight: whitespace[3],
						}}>
							<div title="Alpha">TypeScript<sup>&alpha;</sup></div>
						</Heading>
					</FlexContainer>

					<FlexContainer justify="center" wrap={true}>
						<Heading level={6} style={{ fontWeight: "normal", lineHeight: 2.25, }}>and more coming soon&hellip;</Heading>

					</FlexContainer>

					<LocationStateToggleLink href="/beta" modalName="menuBeta" location={this.props.location} onToggle={(v) => v && Events.BetaModal_Initiated.logEvent({ page_name: location.pathname, location_on_page: PAGE_DASHBOARD })}>
						<div style={{ marginTop: whitespace[3] }}>
							<strong>
								Join the beta list
								<ChevronRight width={16} />
							</strong>
						</div>
					</LocationStateToggleLink>

				</div>

				{this.props.location.state && (this.props.location.state as any).modal === "beta" &&
					<LocationStateModal padded={false} modalName="beta" title="Sign up for our beta">
						<BetaInterestForm
							style={{ width: "100%" }}
							location={this.props.location}
							onSubmit={dismissModal("beta", this.context.router)} />
					</LocationStateModal>
				}

				<div style={{ paddingTop: whitespace[8], paddingBottom: whitespace[5] }}>
					<Panel hoverLevel="high" hover={false}
						style={{
							margin: "auto",
							maxWidth: 960,
							padding: whitespace[5],
						}}>
						<FlexContainer wrap={true} items="center">
							<Heading align="left" level={4} underline="purple" style={{ flex: "0 0 240px" }} >
								Used by developers everywhere
								</Heading>
							<FlexContainer items="center" justify="end" style={{ flex: "1 1 60%" }}>
								<img style={{ marginRight: whitespace[5] }} src={`${context.assetsRoot}/img/Homepage/logo/twitter.svg`} height="24" {...layout.hide.sm} />
								<img style={{ marginRight: whitespace[5] }} src={`${context.assetsRoot}/img/Homepage/logo/red-hat.svg`} height="32" {...layout.hide.sm} />
								<img style={{ marginTop: "5px", marginRight: whitespace[5] }} src={`${context.assetsRoot}/img/Homepage/logo/daily-motion.svg`} height="24" {...layout.hide.sm} />
								<img style={{ marginTop: "8px" }} src={`${context.assetsRoot}/img/Homepage/logo/progressly.svg`} height="17" {...layout.hide.sm} />
							</FlexContainer>
						</FlexContainer>

						<FlexContainer wrap={true} justify="between">
							<Tweet>
								<blockquote className="twitter-tweet">
									<p lang="en" dir="ltr">I LOVE <a href="https://twitter.com/srcgraph">@srcgraph</a> SO MUCH! I&#39;m spelunking in the kubernetes nginx-ingress-controller codebase and it is such a big timesaver.</p>&mdash; Cole Mickens (@colemickens) <a href="https://twitter.com/colemickens/status/768621780076417024">August 25, 2016</a></blockquote>
							</Tweet>
							<Tweet>
								<blockquote className="twitter-tweet" lang="en"><p lang="en" dir="ltr">Used <a href="https://twitter.com/srcgraph">@srcgraph</a> jump-to-definition across 3 projects, 2 langs, finally landing deep in Golang src. Took &lt; 10 min to pin down the issue. 💪🏼</p>&mdash; Gabriel Monroy (@gabrtv) <a href="https://twitter.com/gabrtv/status/738861622882508801">June 3, 2016</a></blockquote>
							</Tweet>
							<Tweet>
								<blockquote className="twitter-tweet"><p lang="en" dir="ltr">&quot;Do you use source graph?&quot; My multi-times a day question for when I talk to people about <a href="https://twitter.com/github">@github</a> projects. <a href="https://twitter.com/hashtag/devBetter?src=hash">#devBetter</a> <a href="https://twitter.com/srcgraph">@srcgraph</a></p>&mdash; Chase Adams (@chaseadamsio) <a href="https://twitter.com/chaseadamsio/status/774284535655653376">September 9, 2016</a></blockquote>
							</Tweet>
							<Tweet>
								<blockquote className="twitter-tweet" data-cards="hidden"><p lang="en" dir="ltr">Learning a new code base using <a href="https://twitter.com/srcgraph">@srcgraph</a> is extra dope! <a href="https://t.co/MKbac0RB0B">https://t.co/MKbac0RB0B</a> <a href="https://t.co/6YWeYyyYZo">pic.twitter.com/6YWeYyyYZo</a></p>&mdash; Kelsey Hightower (@kelseyhightower) <a href="https://twitter.com/kelseyhightower/status/791084672797122561">October 26, 2016</a></blockquote>
							</Tweet>
						</FlexContainer>
					</Panel>
				</div>

				<div style={Object.assign({},
					layout.container.lg,
					{
						maxWidth: 600,
						marginTop: whitespace[8],
						padding: whitespace[3],
						paddingBottom: whitespace[9],
						textAlign: "center",
					}
				)}>
					<Heading align="center" level={3}>
						Read code smarter and faster with Sourcegraph
					</Heading>
					<Heading align="center" color="gray" level={4} style={{
						fontWeight: "normal",
						marginTop: whitespace[3],
					}}>
						Get started for free
					</Heading>

					<LocationStateToggleLink href="/join" modalName="join" location={this.props.location} onToggle={(v) => v && Events.JoinModal_Initiated.logEvent({ page_name: location.pathname, location_on_page: "Footer" })}>
						<Button color="orange" style={{
							marginTop: whitespace[5],
							paddingLeft: whitespace[5],
							paddingRight: whitespace[5],
						}}>Sign up for free</Button>
					</LocationStateToggleLink>
				</div>

				<Footer />

			</BGContainer>

		</div>;
	}
}

interface TweetProps { children?: React.ReactNode[]; }

function Tweet({ children }: TweetProps): JSX.Element {
	return <div
		{...style({ flex: "0 0 49%", maxWidth: "49%" }) }
		{...media(layout.breakpoints.sm, { flex: "0 0 100%", maxWidth: "100%" }) }>{children}</div>;
}
