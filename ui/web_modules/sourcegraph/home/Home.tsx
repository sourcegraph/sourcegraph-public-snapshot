// tslint:disable

import * as React from "react";
import {Link} from "react-router";

import * as classNames from "classnames";
import {Container} from "sourcegraph/Container";
import * as base from "sourcegraph/components/styles/_base.css";
import * as colors from "sourcegraph/components/styles/_colors.css";
import * as typography from "sourcegraph/components/styles/_typography.css";
import * as styles from "sourcegraph/home/styles/home.css";

import {LocationStateToggleLink} from "sourcegraph/components/LocationStateToggleLink";
import {LocationStateModal} from "sourcegraph/components/Modal";
import {Button, Heading, Logo, FlexContainer} from "sourcegraph/components/index";

interface HomeProps {
	location: Object;
}

type HomeState = any;

export class Home extends Container<HomeProps, HomeState> {
	static contextTypes: React.ValidationMap<any> = {
		siteConfig: React.PropTypes.object.isRequired,
		signedIn: React.PropTypes.bool.isRequired,
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

	render(): JSX.Element | null {
		return (
			<div style={{width: "100%", marginRight: "auto", marginLeft: "auto"}}>
				{/* section showing icon and links: about, pricing, login, signup */}
				<div style={{paddingTop: "25px", paddingBottom: "110px", backgroundColor: "rgba(119, 147, 174, 0.1)"}}>
					<div style={{display: "flex", flexDirection: "row", alignItems: "center", maxWidth: "960px"}} className={classNames(base.mt2, base.mb4, base.center, base.ph3)}>
						<Logo width="32"/>

						<div className={typography.tr} style={{flex: "1"}} />

						<div style={{display: "flex", flexDirection: "row", justifyContent: "space-between", alignItems: "center",
									flexWrap: "wrap"}}>

							<p className={base.mr3} style={{margin: "10px"}}>
								<Link className={styles.link} to="/about">About</Link>
							</p>

							<p className={base.mr3} style={{margin: "10px"}}>
								<Link className={styles.link} to="/pricing">Pricing</Link>
							</p>

							<p className={base.mr3} style={{margin: "10px"}}>
								<a className={styles.link} href="/jobs">Jobs</a>
							</p>

							{!(this.context as any).signedIn &&
								<LocationStateToggleLink className={styles.link} href="/login" modalName="login" location={this.props.location} style={{margin: "10px"}}>
									Login
								</LocationStateToggleLink>
							}

							{!(this.context as any).signedIn &&
								<LocationStateToggleLink className={styles.link} href="/join" modalName="join" location={this.props.location} style={{margin: "10px"}}>
									Sign up
								</LocationStateToggleLink>
							}

						</div>
					</div>

					{/* section showing welcome message and examples */}
					<div style={{display: "flex", flexDirection: "row", maxWidth: "960px", flexWrap: "wrap", justifyContent: "space-around", alignItems: "center"}} className={classNames(base.center, base.ph3)}>

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
								</LocationStateToggleLink> or <Link className={styles.link} to="/about">Learn more</Link>
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

					</div>
				</div>

				<div className={classNames(base.center, base.ph3)} style={{maxWidth: "960px", paddingTop: "100px", paddingBottom: "100px"}}>
					{/* section showing questions */}
					<div style={{maxWidth: "600px", marginRight: "auto", marginLeft: "auto"}}>
						<Heading className="hook-title" align="center" level="4" underline="blue">
							How do I use this function? Who can I ask about this code? <em>What does this code even do?</em>
						</Heading>

						<p className={classNames(typography.tc, base.mt0, base.mb5)} >
							These questions require you to constantly context-switch between your editor, terminal, and browser.
							Sourcegraph can help you stop losing focus and wasting time.
						</p>
					</div>

					{/* section showing feature descriptions */}
					<FlexContainer justify="around" wrap={true} className={base.center}>

						{/* column describing examples */}
						<div style={{maxWidth: "300px", flex: "1 1 300px"}} className={base.ph3}>
							<img src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/illo-docs.svg`} width="100%" />

							<p style={{fontWeight: "bold"}}>
								Usage examples and instant documentation
							</p>

							<p>
								Quickly understand new libraries instead of reinventing the wheel.
							</p>
						</div>

						{/* column describing search */}
						<div style={{maxWidth: "300px", flex: "1 1 300px"}} className={base.ph3}>
							<img src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/illo-search.svg`} width="100%" />

							<p style={{fontWeight: "bold"}}>
								Global search by function, package, or symbol name
							</p>

							<p>
								Find exactly the function you're looking for.
								Search your private code and thousands of open-source repositories.
							</p>
						</div>

						{/* column describing team features */}
						<div style={{maxWidth: "300px", flex: "1 1 300px"}} className={base.ph3}>
							<img src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/illo-team.svg`} width="100%" />

							<p style={{fontWeight: "bold"}}>
								Designed with teams in mind
							</p>

							<p>
								Jump to definition in a code review and instantly see who you should ask about a piece of code.
							</p>
						</div>

					</FlexContainer>
				</div>

				{/* section showing language icons */}
				<div style={{backgroundColor: "rgba(119, 147, 174, 0.1)", display: "flex", flexDirection: "column", justifyContent: "center",
							alignItems: "center", padding: "40px"}}>
					<Heading level="7" color="cool_mid_gray">
						Growing language support
					</Heading>
					<div style={{maxWidth: "400px", display: "flex", flexDirection: "row", flexGrow: 1, justifyContent: "space-between"}}>
						<img title="Go supported" className={styles.lang_icon} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/go2.svg`} />
						<img title="Java supported" className={styles.lang_icon} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/java.svg`} />
						<img title="JavaScript coming soon" className={styles.lang_icon} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/js.svg`} />
						<img title="Python coming soon" className={styles.lang_icon} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/python.svg`} />
						{/*
							<img style={{width: "32px", padding: "10px"}} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/php.svg`} />
							<img style={{width: "32px", padding: "10px"}} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/scala.svg`} />
						*/}
					</div>
				</div>

				<div className={colors.bg_purple} style={{paddingTop: "50px", paddingBottom: "50px"}}>
					<div className={base.center} style={{backgroundColor: "white",
						boxShadow: "0 2px 6px 0 rgba(0, 0, 0, 0.11)", maxWidth: "860px"}}>
						{/* section showing clients */}
						<div style={{maxWidth: "760px", padding: "50px 10px"}} className={base.center}>
						<FlexContainer wrap={true}>
							<Heading align="left" level="4" underline="purple" style={{flex: "0 0 240px"}} className="full_sm">
								Used by developers everywhere
							</Heading>

							<div style={{flex: "1 1", display: "flex", flexWrap: "wrap", justifyContent: "flex-end"}}>
								<img className={base.mr4} style={{marginBottom: "10px"}} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/twitter.svg`} />
								<img className={base.mr4} style={{marginBottom: "9px"}} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/red-hat.svg`} />
								<img className={base.mr4} style={{marginBottom: "10px"}} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/daily-motion.svg`} />
								<img className={base.mr4} style={{marginBottom: "10px"}} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/progressly.svg`} />
							</div>
						</FlexContainer>

						{/* section showing favorable user feedback */}
						<FlexContainer justify="around" wrap={true}>
							<div className={styles.tweet_container}>
								<blockquote className="twitter-tweet" lang="en"><p lang="en" dir="ltr">Just found out <a href="https://twitter.com/srcgraph">@srcgraph</a> ! HUGE productivity gain. Great work ! Waiting for more language support.</p>&mdash; Dharmesh Kakadia (@dharmeshkakadia) <a href="https://twitter.com/dharmeshkakadia/status/738874411437035520">June 3, 2016</a></blockquote>
							</div>
							<div className={styles.tweet_container}>
								<blockquote className="twitter-tweet" lang="en"><p lang="en" dir="ltr">The <a href="https://twitter.com/srcgraph">@srcgraph</a> Chrome extension for GitHub is the best! <a href="https://t.co/CKweAOfbsQ">https://t.co/CKweAOfbsQ</a></p>&mdash; Julius Volz (@juliusvolz) <a href="https://twitter.com/juliusvolz/status/748095329564778496">June 29, 2016</a></blockquote>
							</div>
							<div className={styles.tweet_container}>
								<blockquote className="twitter-tweet" lang="en"><p lang="en" dir="ltr">Used <a href="https://twitter.com/srcgraph">@srcgraph</a> jump-to-definition across 3 projects, 2 langs, finally landing deep in Golang src. Took &lt; 10 min to pin down the issue. 💪🏼</p>&mdash; Gabriel Monroy (@gabrtv) <a href="https://twitter.com/gabrtv/status/738861622882508801">June 3, 2016</a></blockquote>
							</div>
							<div className={styles.tweet_container}>
								<blockquote className="twitter-tweet" lang="en"><p lang="en" dir="ltr">Sourcegraph is the cross reference to end all cross references</p>&mdash; Erik Hollensbe (@erikhollensbe) <a href="https://twitter.com/erikhollensbe/status/738880970909089793">June 3, 2016</a></blockquote>
							</div>
						</FlexContainer>
						</div>
					</div>
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

			</div>
		);
	}
}
