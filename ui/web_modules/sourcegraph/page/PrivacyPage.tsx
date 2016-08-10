// tslint:disable: typedef ordered-imports

import * as React from "react";
import {Hero, Heading} from "sourcegraph/components/index";
import * as styles from "./Page.css";
import * as base from "sourcegraph/components/styles/_base.css";
import Helmet from "react-helmet";

export function PrivacyPage(props: {}, {}) {
	return (
		<div>
			<Helmet title="Privacy" />
			<Hero pattern="objects" color="dark" className={base.pv1}>
				<div className={styles.container}>
					<Heading level="3" color="white">Sourcegraph Privacy Policy</Heading>
				</div>
			</Hero>
			<div className={styles.content}>
				<p>Sourcegraph, Inc. (“<b>Sourcegraph</b>,” “<b>we</b>,” “<b>our</b>,” or “<b>us</b>”) understands that privacy is important to our online visitors to our website and users of our online services (collectively, for the purposes of this Privacy Policy, our “<b>Service</b>”). This Privacy Policy explains how we collect, use, share and protect your personal information that we collect through our Service. By using our Service, you agree to the terms of this Privacy Policy and our <a href="https://sourcegraph.com/-/terms">Terms of Service</a>.</p>

				<p>Capitalized terms that are not defined in this Privacy Policy have the meaning given them in our <a href="https://sourcegraph.com/-/terms">Terms of Service</a>.</p>

				<h4>WHAT INFORMATION DO WE COLLECT AND FOR WHAT PURPOSE</h4>

				<p>The categories of information we collect can include:</p>
				<ul>
					<li>
						<b>Information you provide to us directly:</b> We may collect personal information such as a username, first and last name, mailing address, phone number, email address and payment information, when you register for a Sourcegraph account, participate in forums, comment on blogposts, or if you correspond with us.
					</li>
					<li>
						<b>Information we collect from third parties and social media sites:</b> We may collect information about you from third party services. For example, we may receive information about you if you interact with our site through various social media, for example, by logging in through or liking us on Facebook or following us on Twitter. The data we receive is dependent upon your privacy settings with the social network. You should always review, and if necessary, adjust your privacy settings on third-party websites and services before linking or connecting them to the Service.
					</li>
				</ul>

				<p>We use this information to operate, maintain, and provide to you the features of the Service. We may use this information to communicate with you, such as to send you email messages, and to follow up with you to offer news and information about our Service. We may also send you Service-related emails or messages (e.g., account verification, change or updates to features of the Service, technical and security notices). For more information about your communication preferences, see “Your Choices About Your Information” below.</p>

				<h4>HOW WE USE COOKIES AND OTHER TRACKING TECHNOLOGY TO COLLECT INFORMATION</h4>

				<p>We and our third party partners may automatically collect certain types of usage information when you visit our website or use our Service . For instance, when you visit our websites, we may send one or more cookies — a small text file containing a string of alphanumeric characters — to your computer that uniquely identifies your browser and lets us help you log in faster and enhance your navigation through the site. A cookie may also convey information to us about how you use the Service (e.g., the pages you view, the links you click, how frequently you access the Service, and other actions you take on the Service), and allow us to track your usage of the Service over time. We may collect log file information about your browser or mobile device each time you access the Service. Log file information may include anonymous information such as your web request, Internet Protocol (“IP”) address, browser type, information about your mobile device, referring / exit pages and URLs, number of clicks and how you interact with links on the Service, domain names, landing pages, pages viewed, and other such information. We may employ clear gifs (also known as web beacons) which are used to anonymously track the online usage patterns of our Users. In addition, we may also use clear gifs in HTML-based emails sent to our users to track which emails are opened and which links are clicked by recipients. The information allows for more accurate reporting and improvement of the Service. We may also collect analytics data, or use third-party analytics tools, to help us measure traffic and usage trends for the Service. These tools collect information sent by your browser or mobile device, including the pages you visit, your use of third party applications, and other information that assists us in analyzing and improving the Service. Although we do our best to honor the privacy preferences of our Users, we are not able to respond to Do Not Track signals from your browser at this time.</p>

				<p>When you access our Service by or through a mobile device, we may receive or collect and store a unique identification number associated with your device (UDID, IDFA or similar), mobile carrier, device type and manufacturer, and phone number.</p>

				<p>We may use the data collected through cookies, log file, device identifiers, location data and clear gifs information to: (a) remember information so that you will not have to re-enter it during your visit or the next time you visit the site; (b) provide custom, personalized content and information, including advertising; (c) provide and monitor the effectiveness of our Service; (d) monitor aggregate metrics such as total number of visitors, traffic, usage, and demographic patterns on our website and our Service; (e) diagnose or fix technology problems; and (f) otherwise to plan for and enhance our service.</p>

				<p><b style={{textDecoration: "underline"}}>Third Party Tracking and Online Advertising:</b> We may permit third party ad networks, social media companies, and other third party services to collect information about browsing behavior from visitors to our Service through cookies, social plug-ins, or other tracking technology. We may permit third party online advertising networks to collect information about your use of our Services over time so that they may play or display ads that may be relevant to your interests on our website as well as on other websites or services. Typically, the information is collected through cookies or similar tracking technologies. You may be able to “opt out” of the collection of information through cookies or other tracking technology by actively managing the settings on your browser or mobile device. Please refer to your browser’s or mobile device’s technical information for instructions on how to delete and disable cookies, and other tracking/recording tools. (To learn more about cookies, clear gifs/web beacons and related technologies and how you may opt-out of some of this tracking, you may wish to visit <a href="http://www.allaboutcookies.org">http://www.allaboutcookies.org</a> and/or the Network Advertising Initiative’s online resources, at <a href="http://www.networkadvertising.org">http://www.networkadvertising.org</a>). Depending on your mobile device, you may not be able to control tracking technologies through settings. If you have any questions about opting out of the collection of cookies and other tracking/recording tools, you can contact us directly at <a href="mailto:support@sourcegraph.com">support@sourcegraph.com</a>.</p>

				<h4>SHARING YOUR INFORMATION</h4>

				<p>We may share your personal information in the instances described below. For further information on your choices regarding your information, see the “Your Choices About Your Information” section below.</p>

				<p>We may share your personal information with:</p>
				<ul>
					<li>
						Other companies owned by or under common ownership with Sourcegraph. These companies will use your personal information in the same way as we can under this Privacy Policy;
					</li>
					<li>
						Third-party vendors and other service providers that perform services on our behalf, as needed to carry out their work for us, which may include identifying and serving targeted advertisements, billing, payment processing, or providing analytic services;
					</li>
					<li>
						Our business partners who offer a service to you jointly with us, or who partner with us to provide services to you;
					</li>
					<li>
						Other companies whose products or services may be of interest to you (to see how you may opt-out of this sharing, see Your Choices, below);
					</li>
					<li>
						Other parties in connection with a company transaction, such as a merger, sale of company assets or shares, reorganization, financing, change of control or acquisition of all or a portion of our business by another company or third party or in the event of a bankruptcy or related or similar proceedings; and
					</li>
					<li>
						Third parties as required by law or subpoena or to if we reasonably believe that such action is necessary to (a) comply with the law and the reasonable requests of law enforcement; (b) to enforce our <a href="https://sourcegraph.com/-/terms">Terms of Service</a> or to protect the security or integrity of our Service; and/or (c) to exercise or protect the rights, property, or personal safety of Sourcegraph, our Users, or others.
					</li>
				</ul>

				<p>We may also aggregate or otherwise strip data of all personally identifying characteristics and may share that aggregated or anonymized data with third parties.</p>

				<p>Any information or content that you voluntarily disclose for posting to the Service becomes available to the public, as controlled by any applicable privacy settings. If you remove information or content that you posted to the Service, copies may remain viewable in cached and archived pages of the Service, or if other Users have copied or saved that information.</p>

				<h4>YOUR CHOICES ABOUT YOUR INFORMATION</h4>

				<p><b>How to control your communications preferences:</b> You can stop receiving promotional email communications from us by clicking on the “unsubscribe link” provided in such communications. We make every effort to promptly process all unsubscribe requests. You may not opt out of Service-related communications (e.g., account verification, changes/updates to our products or features of the Service, technical and security notices). If you want to opt-out of having your information shared with third parties for marketing purposes, [you may contact us directly at <a href="mailto:support@sourcegraph.com">support@sourcegraph.com</a>.</p>

				<p><b>Modifying or deleting your information:</b> If you have any questions about reviewing, modifying or deleting your account information, you can contact us directly at <a href="mailto:support@sourcegraph.com">support@sourcegraph.com</a>.</p>

				<h4>HOW WE STORE AND PROTECT YOUR INFORMATION</h4>

				<p><b>Storage and processing:</b> Your information collected through the Service may be stored and processed in the United States or any other country in which Sourcegraph or its subsidiaries, affiliates or service providers maintain facilities. If you are located in the European Union or other regions with laws governing data collection and use that may differ from U.S. law, please note that we may transfer information, including personal information, to a country and jurisdiction that does not have the same data protection laws as your jurisdiction, and you consent to the transfer of information to the U.S. or any other country in which Sourcegraph or its parent, subsidiaries, affiliates or service providers maintain facilities and the use and disclosure of information about you as described in this Privacy Policy.</p>

				<p><b>Keeping your information safe:</b> Sourcegraph cares about the security of your information, and uses commercially reasonable physical, administrative, and technological safeguards to preserve the integrity and security of all information collected through the Service. However, no security system is impenetrable and we cannot guarantee the security of our systems 100%. In the event that any information under our control is compromised as a result of a breach of security, Sourcegraph will take reasonable steps to investigate the situation and where appropriate, notify those individuals whose information may have been compromised and take other steps, in accordance with any applicable laws and regulations.</p>

				<h4>USER DATA</h4>

				<p>Sourcegraph’s collection, storage, and use of User Data, and our responsibilities with respect to such User Data, are set forth in our <a href="https://sourcegraph.com/-/terms">Terms of Service</a> and are not covered by this Privacy Policy.</p>

				<h4>CHILDREN’S PRIVACY</h4>

				<p>Sourcegraph does not knowingly collect or solicit any information from anyone under the age of 13 or knowingly allow such persons to register as Users. In the event that we learn that we have collected personal information from a child under age 13, we will delete that information as quickly as possible. If you believe that we might have any information from a child under 13, please contact us at <a href="mailto:support@sourcegraph.com">support@sourcegraph.com</a>.</p>

				<h4>LINKS TO OTHER WEB SITES AND SERVICES</h4>

				<p>Our Service may integrate with or contain links to other third party sites and services. We are not responsible for the practices employed by third party websites or services embedded in, linked to, or linked from the Service and your interactions with any third-party website or service are subject to that third party’s own rules and policies.</p>

				<h4>HOW TO CONTACT US</h4>

				<p>If you have any questions about this Privacy Policy or the Service, please contact us at <a href="mailto:support@sourcegraph.com">support@sourcegraph.com</a>.</p>

				<h4>CHANGES TO OUR PRIVACY POLICY</h4>

				<p>Sourcegraph may modify or update this Privacy Policy from time to time to reflect the changes in our business and practices. When we do so, we will revise the ‘last updated’ date at the top of this page.</p>
			</div>
		</div>
	);
}
