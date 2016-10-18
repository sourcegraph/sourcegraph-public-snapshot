import {hover, media, style} from "glamor";
import * as React from "react";
import {Link} from "react-router";
import {InjectedRouter} from "react-router";
import {context} from "sourcegraph/app/context";
import {Footer} from "sourcegraph/app/Footer";
import {BGContainer, Button, FlexContainer, Heading, Panel} from "sourcegraph/components";
import {LocationStateToggleLink} from "sourcegraph/components/LocationStateToggleLink";
import {LocationStateModal, dismissModal} from "sourcegraph/components/Modal";
import {ChevronRight} from "sourcegraph/components/symbols/Zondicons";
import {colors, layout, whitespace} from "sourcegraph/components/utils";
import {BetaInterestForm} from "sourcegraph/home/BetaInterestForm";
import {FeatureCarousel} from "sourcegraph/home/FeatureCarousel";
import {Nav} from "sourcegraph/home/Nav";
import {Location} from "sourcegraph/Location";

interface HomeProps { location: Location; }

export class Home extends React.Component<HomeProps, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: { router: InjectedRouter };

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
			"-webkit-overflow-scrolling": "touch",
		}}>

			<BGContainer
				img={`${context.assetsRoot}/img/Homepage/bg-circuit.svg`}
				style={{boxShadow: "inset 0 -30px 100px white"}}>

				<Nav location={this.props.location} style={{padding: whitespace[4]}} context={this.context} />

				<div style={layout.container}>

					<Heading
						align="center" level={1} style={Object.assign({},
						layout.container,
						{
							maxWidth: 640,
							marginBottom: whitespace[5],
							marginTop: whitespace[3],
							padding: whitespace[3],
						}
					)}>Read code on the web with the power of an IDE</Heading>

					<p style={{textAlign: "center"}}>
						<LocationStateToggleLink href="/join" modalName="join" location={this.props.location}>
							<Button color="orange" style={{
								margin: whitespace[3],
								paddingLeft: whitespace[3],
								paddingRight: whitespace[3],
							}}>
								Sign up for free
							</Button>
						</LocationStateToggleLink> or
						<Link to="/about" style={{margin: whitespace[2]}}><strong>learn more</strong></Link>
					</p>
				</div>
			</BGContainer>

			<div style={{marginBottom: whitespace[5], marginTop: whitespace[6]}}>
				<Heading level={3} align="center" style={{fontWeight: "normal"}}>
					Read code smarter and faster. Get more done.
				</Heading>
				<FeatureCarousel assetsURL={context.assetsRoot} />
			</div>

			<BGContainer img={`${context.assetsRoot}/img/Homepage/bg-sourcesprinkles.svg`} position="center top" size="cover">

				<div style={{
					margin: "auto",
					marginBottom: whitespace[2],
					maxWidth: 340,
					textAlign: "center",
				}}>

					<img title="Go supported" width="40" src={`${context.assetsRoot}/img/Homepage/logo/go2.svg`} />
					<FlexContainer justify="around" style={{
						height: 100,
						opacity: 0.5,
						margin: "auto",
						width: "100%",
					}}>

						<img title="Java coming soon" width="24" src={`${context.assetsRoot}/img/Homepage/logo/java.svg`} />
						<img title="Typescript coming soon" width="24" src={`${context.assetsRoot}/img/Homepage/logo/typescript.svg`} />
						<img title="JavaScript coming soon"width="24"  src={`${context.assetsRoot}/img/Homepage/logo/js.svg`} />
						<img title="Python coming soon"width="24"  src={`${context.assetsRoot}/img/Homepage/logo/python.svg`} />
						<img title="PHP coming soon" width="24" src={`${context.assetsRoot}/img/Homepage/logo/php.svg`} />
						<img title="Scala coming soon" width="24" src={`${context.assetsRoot}/img/Homepage/logo/scala.svg`} />
					</FlexContainer>

					<Heading level={4} align="center" style={{fontWeight: "normal"}}>Support for Go</Heading>

					<p style={{
						color: colors.coolGray3(),
						paddingLeft: whitespace[4],
						paddingRight: whitespace[4],
					}}>
						TypeScript, JavaScript, Java, Python, Ruby, Scala, C, PHP, C++, and more coming soon...
					</p>
					<LocationStateToggleLink href="/beta" modalName="beta" location={this.props.location}>
						<strong>
							Join the beta list
							<ChevronRight color={colors.blueText()} width={7} style={{marginLeft: 4}} />
						</strong>
					</LocationStateToggleLink>

				</div>

				{this.props.location.state && (this.props.location.state as any).modal === "beta" &&
					<LocationStateModal modalName="beta" location={this.props.location} router={this.context.router}>
						<Panel style={{
							maxWidth: 440,
							minWidth: 320,
							maxHeight: "85%",
							padding: whitespace[4],
							margin: "auto",
							marginTop: "20vh",
						}}>
							<Heading level={4} align="center">Join the Sourcegraph beta</Heading>
							<BetaInterestForm
								style={{width: "100%"}}
								loginReturnTo="/"
								onSubmit={dismissModal("beta", this.props.location, (this.context as any).router)} />
						</Panel>
					</LocationStateModal>
				}

				<div style={{paddingTop: whitespace[5], paddingBottom: whitespace[4]}}>
					<Panel hoverLevel="high" hover={false}
						style={{
							margin: "auto",
							maxWidth: 960,
							padding: whitespace[4],
						}}>
							<FlexContainer wrap={true}>
								<Heading align="left" level={4} underline="purple" style={{flex: "0 0 240px"}} >
									Used by developers everywhere
								</Heading>
								<FlexContainer justify="end" style={{flex: "1 1 60%"}}>
									<img style={{marginBottom: "10px", marginRight: whitespace[4]}} src={`${context.assetsRoot}/img/Homepage/logo/twitter.svg`} />
									<img style={{marginBottom: "9px", marginRight: whitespace[4]}} src={`${context.assetsRoot}/img/Homepage/logo/red-hat.svg`} />
									<img style={{marginBottom: "7px", marginRight: whitespace[4]}} src={`${context.assetsRoot}/img/Homepage/logo/daily-motion.svg`} />
									<img style={{marginBottom: "5px"}} src={`${context.assetsRoot}/img/Homepage/logo/progressly.svg`} />
								</FlexContainer>
							</FlexContainer>

							<FlexContainer wrap={true} justify="between">
								<Tweet>
									<blockquote className="twitter-tweet" lang="en"><p lang="en" dir="ltr">Just found out <a href="https://twitter.com/srcgraph">@srcgraph</a> ! HUGE productivity gain. Great work ! Waiting for more language support.</p>&mdash; Dharmesh Kakadia (@dharmeshkakadia) <a href="https://twitter.com/dharmeshkakadia/status/738874411437035520">June 3, 2016</a></blockquote>
								</Tweet>
								<Tweet>
									<blockquote className="twitter-tweet" lang="en"><p lang="en" dir="ltr">The <a href="https://twitter.com/srcgraph">@srcgraph</a> Chrome extension for GitHub is the best! <a href="https://t.co/CKweAOfbsQ">https://t.co/CKweAOfbsQ</a></p>&mdash; Julius Volz (@juliusvolz) <a href="https://twitter.com/juliusvolz/status/748095329564778496">June 29, 2016</a></blockquote>
								</Tweet>
								<Tweet>
									<blockquote className="twitter-tweet" lang="en"><p lang="en" dir="ltr">Used <a href="https://twitter.com/srcgraph">@srcgraph</a> jump-to-definition across 3 projects, 2 langs, finally landing deep in Golang src. Took &lt; 10 min to pin down the issue. 💪🏼</p>&mdash; Gabriel Monroy (@gabrtv) <a href="https://twitter.com/gabrtv/status/738861622882508801">June 3, 2016</a></blockquote>
								</Tweet>
								<Tweet>
									<blockquote className="twitter-tweet" lang="en"><p lang="en" dir="ltr">Sourcegraph is the cross reference to end all cross references</p>&mdash; Erik Hollensbe (@erikhollensbe) <a href="https://twitter.com/erikhollensbe/status/738880970909089793">June 3, 2016</a></blockquote>
								</Tweet>
							</FlexContainer>
					</Panel>
				</div>

				<div style={Object.assign({},
					layout.container,
					{
						maxWidth: 600,
						marginTop: whitespace[5],
						padding: whitespace[3],
						paddingBottom: whitespace[6],
						textAlign: "center",
					}
				)}>
					<Heading align="center" level={3}>
						Understand code smarter and faster with Sourcegraph
					</Heading>
					<Heading align="center" color="gray" level={4} style={{
						fontWeight: "normal",
						marginTop: whitespace[3],
					}}>
						Free for public and personal private code
					</Heading>

					<LocationStateToggleLink href="/join" modalName="join" location={this.props.location}>
						<Button color="orange" style={{
							marginTop: whitespace[4],
							paddingLeft: whitespace[4],
							paddingRight: whitespace[4],
						}}>Sign up for free</Button>
					</LocationStateToggleLink>
				</div>

				<Footer />

			</BGContainer>

		</div>;
	}
}

interface TweetProps { children?: React.ReactNode[]; }

function Tweet({children}: TweetProps): JSX.Element {
	return <div
		{...style({flex: "0 0 49%", maxWidth: "49%"})}
		{...media(layout.breakpoints["sm"], {flex: "0 0 100%", maxWidth: "100%"})}>{children}</div>;
}
