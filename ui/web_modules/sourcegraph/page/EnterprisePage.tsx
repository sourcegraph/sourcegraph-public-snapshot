import { media } from "glamor";
import * as React from "react";

import { Link } from "react-router";
import { context } from "sourcegraph/app/context";
import { Footer } from "sourcegraph/app/Footer";
import { LocationProps, Router } from "sourcegraph/app/router";
import { BGContainer, Button, FlexContainer, Heading, Panel } from "sourcegraph/components";
import { LocationStateToggleLink } from "sourcegraph/components/LocationStateToggleLink";
import { ChevronRight, Globe, Graduate, Lock, Search } from "sourcegraph/components/symbols/Primaries";
import { colors, layout, whitespace } from "sourcegraph/components/utils";
import { Nav } from "sourcegraph/home/Nav";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";

const pageWidth = 560;

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

				<div style={{ padding: whitespace[4] }}>

					<div style={{
						...layout.container.lg,
						maxWidth: 680,
						margin: "auto",
						padding: whitespace[3],
					}}>
						<Heading align="center" level={2}>Enterprise code intelligence</Heading>
						<Heading align="center" level="4" style={{ fontWeight: "normal" }}>Scalable code search and intelligence engine, for building better software faster</Heading>
					</div>

					<p style={{ textAlign: "center", marginTop: 0 }}>
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

			<div style={{ padding: whitespace[3], paddingBottom: whitespace[5] }}>
				<SectionHeader feature="Integrations">
					Sourcegraph integrates with most enterprise code hosts and review systems, including:
				</SectionHeader>
				<EnterpriseLogos />
			</div>

			<BGContainer
				img={`${context.assetsRoot}/img/enterprise/sg-ent-bg-1.svg`}
				size="350px" position="-40px -90px"
				repeat="no-repeat"
				style={{ padding: whitespace[6], backgroundColor: colors.blueGrayL3() }}>
				<div style={{
					margin: "auto",
					marginBottom: whitespace[2],
					maxWidth: 830,
					textAlign: "center",
				}}>
					<SectionHeader feature="Code intelligence">
						Building software is key to your business. Sourcegraph helps development teams build better software faster
					</SectionHeader>

					<FlexContainer justify="center" items="stretch" wrap={true} style={{ marginTop: whitespace[5] }}>
						<FeatureItem
							icon="learn"
							desc="Find and reuse existing code, within your enterprise or from open source." />
						<FeatureItem
							icon="search"
							desc="Search for code enterprise-wide, instantly." />
						<FeatureItem
							icon="globe"
							desc="Explore code with full context and history." />
					</FlexContainer>
				</div>

				<div {...layout.hide.sm} style={{ maxWidth: 850, margin: "auto", marginBottom: `-${whitespace[9]}` }}>
					<img width="100%" src={`${context.assetsRoot}/img/enterprise/sg-ent-preview-thumb-1.png`} />
				</div>

			</BGContainer>

			<BGContainer
				img={`${context.assetsRoot}/img/enterprise/sg-ent-bg-1.svg`}
				position="top right" repeat="no-repeat"
				style={{ padding: whitespace[6], paddingTop: whitespace[9], backgroundColor: colors.blueGrayL2() }}>
				<SectionHeader feature="Scalable">
					<p>
						<strong>2 million</strong> functions <br />
						<strong>300,000</strong> repositories<br />
						<strong>25 terabytes</strong> of code
					</p>
					<p>Sourcegraph Enterprise uses the same scalable architecture as Sourcegraph.com, so it will scale to meet the needs of the largest enterprises.</p>
				</SectionHeader>
			</BGContainer>

			<BGContainer
				img={`${context.assetsRoot}/img/enterprise/sg-ent-bg-2.svg`}
				position="top"
				size="cover"
				repeat="no-repeat"
				style={{ padding: whitespace[6], backgroundColor: colors.blueGray(), textAlign: "center" }}>
				<div style={{
					backgroundColor: "white",
					borderRadius: "50%",
					display: "inline-block",
					lineHeight: "1",
					margin: "auto",
					marginBottom: whitespace[3],
					padding: whitespace[2]
				}}>
					<Lock width={40} color={colors.blueGrayD1()} />
				</div>
				<SectionHeader feature="Security" style={{ color: "white" }}>
					Security is core to everything we do. Learn more at <Link style={{ color: colors.blueL2() }} to="/security">sourcegraph.com/security</Link>
				</SectionHeader>
			</BGContainer>

			<BGContainer
				img={`${context.assetsRoot}/img/enterprise/sg-ent-bg-3.svg`}
				position="center top"
				repeat="repeat-x"
				style={{ color: "white", padding: whitespace[6], backgroundColor: colors.blueGrayD1() }}>

				<Heading level={2} align="center">Sourcegraph Enterprise</Heading>
				<Heading level={5} align="center" style={{ fontWeight: "normal" }}>Use Sourcegraph with code hosted on your own servers</Heading>

				<Panel color="blue" hover={false} hoverLevel="low" style={{
					color: "white",
					maxWidth: 640,
					margin: "auto",
					marginTop: whitespace[5],
					marginBottom: whitespace[8]
				}}>
					<FlexContainer wrap={true}>
						<div style={{ padding: whitespace[4], textAlign: "center", flex: "0 0 200px" }} { ...media(layout.breakpoints.sm, { flex: "1 1 100% !important" }) }>
							<Heading level={5} align="center" style={{ marginTop: 0 }}>Enterprise</Heading>
							<Heading level={1} align="center" compact={true}>
								<sup style={{ opacity: 0.8 }}><em>$</em></sup>
								50
							</Heading>
							<span>per user, per month</span>
						</div>
						<ul style={{
							backgroundColor: colors.blueD1(),
							display: "flex",
							flexDirection: "column",
							justifyContent: "space-between",
							flex: "1 1",
							margin: 0,
							padding: whitespace[5],
							paddingLeft: whitespace[7],
							borderTopRightRadius: 4,
							borderBottomRightRadius: 4,
						}}>
							<li>Enterprise integrations: GitHub Enterprise, Phabricator, and other tools</li>
							<li>Instant, global code search and code intelligence</li>
							<li>Dedicated Customer Success Manager</li>
						</ul>
					</FlexContainer>
				</Panel>

				<Heading level={2} align="center" style={{ fontWeight: "bold" }}>Get started</Heading>
				<Heading level={5} align="center" style={{ fontWeight: "normal" }}>Contact us to get Sourcegraph on your code at your company</Heading>

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
			</BGContainer >

			<Footer />

		</div>;
	}
};

