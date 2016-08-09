// tslint:disable

import * as React from "react";

import {Link} from "react-router";
import Logo from "sourcegraph/components/Logo";
import * as styles from "./styles/Footer.css";

class Footer extends React.Component<{}, any> {
	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
	};

	render(): JSX.Element | null {
		return (
			<div id="global-footer" className={styles.footer}>
				<ul className={styles.left_box}>
					<li className={styles.item}>
						<Link to="/" className={styles.link}>
							<Logo width="16px" className={styles.logo_mark} />
						</Link>
					</li>
					<li className={styles.item}><a className={styles.link} href="https://text.sourcegraph.com">Blog</a></li>
					<li className={styles.item}><Link className={styles.link} to="/about">About</Link></li>
					<li className={styles.item}><Link className={styles.link} to="/pricing">Pricing</Link></li>
					<li className={styles.item}><a className={styles.link} href="https://boards.greenhouse.io/sourcegraph" target="_blank">We're hiring</a></li>
				</ul>
				<ul className={styles.right_box}>
					<li className={styles.item}><Link className={styles.link} to="/contact">Contact</Link></li>
					<li className={styles.item}><Link className={styles.link} to="/security">Security</Link></li>
					<li className={styles.item}><Link className={styles.link} to="/-/privacy">Privacy</Link></li>
					<li className={styles.item}><Link className={styles.link} to="/-/terms">Terms</Link></li>
				</ul>
			</div>
		);
	}
}
export default Footer;
