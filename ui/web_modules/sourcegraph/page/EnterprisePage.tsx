import * as React from "react";

import { Link } from "react-router";
import { context } from "sourcegraph/app/context";
import { Footer } from "sourcegraph/app/Footer";
import { LocationProps, Router } from "sourcegraph/app/router";
import { BGContainer, Button, FlexContainer, Heading, Panel } from "sourcegraph/components";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { ChevronRight } from "sourcegraph/components/symbols/Primaries";
import { layout, whitespace } from "sourcegraph/components/utils";
import { Nav } from "sourcegraph/home/Nav";
import * as styles from "sourcegraph/page/Page.css";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";

export class EnterprisePage extends React.Component<LocationProps, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: Router };

	constructor(props: LocationProps) {
		super(props);
	}

	render(): JSX.Element {
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
						align="center" level={2} style={Object.assign({},
							layout.container.lg,
							{
								maxWidth: 680,
								margin: "auto",
								marginBottom: whitespace[9],
								marginTop: whitespace[3],
								padding: whitespace[3],
							}
						)}>Enterprise code intelligence<div style={{ fontSize: "24px" }}>Semantic code search and cross-reference engine for your code.</div></Heading>

					<p style={{ textAlign: "center" }}>
						<LocationStateToggleLink href="/join" modalName="join" location={this.props.location} onToggle={(v) => v && Events.JoinModal_Initiated.logEvent({ page_name: location.pathname, location_on_page: "Header" })}>
							<Button color="orange" style={{
								margin: whitespace[3],
								paddingLeft: whitespace[3],
								paddingRight: whitespace[3],
							}}>
								Get started <ChevronRight />
							</Button>
						</LocationStateToggleLink>
					</p>
				</div>
			</BGContainer>

			<div style={{ paddingTop: "20px", paddingBottom: "20px" }}>
				<Heading level={4} align="center" style={{ fontWeight: "bold", maxWidth: 680, margin: "auto" }}>Integrations</Heading>
				<Heading level={4} align="center" style={{ fontWeight: "normal", maxWidth: 680, margin: "auto" }}>Sourcegraph integrates with most enterprise code hosts and review systems, including:</Heading>

				<FlexContainer justify="center" items="center" wrap={true} style={{ padding: whitespace[4] }}>
					<div {...layout.hide.sm} style={{ flex: 1 }} />
					<div style={{ display: "flex", justifyContent: "center", flex: 1, padding: whitespace[2], }}>
						<img src={`${context.assetsRoot}/img/enterprise/sg-ent-gh-e-logo.svg`} />
					</div>
					<div style={{ display: "flex", justifyContent: "center", flex: 1, padding: whitespace[2] }}>
						<img src={`${context.assetsRoot}/img/enterprise/sg-ent-bb-logo.svg`} />
					</div>
					<div style={{ display: "flex", justifyContent: "center", flex: 1, padding: whitespace[2] }}>
						<img src={`${context.assetsRoot}/img/enterprise/sg-ent-gl-logo.svg`} />
					</div>
					<div style={{ display: "flex", justifyContent: "center", flex: 1, padding: whitespace[2] }}>
						<img src={`${context.assetsRoot}/img/enterprise/sg-ent-phab-logo.svg`} />
					</div>
					<div style={{ display: "flex", justifyContent: "center", flex: 1, padding: whitespace[2] }}>
						<img src={`${context.assetsRoot}/img/enterprise/sg-ent-git-logo.svg`} />
					</div>
					<div {...layout.hide.sm} style={{ flex: 1 }} />
				</FlexContainer>
			</div>

			<BGContainer img={`${context.assetsRoot}/img/enterprise/sg-ent-bg-1.svg`} position="top left" repeat="no-repeat" style={{ padding: whitespace[6], backgroundColor: "#cad4e0" }}>
				<div style={{
					margin: "auto",
					marginBottom: whitespace[2],
					maxWidth: 680,
					textAlign: "center",
				}}>

					<Heading level={4} align="center" style={{ fontWeight: "bold", maxWidth: "680px", margin: "auto" }}>Code intelligence</Heading>
					<Heading level={4} align="center" style={{ fontWeight: "normal", maxWidth: "680px", margin: "auto" }}>Part of <Link to="/plan">our mission</Link> is to make code intelligence ubiquitous, in every editor and every language.</Heading>

					<FlexContainer justify="center" style={{ marginTop: "20px" }}>
						<div style={{ padding: whitespace[1] }}>Find and reuse existing code, from your enterprise or from open source.</div>
						<div style={{ padding: whitespace[1] }}>Search for code enterprise-wide, instantly.</div>
						<div style={{ padding: whitespace[1] }}>Explore a code with full context and history.</div>
						<div style={{ padding: whitespace[1] }}>Find and answer questions in code without interrupting flow.</div>
					</FlexContainer>
				</div>

				<div {...layout.hide.sm} style={{ maxWidth: "850px", margin: "auto" }}>
					<img width="100%" src={`${context.assetsRoot}/img/enterprise/sg-ent-preview-thumb-1.png`} />
				</div>

			</BGContainer>

			<BGContainer img={`${context.assetsRoot}/img/enterprise/sg-ent-bg-1.svg`} position="top right" repeat="no-repeat" style={{ padding: whitespace[6], backgroundColor: "#97a9c4" }}>
				<Heading level={4} align="center" style={{ fontWeight: "bold", maxWidth: "680px", margin: "auto" }}>Scalable</Heading>
				<Heading level={2} align="center" style={{ fontWeight: "normal", maxWidth: "680px", margin: "auto" }}>1,999,999,999 functions (and growing) on the public code graph</Heading>
				<Heading level={4} align="center" style={{ fontWeight: "normal", maxWidth: "680px", margin: "auto" }}>Sourcegraph Enterprise is the exact same codebase and architecture as Sourcegraph.com, which scales to hundreds of thousands of repositories and terabytes of code.</Heading>
			</BGContainer>

			<BGContainer img={`${context.assetsRoot}/img/enterprise/sg-ent-bg-2.svg`} position="top" size="cover" repeat="no-repeat" style={{ padding: whitespace[6], backgroundColor: "#637fa6" }}>
				<Heading color="white" level={4} align="center" style={{ fontWeight: "bold", maxWidth: "680px", margin: "auto" }}>Security</Heading>
				<Heading color="white" level={4} align="center" style={{ fontWeight: "normal", maxWidth: "680px", margin: "auto" }}>Securty is core to everything we do. Learn more at <Link style={{ color: "#74bef6" }} to="/security">sourcegraph.com/security</Link>.</Heading>
			</BGContainer>

			<BGContainer img={`${context.assetsRoot}/img/enterprise/sg-ent-bg-3.svg`} position="center top" repeat="no-repeat" style={{ padding: whitespace[6], backgroundColor: "#445876" }}>
				<Heading color="white" level={2} align="center" style={{ fontWeight: "bold", maxWidth: "680px", margin: "auto" }}>Sourcegraph Enterprise</Heading>
				<Heading color="white" level={4} align="center" style={{ fontWeight: "normal", maxWidth: "680px", margin: "auto" }}>Use Sourcegraph with code hosted on your own servers.</Heading>

				<div style={{ marginTop: whitespace[4], marginBottom: whitespace[4] }}>
					<FlexContainer justify="center" items="center" style={{ maxWidth: "680px", margin: "auto" }}>
						<Panel color="blue" hover={false} className={styles.plan_panel || ""} style={{ flex: 1 }}>
							<Heading level={3} color="white" align="center">Enterprise</Heading>
							<Heading level={1} color="white" align="center" style={{}}><span style={{ opacity: 0.8, fontStyle: "italic", verticalAlign: "2rem" }}>$</span><span style={{ fontSize: "4.5rem" }}>50</span></Heading>
							<span style={{ fontSize: "1.0rem" }}>per user, per month</span>
						</Panel>
						<div style={{ flex: 3, backgroundColor: "#2f72b0", color: "white" }}>
							<ul style={{ display: "flex", flexDirection: "column" }}>
								<li style={{ flex: 1, padding: whitespace[2] }}>Enterprise integrations: GitHub Enterprise, Phabricator, and other tools</li>
								<li style={{ flex: 1, padding: whitespace[2] }}>Instant, global code search and code intelligence</li>
								<li style={{ flex: 1, padding: whitespace[2] }}>Dedicated Customer Success Manager</li>
							</ul>
						</div>
					</FlexContainer>
				</div>

				<Heading color="white" level={2} align="center" style={{ fontWeight: "bold" }}>Get started</Heading>
				<Heading color="white" level={4} align="center" style={{ fontWeight: "normal" }}>Contact us to get Sourcegraph on your code at your company.</Heading>

				<p style={{ textAlign: "center" }}>
					<a href="mailto:sales@sourcegraph.com">
						<Button color="green" style={{
							margin: whitespace[3],
							paddingLeft: whitespace[3],
							paddingRight: whitespace[3],
						}}>
							Contact us <ChevronRight />
						</Button>
					</a>
				</p>
			</BGContainer>

			<Footer />

		</div>;
	}
};
