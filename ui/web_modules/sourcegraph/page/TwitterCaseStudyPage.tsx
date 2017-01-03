import * as React from "react";
import { Link } from "react-router";
import { PageTitle } from "sourcegraph/components/PageTitle";

import { context } from "sourcegraph/app/context";
import { Heading, Hero } from "sourcegraph/components";
import { GitHubAuthButton } from "sourcegraph/components/GitHubAuthButton";
import * as base from "sourcegraph/components/styles/_base.css";
import * as styles from "sourcegraph/page/Page.css";
import { Button } from "sourcegraph/components/Button";
import { Footer } from "sourcegraph/app/Footer";

export function TwitterCaseStudyPage(): JSX.Element {
	return (
		<div className={styles.pos_rel}>
			<PageTitle title="Sourcegraph Case Study: Twitter" />
			<Hero className={base.pv5+ ' ' + styles.case_hero}>
				<div className={styles.container}>
					<div className={styles.logos}>
						<div>
							<svg
								width="64px"
								height="64px"
								viewBox="0 0 64 64"
								version="1.1"
								xmlns="http://www.w3.org/2000/svg"
								xmlnsXlink="http://www.w3.org/1999/xlink">
								<g stroke="none" strokeWidth="1" fill="none" fillRule="evenodd">
									<path d="M59.8933333,16 C57.84,16.9333333 55.6266667,17.5466667 53.3333333,17.84 C55.68,16.4266667 57.4933333,14.1866667 58.3466667,11.4933333 C56.1333333,12.8266667 53.68,13.76 51.0933333,14.2933333 C48.9866667,12 46.0266667,10.6666667 42.6666667,10.6666667 C36.4,10.6666667 31.28,15.7866667 31.28,22.1066667 C31.28,23.0133333 31.3866667,23.8933333 31.5733333,24.72 C22.08,24.24 13.6266667,19.68 8,12.7733333 C7.01333333,14.4533333 6.45333333,16.4266667 6.45333333,18.5066667 C6.45333333,22.48 8.45333333,26 11.5466667,28 C9.65333333,28 7.89333333,27.4666667 6.34666667,26.6666667 L6.34666667,26.7466667 C6.34666667,32.2933333 10.2933333,36.9333333 15.52,37.9733333 C14.56,38.24 13.5466667,38.3733333 12.5066667,38.3733333 C11.7866667,38.3733333 11.0666667,38.2933333 10.3733333,38.16 C11.8133333,42.6666667 16,46.0266667 21.04,46.1066667 C17.1466667,49.2 12.2133333,51.0133333 6.82666667,51.0133333 C5.92,51.0133333 5.01333333,50.96 4.10666667,50.8533333 C9.17333333,54.1066667 15.2,56 21.6533333,56 C42.6666667,56 54.2133333,38.56 54.2133333,23.44 C54.2133333,22.9333333 54.2133333,22.4533333 54.1866667,21.9466667 C56.4266667,20.3466667 58.3466667,18.32 59.8933333,16" id="Fill-2" fill="#FFFFFF"></path>
								</g>
							</svg>
						</div>
						<div className={styles.add_icon}>
							<svg
								width="32px"
								height="32px"
								viewBox="96 16 32 32"
								version="1.1"
								xmlns="http://www.w3.org/2000/svg"
								xmlnsXlink="http://www.w3.org/1999/xlink">
								<g id="add-icon" stroke="none" strokeWidth="1" fill="none" fillRule="evenodd" transform="translate(96.000000, 16.000000)">
									<path d="M25,17 L17,17 L17,25 C17,25.552 16.552,26 16,26 C15.448,26 15,25.552 15,25 L15,17 L7,17 C6.448,17 6,16.552 6,16 C6,15.448 6.448,15 7,15 L15,15 L15,7 C15,6.448 15.448,6 16,6 C16.552,6 17,6.448 17,7 L17,15 L25,15 C25.552,15 26,15.448 26,16 C26,16.552 25.552,17 25,17" id="Fill" fill="#FFFFFF"></path>
								</g>
							</svg>
						</div>
						<div>
							<svg
								width="56px"
								height="56px"
								viewBox="160 4 56 56"
								version="1.1"
								xmlns="http://www.w3.org/2000/svg"
								xmlnsXlink="http://www.w3.org/1999/xlink">
								<g id="LogoMark" stroke="none" strokeWidth="1" fill="none" fillRule="evenodd" transform="translate(160.000000, 4.000000)">
									<g id="logomark" fill="#FFFFFF">
										<g id="Page-1">
											<g id="sg-logo">
												<g id="logomark">
													<path d="M15.7476667,7.336 L27.8413333,51.6156667 C28.6556667,54.5906667 31.7216667,56.343 34.692,55.531 C37.6623333,54.7143333 39.41,51.6413333 38.598,48.6663333 L26.502,4.38433333 C25.6876667,1.40933333 22.6216667,-0.343 19.6513333,0.471333333 C16.6833333,1.28566667 14.9333333,4.35866667 15.7476667,7.336" id="Shape"></path>
													<path d="M38.143,7.168 L7.85633333,41.419 C5.81466667,43.729 6.027,47.2616667 8.33233333,49.308 C10.6376667,51.3543333 14.161,51.1396667 16.2026667,48.8296667 L46.4893333,14.5786667 C48.531,12.2686667 48.3186667,8.73833333 46.0133333,6.692 C43.708,4.64566667 40.187,4.858 38.143,7.168" id="Shape"></path>
													<path d="M3.82666667,26.1496667 L47.019,40.4576667 C49.9426667,41.4283333 53.095,39.837 54.0633333,36.9086667 C55.0316667,33.978 53.4426667,30.8186667 50.5213333,29.8526667 L7.32666667,15.5376667 C4.403,14.5716667 1.25066667,16.1583333 0.284666667,19.089 C-0.683666667,22.0196667 0.903,25.179 3.82666667,26.1496667" id="Shape"></path>
												</g>
											</g>
										</g>
									</g>
								</g>
							</svg>
						</div>
					</div>


					<Heading level={1} color="white">Masters of Scale</Heading>
					<Heading level={4} color="white">How investing in fast, semantic code browsing helps Twitter scale engineering productivity</Heading>
				</div>
			</Hero>

			<div className={styles.content}>

				<Heading level={7} align="center">December 2016</Heading>

				<p className={styles.drop_capped}>Few companies have had the impact Twitter has had since it launched in 2006. The social networking service has been described as the “pulse of the planet,” playing a crucial role in just about every culture shift in the last decade. Behind the scenes, Twitter is also an innovator in engineering culture—a fact that becomes even more impressive when you consider that the challenges Twitter faces are formidable.</p>

				<Heading level={6} align="center">Scaling 140 characters to 313 million monthly active users.</Heading>

				<p>Twitter’s codebase is huge and highly complex. It takes a sophisticated engineering organization to build and maintain a product that supports sharing messages, images, video, and more across a global community of 313 million monthly active users (and counting). To scale its products and infrastructure, Twitter uses a combination of languages, including Java, Scala, and others. In many cases, engineering productivity actually goes down with scale as problems with communication and coordination limit the benefits of collaboration. To combat all of this, Twitter has an entire Engineering Effectiveness department focused on investing in people, processes, and tooling to boost the productivity of every Twitter developer.</p>

				<br/>
				<Heading level={4} underline="purple" className={styles.h5}>The Problem:<br/>Understanding and reusing existing code.</Heading>

				<p>Last year, a small team of engineers in Twitter’s Engineering Effectiveness department got together and discussed high-impact ways to improve developer productivity.</p>

				<p>The problem? <strong>Twitter’s codebase was so large and complex that it was hard to understand how each piece of code affected—or was affected by—everything else.</strong> Moreover, the existing internal code browser simply couldn’t handle the scale of Twitter’s codebase. The net result was that navigating the Twitter codebase was slow. And because engineers could not easily answer code-related questions on their own, they often interrupted their teammates with questions, adding to the communication and coordination overhead.</p>

				<p>Building a solution in-house was going to take too long and require too big an investment, especially given all the other infrastructure and product priorities at the time. That’s when the team, led by veteran engineering director David Keenan, started searching for out-of-the-box solutions.</p>

				<p>Sourcegraph met their requirements.</p>

				<br/>
				<Heading level={4} underline="purple" className={styles.h5}>The Solution:<br/>Fast, semantic code browsing.</Heading>
			</div>



				<Hero color="white" className={base.pv5} >
					<div className={styles.container + ' ' + styles.explainer}>
						<p className={styles.case_code}>
							Sourcegraph is a <span className={styles.span_blue}>fast</span>, <span className={styles.span_purple}>semantic code search</span> and <span className={styles.span_orange}>cross-reference engine</span>. It allows users to <strong>search for any function, type, or package and see how other developers use it,</strong> <span className={styles.span_green}>globally</span>. It’s also massively scalable, with 2,000,000,000+ functions in its public code index (and growing).
						</p>
						<img className={styles.search_illustration} src="http://uploads.webflow.com/58522d5bc4d300e777338c69/58523a96b3ee40250f0c94a2_tw-case-search-example.png"></img>
					</div>
				</Hero>

				<div className={styles.content}>

				<p>Under Keenan’s leadership, Twitter’s team brought Sourcegraph in to boost engineering productivity. They chose Sourcegraph because they believed it could become a go-to resource in their internal suite of developer tools.</p>

				<p>During Hack Week, Keenan’s team built Scala support on the Sourcegraph API, and the tool was deployed to all of Twitter engineering within a week. “Sourcegraph is easy to integrate into your internal ecosystem because all it needs is a Git clone URL,” says Keenan.</p>

				<Heading level={6} align="center">“It works even with a completely homegrown repository hosting system like Twitter’s.”</Heading>

				<p>Sourcegraph indexes the main codebase inside of Twitter and helps developers find the answers they need in seconds, not minutes. <strong>It gives them something no IDE can: the ability to easily explore the entire codebase with all its dependencies and discuss code efficiently by linking to specific functions and types.</strong></p>

				</div>

				<div className={styles.repo_illustration} >

					<ul className={styles.scene}>
						<li className={styles.one}>
							<img src="http://uploads.webflow.com/58522d5bc4d300e777338c69/5852fa660fd218da6c4759eb_sg-tw-case-illu-1.png" width="72"></img>
						</li>
						<li className={styles.two}>
							<img src="http://uploads.webflow.com/58522d5bc4d300e777338c69/5852fa66b3ee40250f0d7a82_sg-tw-case-illu-2.png" width="72"></img>
						</li>
						<li className={styles.three}>
							<img src="http://uploads.webflow.com/58522d5bc4d300e777338c69/5852fa66575ab68b5607d904_sg-tw-case-illu-3.png" width="54"></img>
						</li>
						<li className={styles.four}>
							<img src="http://uploads.webflow.com/58522d5bc4d300e777338c69/5852fb2b05b004e37e43473c_sg-tw-case-illu-4.png" width="72"></img>
						</li>
						<li className={styles.zero}>
							<img src="http://uploads.webflow.com/58522d5bc4d300e777338c69/5852fa66e93a3bbc3a5c50fc_sg-tw-case-illu-0.png" width="108"></img>
						</li>
						<li className={styles.five}>
							<img src="http://uploads.webflow.com/58522d5bc4d300e777338c69/5852fa668fdd9c186ddda674_sg-tw-case-illu-5.png" width="72"></img>
						</li>
						<li className={styles.six}>
							<img src="http://uploads.webflow.com/58522d5bc4d300e777338c69/5852fa666516e3f473cf0428_sg-tw-case-illu-6.png" width="72"></img>
						</li>
						<li className={styles.seven}>
							<img src="http://uploads.webflow.com/58522d5bc4d300e777338c69/5852fa67e93a3bbc3a5c50fd_sg-tw-case-illu-7.png" width="54"></img>
						</li>
						<li className={styles.eight}>
							<img src="http://uploads.webflow.com/58522d5bc4d300e777338c69/5852fa67b44310897a8b6324_sg-tw-case-illu-8.png" width="72"></img>
						</li>
					</ul>

				</div>

				<div className={styles.content}>

				<br/>
				<Heading level={4} underline="purple" className={styles.h5}>The Results:<br/>Time saved and limitless potential.</Heading>

				<p>Engineers across many different teams now use Sourcegraph multiple times every week, making it a key part of Twitter’s Engineering Effectiveness toolkit. In the words of one senior engineer:</p>

				<Heading level={6} align="center">“It’s very helpful to have functions viewable and clickable in the browser, so you don’t have to lose your place in your code editor.”</Heading>

				<p>The team saves time in three ways:</p>

				<p>First, Sourcegraph supports jump-to-definition across the entire Twitter repository, making it easier for developers to understand how different parts of the codebase relate to one another.</p>

				<p>Second, hover-over usage examples instantly show how existing code calls a function or uses a type.</p>

				<p>And third, Sourcegraph handles the massive scale of Twitter’s codebase, while remaining fast and efficient.</p>

				<p>Sourcegraph’s reception at Twitter has been overwhelmingly positive across the organization. “Sourcegraph is pretty amazing,” says Keenan. “It makes the Scala code so much easier to navigate. We're looking forward to getting this on Java, too.”</p>

				<p>To learn more about how Sourcegraph can help your engineering team, visit us at sourcegraph.com.</p>

				<div>
				<Link target="_blank" to="https://www.dropbox.com/s/1if8ptxvsmcksm2/sg-tw-case.pdf?dl=0">
					<Button color="blue" className={styles.download_button} onClick={"http://google.com"}>
						<div className={styles.button_icon}>
							<svg 
								width="24px"
								height="24px"
								viewBox="0 0 24 24"
								version="1.1"
								xmlns="http://www.w3.org/2000/svg"
								xmlnsXlink="http://www.w3.org/1999/xlink">
								<g id="Group-2" stroke="none" strokeWidth="1" fill="none" fillRule="evenodd">
									<path d="M9.3816,17.9959875 L11.24985,17.9959875 L11.24985,10.5001125 C11.24985,10.0857375 11.58585,9.7501125 11.99985,9.7501125 C12.414225,9.7501125 12.74985,10.0857375 12.74985,10.5001125 L12.74985,17.9959875 L14.62335,17.9959875 C14.835975,17.9997375 15.0051,18.1752375 15.00135,18.3882375 C14.99985,18.4658625 14.9751,18.5412375 14.930475,18.6042375 L12.3096,22.3448625 C12.192225,22.5143625 11.959725,22.5567375 11.790225,22.4397375 C11.7531,22.4138625 11.72085,22.3816125 11.69535,22.3448625 L9.074475,18.6042375 C8.9511,18.4309875 8.9916,18.1906125 9.165225,18.0672375 C9.2286,18.0218625 9.303975,17.9971125 9.3816,17.9959875 M17.24985,16.5001125 L14.99985,16.5001125 C14.58585,16.5001125 14.24985,16.1641125 14.24985,15.7501125 C14.24985,15.3357375 14.58585,15.0001125 14.99985,15.0001125 L17.2146,15.0001125 C18.86985,15.0076125 20.22285,13.6823625 20.24985,12.0271125 C20.264475,10.0107375 18.70035,9.3169875 17.778975,9.0031125 C17.499975,8.9082375 17.302725,8.6581125 17.2761,8.3644875 L17.237475,7.9587375 C17.069475,6.0353625 15.4821,4.5458625 13.551225,4.5004875 C11.8011,4.4798625 10.8816,5.5043625 10.286475,6.2566125 C10.1076,6.4793625 9.81735,6.5806125 9.538725,6.5176125 C9.217725,6.4257375 8.8851,6.3784875 8.550975,6.3762375 C7.357725,6.4182375 6.40485,7.3842375 6.37935,8.5778625 L6.36735,9.0664875 C6.36135,9.3668625 6.175725,9.6338625 5.8971,9.7444875 C4.924725,10.1251125 3.6891,10.7416125 3.752475,12.5033625 C3.8151,13.9043625 4.972725,15.0061125 6.37485,15.0001125 L8.99985,15.0001125 C9.414225,15.0001125 9.74985,15.3357375 9.74985,15.7501125 C9.74985,16.1641125 9.414225,16.5001125 8.99985,16.5001125 L6.37485,16.5001125 C4.09785,16.5012375 2.250975,14.6558625 2.24985,12.3788625 C2.2491,10.6774875 3.29385,9.1501125 4.879725,8.5339875 C4.9266,6.4666125 6.639975,4.8289875 8.70735,4.8758625 C8.94435,4.8811125 9.1806,4.9092375 9.412725,4.9591125 C11.2281,2.7008625 14.53035,2.3419875 16.78935,4.1577375 C17.892975,5.0453625 18.5916,6.3417375 18.7251,7.7524875 C21.072225,8.5658625 22.31535,11.1278625 21.501975,13.4749875 C20.874225,15.2862375 19.1676,16.5008625 17.24985,16.5001125" id="Fill" fill="#FFFFFF"></path>
								</g>
							</svg>
						</div>
						Download Case Study PDF
					</Button>
					</Link>
				</div>

				{!context.user && <div className={styles.cta}>
					<GitHubAuthButton color="purple">
						<strong>Sign up with GitHub</strong>
					</GitHubAuthButton>
				</div>}

			</div>

			<img className={styles.sprinkle_blue} src="http://uploads.webflow.com/58522d5bc4d300e777338c69/58542ca3860be4da644481bd_sg-tw-case-blue-sprinkle.svg"></img>
			<img className={styles.sprinkle_purple} src="http://uploads.webflow.com/58522d5bc4d300e777338c69/58543a7d7725129c64eccf16_sg-tw-case-purple-sprinkle.svg"></img>
			<img className={styles.sprinkle_orange} src="http://uploads.webflow.com/58522d5bc4d300e777338c69/58543bbd860be4da6444938d_sg-tw-case-orange-sprinkle.svg"></img>

			<Footer />

		</div>
	);
}
