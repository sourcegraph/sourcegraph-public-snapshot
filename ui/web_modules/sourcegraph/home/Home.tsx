// tslint:disable

import * as React from "react";
import {Link} from "react-router";

import * as classNames from "classnames";
import {Container} from "sourcegraph/Container";
import * as base from "sourcegraph/components/styles/_base.css";
import * as colors from "sourcegraph/components/styles/_colors.css";
import * as typography from "sourcegraph/components/styles/_typography.css";
import * as styles from "sourcegraph/home/styles/home.css";
import {BetaInterestForm} from "sourcegraph/home/BetaInterestForm";
import {LocationStateModal, dismissModal} from "sourcegraph/components/Modal";

import {LocationStateToggleLink} from "sourcegraph/components/LocationStateToggleLink";
import {Button, Heading, Logo, FlexContainer, Panel} from "sourcegraph/components";
import {context} from "sourcegraph/app/context";

interface HomeProps {
	location: any;
}

type HomeState = any;

export class Home extends Container<HomeProps, HomeState> {
	static contextTypes: React.ValidationMap<any> = {
		siteConfig: React.PropTypes.object.isRequired,
	};

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

	reconcileState(state: HomeState, props: HomeProps): void {
		Object.assign(state, props);
	}

	render(): JSX.Element | null {
		return (
			<div style={{width: "100%", marginRight: "auto", marginLeft: "auto"}}>
				{/* section showing icon and links: about, pricing, login, signup */}
				<div className={classNames(base.pt4, base.pb5, colors.bg_cool_mid_gray_1)}>
					<FlexContainer items="center" wrap={true} style={{maxWidth: "960px"}} className={classNames(base.mt2, base.mb5, base.center, base.ph3)}>
						<Logo width="32"/>

						<div className={typography.tr} style={{flex: "1"}} />

						<FlexContainer items="center" justify="between">

							<Link className={classNames(styles.link, base.mr2, base.ph2, base.pv1)} to="/about">
								About
							</Link>

							<Link className={classNames(styles.link, base.mr2, base.ph2, base.pv1)} to="/pricing">
								Pricing
							</Link>

							<a className={classNames(styles.link, base.mr2, base.ph2, base.pv1)} href="/jobs">Jobs</a>

							{!(this.context as any).signedIn &&
								<LocationStateToggleLink className={classNames(styles.link, base.mr2, base.ph2, base.pv1)} href="/login" modalName="login" location={this.props.location}>
									Login
								</LocationStateToggleLink>
							}

							{!(this.context as any).signedIn &&
								<LocationStateToggleLink className={classNames(base.mr0, base.ml2, base.pv1, base.bb, base.bbw2, colors.purple, styles.hover_no_border, styles.hover_dark_purple)} href="/join" modalName="join" location={this.props.location}>
									<strong>Sign up</strong>
								</LocationStateToggleLink>
							}

						</FlexContainer>
					</FlexContainer>

					{/* section showing welcome message and examples */}
					<FlexContainer justify="between" wrap={true} style={{maxWidth: "960px"}} className={classNames(base.center, base.ph3)}>

						{/* column with welcome message, short description, and sign up button */}
						<div style={{maxWidth: "400px", flex: "1 1 400px"}}>
							<Heading align="left" level="2" underline="purple">
								Welcome to the global graph of code
							</Heading>

							<p className={classNames(typography.f5, base.mt0)}>
								<strong>Sourcegraph</strong> answers everyday programming questions in seconds
							</p>

							<p className={base.mt4}>
								<LocationStateToggleLink href="/join" modalName="join" location={this.props.location}>
									<Button type="button" color="purple" className={classNames(base.ph3, base.mr3)}>Sign up for free</Button>
								</LocationStateToggleLink> or
								<Link className={classNames(base.ml3, base.bb, base.bbw2, base.pv2, styles.hover_no_border)} to="/about"><strong>Learn more</strong></Link>
							</p>
						</div>

						<div style={{maxWidth: "400px", maxHeight: "275px", flex: "1 1 400px", position: "relative"}}>
							<LocationStateToggleLink modalName="demo_video" location={this.props.location}>
								<FlexContainer direction="top_bottom" justify="center" style={{position: "absolute", top: "0px", bottom: "0px", right: "0px", left: "0px"}} className={classNames(colors.bg_dark_purple_8, base.br3, typography.tc)}>
									<img src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/play.svg`} />
								</FlexContainer>
								<img src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/video-screenshot.png`} width="100%" height="auto%" className={base.br3} />
							</LocationStateToggleLink>
						</div>

					</FlexContainer>
				</div>

				<div className={classNames(base.center, base.ph3, base.pv5)} style={{maxWidth: "990px"}}>
					{/* section showing questions */}
					<div className={base.center} style={{maxWidth: "600px"}}>
						<Heading align="center" level="4" underline="blue">
							How do I use this function? Who can I ask about this code? <em>What does this code even do?</em>
						</Heading>

						<p className={classNames(typography.tc, base.mt0, base.mb4)} >
							These questions require you to constantly context-switch between your editor, terminal, and browser.
							Sourcegraph can help you stop losing focus and wasting time.
						</p>
					</div>

					{/* section showing feature descriptions */}
					<FlexContainer justify="between" wrap={true} className={base.center}>

						{/* column describing examples */}
						<div style={{maxWidth: "300px", flex: "1 1 300px"}} className={classNames(base.ph3, base.mt4)}>
							<img src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/illo-docs.svg`} width="100%" />

							<Heading level="5" className={base.mt3}>
								Usage examples and instant documentation
							</Heading>

							<p>
								Quickly understand new libraries instead of reinventing the wheel.
							</p>
						</div>

						{/* column describing search */}
						<div style={{maxWidth: "300px", flex: "1 1 300px"}} className={classNames(base.ph3, base.mt4)}>
							<img src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/illo-search.svg`} width="100%" />