export function EnterpriseLogos(): JSX.Element {
	return <FlexContainer
		justify="between"
		wrap={true}
		style={{ padding: `${whitespace[4]} 0`, margin: "auto", marginTop: whitespace[6], maxWidth: 960 }}>
		<LogoItem>
			<img src={`${context.assetsRoot}/img/enterprise/sg-ent-gh-e-logo.svg`} />
		</LogoItem>
		<LogoItem style={{ top: -15 }}>
			<img src={`${context.assetsRoot}/img/enterprise/sg-ent-bb-logo.svg`} />
		</LogoItem>
		<LogoItem style={{ top: -8 }}>
			<img src={`${context.assetsRoot}/img/enterprise/sg-ent-gl-logo.svg`} width="124" height="auto" />
		</LogoItem>
		<LogoItem>
			<img src={`${context.assetsRoot}/img/enterprise/sg-ent-phab-logo.svg`} />
		</LogoItem>
		<LogoItem>
			<img src={`${context.assetsRoot}/img/enterprise/sg-ent-git-logo.svg`} />
		</LogoItem>
	</FlexContainer>;
}

function LogoItem({ style, children }: { style?: React.CSSProperties, children?: React.ReactNode[] }): JSX.Element {
	return <div style={{
		flex: "1 1 auto",
		padding: `${whitespace[4]} ${whitespace[3]}`,
		position: "relative",
		textAlign: "center",
		...style
	}}>{children}</div>;
};

function SectionHeader({ feature, children, style }: {
	feature: string,
	children?: React.ReactNode[],
	style?: React.CSSProperties,
}): JSX.Element {
	return <div style={{ maxWidth: pageWidth, margin: "auto", ...style }}>
		<Heading level={4} align="center" style={{ marginBottom: whitespace[3] }}>
			{feature}
		</Heading>
		<Heading level={5} align="center" style={{ fontWeight: "normal" }}>
			{children}
		</Heading>
	</div>;
};

function FeatureItem({ desc, icon, style }: {
	desc: string,
	icon: "search" | "learn" | "globe",
	style?: React.CSSProperties,
}): JSX.Element {
	const circleSx = { background: "white", fill: colors.blueGray(), borderRadius: "50%", marginRight: whitespace[3], padding: whitespace[2] };
	const iconSize = 24;

	return <FlexContainer
		items="start"
		style={{ flex: "1 1 30%", padding: whitespace[1], marginRight: whitespace[3], ...style }}
		{ ...media(layout.breakpoints.sm, { flex: "1 1 100% !important" }) }>
		{icon === "search" && <div><Search width={iconSize} style={circleSx} /></div>}
		{icon === "learn" && <div><Graduate width={iconSize} style={circleSx} /></div>}
		{icon === "globe" && <div><Globe width={iconSize} style={circleSx} /></div>}
		<p style={{ marginTop: 0, textAlign: "left" }}>{desc}</p>
	</FlexContainer>;
};
