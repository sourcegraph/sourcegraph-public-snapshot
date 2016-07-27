// @flow

import * as React from "react";
import {Hero, Heading} from "sourcegraph/components";
import styles from "./Page.css";
import base from "sourcegraph/components/styles/_base.css";
import CSSModules from "react-css-modules";
import Helmet from "react-helmet";

function ContactPage(props, {signedIn}): React$Element<any> {
	return (
		<div>
			<Helmet title="Contact" />
			<Hero pattern="objects" color="dark" className={base.pv1}>
				<div styleName="container">
					<Heading level="3" color="white">Contact Sourcegraph</Heading>
				</div>
			</Hero>
			<div styleName="content">
				<p styleName="p">Sourcegraph is how developers discover and understand code. Need to reach a real human being on our team?</p>

				<Heading level="4" underline="blue" className={styles.h5}>Help using Sourcegraph</Heading>
				<p styleName="p">Send an email to <a href="mailto:support@sourcegraph.com">support@sourcegraph.com</a> with as much information as possible.</p>
				<ul>
					<li>List the steps you took, what you expected, and what you saw.</li>
					<li>Include links and screenshots, if relevant.</li>
				</ul>
				<p styleName="p">Reporting a security vulnerability? See our <a href="/security">responsible disclosure policy</a>.</p>

				<Heading level="4" underline="blue" className={styles.h5}>Around the web</Heading>
				<p styleName="p">Find us elsewhere:</p>
				<ul>
					<li>Twitter: <a href="https://twitter.com/srcgraph" target="_blank">@srcgraph</a></li>
					<li>GitHub: <a href="https://github.com/sourcegraph" target="_blank">github.com/sourcegraph</a></li>
					<li>Facebook: <a href="https://facebook.com/sourcegraph" target="_blank">Sourcegraph Facebook page</a></li>
					<li>YouTube: <a href="https://www.youtube.com/channel/UCOy2N25-AHqE43XupT9mwZQ/videos" target="_blank">Sourcegraph YouTube channel</a></li>
				</ul>

				<Heading level="4" underline="blue" className={styles.h5}>Other inquiries</Heading>
				<p styleName="p">For anything else related to Sourcegraph, contact us at <a href="mailto:hi@sourcegraph.com">hi@sourcegraph.com</a>.</p>

				<Heading level="4" underline="blue" className={styles.h5}>In the real world</Heading>
				<p styleName="p">Sourcegraph<br/>121 2nd St, Suite 200<br/>San Francisco CA, 94105, USA</p>
				<p styleName="p"><a href="https://www.google.com/maps/place/Sourcegraph/@37.7878302,-122.4013944,17z/data=!3m1!4b1!4m5!3m4!1s0x80858062cd4c9f97:0xf3a9d5164f1d61ec!8m2!3d37.7878302!4d-122.3992004" target="_blank">View on Google Maps</a></p>
			</div>
		</div>
	);
}

export default CSSModules(ContactPage, styles);
