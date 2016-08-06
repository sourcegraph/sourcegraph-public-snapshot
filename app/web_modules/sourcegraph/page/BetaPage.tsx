// tslint:disable

import * as React from "react";
import {Hero, Heading} from "sourcegraph/components/index";
import * as styles from "./Page.css";
import * as base from "sourcegraph/components/styles/_base.css";
import CSSModules from "react-css-modules";
import Helmet from "react-helmet";
import BetaInterestForm from "sourcegraph/home/BetaInterestForm";

function BetaPage(props) {
	return (
		<div>
			<Helmet title="Beta" />
			<Hero pattern="objects" className={base.pv5}>
				<div styleName="container">
					<Heading level="2" color="blue">Get the future Sourcegraph sooner</Heading>
				</div>
			</Hero>
			<div styleName="content">
				<p styleName="p">Sourcegraph is all about keeping you <em>in flow</em> while you code, no matter what tools or languages you use. By joining the Sourcegraph beta program, you can help us build Sourcegraph for your preferred environment&mdash;and help shape the future of the product.</p>

				<Heading level="3" underline="blue" className={styles.h5}>Sourcegraph beta program</Heading>
				<p styleName="p">As a Sourcegraph beta participant, you'll get early access to future releases, including:</p>
				<ul>
					<li styleName="p">Support for more programming languages</li>
					<li styleName="p">More editor integrations</li>
					<li styleName="p">Browser extensions for Firefox, Safari, Internet Explorer, etc.</li>
				</ul>
				<p styleName="p">Here's how it works:</p>
				<ul>
					<li styleName="p">Fill out the form below to join. We'll be in touch when we have something ready for you.</li>
					<li styleName="p">Please don't write publicly about unreleased features.</li>
					<li styleName="p"><a href="https://github.com/sourcegraph/sourcegraph/blob/master/CONTRIBUTING.md#reporting-bugs-and-creating-issues" target="_blank">Report bugs</a> that you encounter.</li>
					<li styleName="p">Share feedback with us and help shape the future of Sourcegraph.</li>
				</ul>
				<br/>
				<Heading level="3" underline="blue">Register for beta access</Heading>

				<BetaInterestForm loginReturnTo="/beta" />
			</div>
		</div>
	);
}
(BetaPage as any).contextTypes = {
	signedIn: React.PropTypes.bool,
};

export default CSSModules(BetaPage, styles);
