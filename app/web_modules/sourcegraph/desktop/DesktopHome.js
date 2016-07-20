// @flow

import React from "react";
import CSSModules from "react-css-modules";
import base from "sourcegraph/components/styles/_base.css";
import layout from "sourcegraph/components/styles/_layout.css";
import colors from "sourcegraph/components/styles/_colors.css";
import typography from "sourcegraph/components/styles/_typography.css";

import {Heading, List} from "sourcegraph/components";
import {Cone} from "sourcegraph/components/symbols";


class DesktopHome extends React.Component {
	render() {
		return (
			<div className={`${layout.containerFixed} ${base.pv5} ${base.ph4}`} style={{maxWidth: "700px"}}>
				<Heading align="center" level="4" underline="blue">
					See live examples, search code, and view inline
					<br className={base["hidden-s"]} />&nbsp;documentation to write better code, faster
				</Heading>

				<div className={base.mv4}>
					<Heading level="5">Go definitions and usages as you code</Heading>
					<p>
						Install one of our editor integrations, and as you write Go code, this pane will update with contextually relevant information.
					</p>
				</div>
				<div className={base.mv4}>
					<Heading level="5">Semantic, global code search</Heading>
					<p>
						Just hit KEY CODES on OS X or click the search box at the top of this page to semantically search for functions and symbols.
					</p>
				</div>
				<div className={base.mv4}>
					<Heading level="5">Powerful search for your private code</Heading>
					<p>
						To enable semantic search and usage examples for your private code, authorize Sourcegraph to access your private repositories.
					</p>
				</div>
				<div className={`${base.mt5} ${typography.f7}`}>
					<Heading level="6">
						<Cone width={16} className={`${colors["fill-orange"]} ${base.mr2}`} style={{
							verticalAlign: "baseline",
							position: "relative",
							top: "1px",
						}} />
						Sourcegraph Desktop is currently in beta
					</Heading>
					<p>
						If the app is not working as expected, see our GitHub to:
					</p>
					<List className={base.mv3}>
						<li><strong><a href="#">Browse troubleshooting tips</a></strong></li>
						<li><strong><a href="#">File an issue</a></strong></li>
					</List>
					<p>
						We love feedback! Shoot us an email at <strong><a href="URL TOEMAIL TEMPLATE">support@sourcegraph.com</a></strong> with ideas on how we can make Sourcegraph Desktop better.
					</p>
				</div>
			</div>
		);
	}
}

export default CSSModules(DesktopHome, {allowMultiple: true});
