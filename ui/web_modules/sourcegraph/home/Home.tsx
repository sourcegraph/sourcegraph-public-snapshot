// tslint:disable

import * as React from "react";
import {Link} from "react-router";

import * as classNames from "classnames";
import {Container} from "sourcegraph/Container";
import * as base from "sourcegraph/components/styles/_base.css";
import * as colors from "sourcegraph/components/styles/_colors.css";
import * as typography from "sourcegraph/components/styles/_typography.css";
import * as styles from "./styles/home.css";

import {LocationStateToggleLink} from "sourcegraph/components/LocationStateToggleLink";
import {LocationStateModal} from "sourcegraph/components/Modal";
import {Button, Heading, Logo, FlexContainer} from "sourcegraph/components/index";

type HomeProps = {
	location: Object,
};

export class Home extends Container<HomeProps, any> {

	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
		signedIn: React.PropTypes.bool.isRequired,
	};

	constructor(props) {
		super(props);
	}

	componentDidMount() {
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
					<div style={{display: "flex", flowDirection: "row", alignItems: "center", maxWidth: "960px"}} className={classNames(base.mt2, base.mb4, base.center, base.ph3)}>
						<Logo width="32"/>

						<div className={typography.tr} style={{flex: "1"}} />

						<div style={{display: "flex", flowDirection: "row", justifyContent: "space-between", alignItems: "center",
									flexWrap: "wrap"}}>

							<p className={base.mr3} style={{margin: "10px"}}>
								<Link className={styles.link} to="/about">About</Link>
							</p>

							<p className={base.mr3} style={{margin: "10px"}}>
								<Link className={styles.link} to="/pricing">Pricing</Link>
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
					<div style={{display: "flex", flowDirection: "row", maxWidth: "960px"}} className={classNames(base.center, base.ph3)} styleName="wrap_sm">

						{/* column with welcome message, short description, and sign up button */}
						<div style={{flex: "1 0 340px", maxWidth: "400px"}}>
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
						<div style={{flex: "0 1 32px"}} className={base.hidden_s}></div>
						<div style={{flex: "1", position: "relative", lineHeight: "0"}} styleName="full_sm">
							<LocationStateToggleLink modalName="demo_video" location={this.props.location} styleName="video">
								<FlexContainer direction="top_bottom" justify="center" style={{position: "absolute", top: "0px", bottom: "0px", right: "0px", left: "0px"}} className={classNames(colors.bg_dark_purple_8, base.br3, typography.tc)} styleName="video_overlay">
									<img src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/play.svg`} />
								</FlexContainer>
								<img src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/video-screenshot.png`} width="100%" height="auto%" className={base.br3} />
							</LocationStateToggleLink>
						</div>

					</div>
				</div>

				<div className={base.center} style={{maxWidth: "960px", paddingTop: "100px", paddingBottom: "100px"}}>
					{/* section showing questions */}
					<div style={{maxWidth: "600px", marginRight: "auto", marginLeft: "auto"}}>
						<Heading className="hook-title" align="center" level="4" underline="blue">
							How do I use this function? Who can I ask about this code? <em>What does this code even do?</em>
						</Heading>

						<p className={classNames(typography.tc, base.mt0, base.mb5)} >
							These questions require you to constantly context-switch between your editor, terminal, and browser.
							Sourcegraph can help you stop losing focus and wasting hours.
						</p>
					</div>

					{/* section showing feature descriptions */}
					<div style={{display: "flex", flowDirection: "row", justifyContent: "space-between",
								marginRight: "auto", marginLeft: "auto", flexWrap: "wrap"}}>

						{/* column describing examples */}
						<div style={{flex: "1 1 33%"}} className={base.ph3}>
							<img src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/illo-docs.svg`} width="100%" />

							<p style={{fontWeight: "bold"}}>
								Usage examples and instant documentation
							</p>

							<p>
								Quickly understand new libraries instead of reinventing the wheel.
							</p>
						</div>

						{/* column describing search */}
						<div style={{flex: "1 1 33%"}} className={base.ph3}>
							<img src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/illo-search.svg`} width="100%" />

							<p style={{fontWeight: "bold"}}>
								Global search by function, package, or symbol name
							</p>

							<p>
								Find exactly the function you're looking for.
								Search your private code and thousands of open source repositories.
							</p>
						</div>

						{/* column describing team features */}
						<div style={{flex: "1 1 33%"}} className={base.ph3}>
							<img src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/illo-team.svg`} width="100%" />

							<p style={{fontWeight: "bold"}}>
								Designed with teams in mind
							</p>

							<p>
								Jump to definition in a code review and instantly see who you should ask about a piece of code.
							</p>
						</div>

					</div>
				</div>

				{/* section showing language icons */}
				<div style={{backgroundColor: "rgba(119, 147, 174, 0.1)", display: "flex", flowDirection: "row", justifyContent: "center",
							paddingLeft: "200px", paddingRight: "200px", paddingTop: "10px", paddingBottom: "10px"}}>
					<div style={{maxWidth: "400px", display: "flex", flowDirection: "row", flexGrow: 1, justifyContent: "space-between"}}>
						<img width="32px" src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/go2.svg`} />
						<img width="32px" src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/java.svg`} />
						<img width="32px" src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/js.svg`} />
						<img width="32px" src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/python.svg`} />
						{/*
							<img width="32px" src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/php.svg`} />
							<img width="32px" src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/scala.svg`} />
						*/}
					</div>
				</div>

				<div className={colors.bg_purple} style={{paddingTop: "50px", paddingBottom: "50px"}}>
					<div className={classNames(base.pa5, base.center)} style={{backgroundColor: "white",
								boxShadow: "0 2px 6px 0 rgba(0, 0, 0, 0.11)", maxWidth: "860px", width: "100%"}}>
						{/* section showing clients */}
						<div style={{display: "flex", flowDirection: "row", flexWrap: "wrap"}}>
							<Heading align="left" level="4" underline="purple" style={{flex: "0 0 240px"}} className="full_sm">
								Used by developers everywhere
							</Heading>

							<div style={{flex: "1 1", display: "flex", flexWrap: "wrap", justifyContent: "flex-end"}} styleName="left_sm">
								<img className={base.mr4} style={{marginBottom: "10px"}} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/twitter.svg`} />
								<img className={base.mr4} style={{marginBottom: "9px"}} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/red-hat.svg`} />
								<img className={base.mr4} style={{marginBottom: "10px"}} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/daily-motion.svg`} />
								<img className={classNames(base.mr4, base.hidden_s)} style={{marginBottom: "10px"}} src={`${(this.context as any).siteConfig.assetsRoot}/img/Homepage/logo/progressly.svg`} />
							</div>
						</div>

						{/* section showing favorable user feedback */}
						<div>
							<table>
								<tbody>
									<tr>
										<td style={{width: "400px", verticalAlign: "top"}}>
											<blockquote className="twitter-tweet" dataLang="en"><p lang="en" dir="ltr">Just found out <a href="https://twitter.com/srcgraph">@srcgraph</a> ! HUGE productivity gain. Great work ! Waiting for more language support.</p>&mdash; Dharmesh Kakadia (@dharmeshkakadia) <a href="https://twitter.com/dharmeshkakadia/status/738874411437035520">June 3, 2016</a></blockquote>
										</td>
										<td style={{width: "400px", verticalAlign: "top"}}>
											<blockquote className="twitter-tweet" dataLang="en"><p lang="en" dir="ltr">The <a href="https://twitter.com/srcgraph">@srcgraph</a> Chrome extension for GitHub is the best! <a href="https://t.co/CKweAOfbsQ">https://t.co/CKweAOfbsQ</a></p>&mdash; Julius Volz (@juliusvolz) <a href="https://twitter.com/juliusvolz/status/748095329564778496">June 29, 2016</a></blockquote>
										</td>
									</tr>
									<tr>
										<td style={{width: "400px", verticalAlign: "top"}}>
											<blockquote className="twitter-tweet" dataLang="en"><p lang="en" dir="ltr">Used <a href="https://twitter.com/srcgraph">@srcgraph</a> jump-to-definition across 3 projects, 2 langs, finally landing deep in Golang src. Took &lt; 10 min to pin down the issue. 💪🏼</p>&mdash; Gabriel Monroy (@gabrtv) <a href="https://twitter.com/gabrtv/status/738861622882508801">June 3, 2016</a></blockquote>
										</td>
										<td style={{width: "400px", verticalAlign: "top"}}>
											<blockquote className="twitter-tweet" dataLang="en"><p lang="en" dir="ltr">Sourcegraph is the cross reference to end all cross references</p>&mdash; Erik Hollensbe (@erikhollensbe) <a href="https://twitter.com/erikhollensbe/status/738880970909089793">June 3, 2016</a></blockquote>
										</td>
									</tr>
								</tbody>
							</table>
						</div>
					</div>
				</div>

				{/* section showing tagline with a CTA to sign up */}
				<div style={{maxWidth: "660px"}} className={classNames(base.center, base.mv5, base.ph3)}>
					<Heading align="center" level="3">
						Programming should be about algorithms and architectures, not searching for docs and usage examples
					</Heading>

					<LocationStateToggleLink href="/join" modalName="join" location={this.props.location} className={classNames(base.mb5, base.mt4, typography.tc)} style={{display: "block"}}>
						<Button type="button" color="purple" className={base.ph4}>Sign up for free</Button>
					</LocationStateToggleLink>
				</div>

			</div>
		);
	}
}
