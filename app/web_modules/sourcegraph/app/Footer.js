import React from "react";

import context from "sourcegraph/app/context";
import {Link} from "react-router";
import CSSModules from "react-css-modules";
import styles from "./styles/Footer.css";

class Footer extends React.Component {
	render() {
		return (
			<div styleName="footer">
				<ul styleName="left-box">
					<li styleName="item">
						<Link to="/" styleName="link">
							<img styleName="logo-mark" title="Sourcegraph" alt="Sourcegraph logo" src={`${context.assetsRoot}/img/sourcegraph-mark.svg`} />
						</Link>
					</li>
					<li styleName="item"><a styleName="link" href="/blog/">Blog</a></li>
					<li styleName="item"><a styleName="link" href="/about/">About</a></li>
					<li styleName="item"><a styleName="link" href="/careers/">Careers</a></li>
				</ul>
				<ul styleName="right-box">
					<li styleName="item"><a styleName="link" href="mailto:hi@sourcegraph.com">Contact</a></li>
					<li styleName="item"><a styleName="link" href="/security/">Security</a></li>
					<li styleName="item"><a styleName="link" href="/privacy/">Privacy</a></li>
					<li styleName="item"><a styleName="link" href="/legal/">Terms</a></li>
				</ul>
			</div>
		);
	}
}
export default CSSModules(Footer, styles);