							<Heading level="5" className={base.mt3}>
								Global search by function, package, or symbol name
							</Heading>

							<p>
								Find exactly the function you&rsquo;re looking for.
								Search your private code and thousands of open-source repositories.
							</p>
						</div>

						{/* column describing team features */}
						<div style={{maxWidth: "300px", flex: "1 1 300px"}} className={classNames(base.ph3, base.mt4)}>
							<img src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/illo-team.svg`} width="100%" />

							<Heading level="5" className={base.mt3}>
								Designed with teams in mind
							</Heading>

							<p>
								Jump to definition in a code review and instantly see who you should ask about a piece of code.
							</p>
						</div>

					</FlexContainer>
				</div>

				{/* section showing language icons */}
<div className={classNames(base.pv4, typography.tc, colors.bg_cool_mid_gray_1)}>
					<Heading level="7" color="cool_mid_gray" className="base.pv3">
						Growing language support
					</Heading>

					<FlexContainer justify="between" className={classNames(base.center, base.mt4)} style={{maxWidth: "440px"}}>
						<img title="Go supported" className={styles.lang_icon} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/go2.svg`} />
						<img title="Java supported" className={styles.lang_icon} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/java.svg`} />
						<div style={{display: "inline-block", position: "relative", cursor: "pointer"}} onMouseOver={() => this.setState({langMouseover: true})} onMouseLeave={() => this.setState({langMouseover: false})}>
							<img title="JavaScript coming soon" style={{opacity: this.state.langMouseover ? .1 : .3}} className={styles.lang_icon} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/js.svg`} />
							<img title="Python coming soon" style={{opacity: this.state.langMouseover ? .1 : .3}} className={styles.lang_icon} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/python.svg`} />
							<img title="PHP coming soon" style={{opacity: this.state.langMouseover ? .1 : .3}} className={styles.lang_icon} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/php.svg`} />
							<img title="Scala coming soon" style={{opacity: this.state.langMouseover ? .1 : .3}} className={styles.lang_icon} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/scala.svg`} />
							{this.state.langMouseover &&
								<LocationStateToggleLink style={{display: "flex", alignItems: "center", justifyContent: "center", position: "absolute", left: 0, right: 0, top: 0, bottom: 0}} href="/beta" modalName="beta" location={this.props.location}>
									<div>
										<div className={colors.blue} style={{lineHeight: 1}}>
											<strong>Notify me</strong>
										</div>
										<div className={typography.f7}>
											when new languages are supported
										</div>
									</div>
								</LocationStateToggleLink>
							}
						</div>
					</FlexContainer>
				</div>

				{this.props.location.state && this.props.location.state.modal === "beta" &&
					<LocationStateModal modalName="beta" location={this.props.location}>
						<div className={styles.modal}>
							<h2 className={typography.tc}>Join the Sourcegraph beta</h2>
							<BetaInterestForm
								className={styles.modalForm}
								loginReturnTo="/"
								onSubmit={dismissModal("beta", this.props.location, (this.context as any).router)} />
						</div>
					</LocationStateModal>
				}

				<div className={colors.bg_purple} style={{paddingTop: "50px", paddingBottom: "50px"}}>
					<Panel className={base.center} style={{maxWidth: "930px"}}>
						{/* section showing clients */}
						<div className={classNames(base.center, base.pa4)}>
							<FlexContainer wrap={true}>
								<Heading align="left" level="4" underline="purple" style={{flex: "0 0 240px"}} className="full_sm">
									Used by developers everywhere
								</Heading>

								<FlexContainer justify="end" style={{flex: "1 1"}}>
									<img className={base.mr4} style={{marginBottom: "10px"}} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/twitter.svg`} />
									<img className={base.mr4} style={{marginBottom: "9px"}} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/red-hat.svg`} />
									<img className={base.mr4} style={{marginBottom: "7px"}} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/daily-motion.svg`} />
									<img className={base.mr4} style={{marginBottom: "5px"}} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/progressly.svg`} />
								</FlexContainer>
							</FlexContainer>

							{/* section showing favorable user feedback */}
							<FlexContainer justify="between" wrap={true}>
								<div className={classNames(styles.tweet_container, base.pr4, base.mb3)} >
									<blockquote className="twitter-tweet" lang="en"><p lang="en" dir="ltr">Just found out <a href="https://twitter.com/srcgraph">@srcgraph</a> ! HUGE productivity gain. Great work ! Waiting for more language support.</p>&mdash; Dharmesh Kakadia (@dharmeshkakadia) <a href="https://twitter.com/dharmeshkakadia/status/738874411437035520">June 3, 2016</a></blockquote>
								</div>
								<div className={classNames(styles.tweet_container, base.mb3)}>
									<blockquote className="twitter-tweet" lang="en"><p lang="en" dir="ltr">The <a href="https://twitter.com/srcgraph">@srcgraph</a> Chrome extension for GitHub is the best! <a href="https://t.co/CKweAOfbsQ">https://t.co/CKweAOfbsQ</a></p>&mdash; Julius Volz (@juliusvolz) <a href="https://twitter.com/juliusvolz/status/748095329564778496">June 29, 2016</a></blockquote>
								</div>
							</FlexContainer>
							<FlexContainer justify="between" wrap={true}>
								<div className={classNames(styles.tweet_container, base.pr4)}>
									<blockquote className="twitter-tweet" lang="en"><p lang="en" dir="ltr">Used <a href="https://twitter.com/srcgraph">@srcgraph</a> jump-to-definition across 3 projects, 2 langs, finally landing deep in Golang src. Took &lt; 10 min to pin down the issue. 💪🏼</p>&mdash; Gabriel Monroy (@gabrtv) <a href="https://twitter.com/gabrtv/status/738861622882508801">June 3, 2016</a></blockquote>
								</div>
								<div className={styles.tweet_container}>
									<blockquote className="twitter-tweet" lang="en"><p lang="en" dir="ltr">Sourcegraph is the cross reference to end all cross references</p>&mdash; Erik Hollensbe (@erikhollensbe) <a href="https://twitter.com/erikhollensbe/status/738880970909089793">June 3, 2016</a></blockquote>
								</div>
							</FlexContainer>
						</div>
					</Panel>
				</div>

				{/* section showing tagline with a CTA to sign up */}
				<div style={{maxWidth: "660px"}} className={classNames(base.center, base.mv5, base.ph3)}>
					<Heading align="center" level="3">
						Programming should be about building architectures and algorithms, not struggling with how to use a library or function
					</Heading>

					<LocationStateToggleLink href="/join" modalName="join" location={this.props.location} className={classNames(base.mb5, base.mt4, typography.tc)} style={{display: "block"}}>
						<Button type="button" color="purple" className={base.ph4}>Sign up for free</Button>
					</LocationStateToggleLink>
				</div>


			<div className={classNames(base.center, base.mv5, base.ph3, styles.footer)}>
				<a href="/about">About</a>
				<a href="https://text.sourcegraph.com">Blog</a>
				<a href="/jobs">Careers</a>
				<a href="/contact">Contact</a>
				<a href="/pricing">Pricing</a>
				<a href="/-/privacy">Privacy</a>
				<a href="/security">Security</a>
				<a href="/sitemap">Sitemap</a>
				<a href="/-/terms">Terms</a>
			</div>

			</div>
		);
	}
}
