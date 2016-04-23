// @flow weak

import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";
import DashboardPromo from "./DashboardPromo";
import context from "sourcegraph/app/context";
import EventLogger, {EventLocation} from "sourcegraph/util/EventLogger";
import {Button} from "sourcegraph/components";

class DashboardPromos extends React.Component {
	state = {
		selectedItem: 0,
	};

	clickHandler(idx) {
		this.setState({selectedItem: idx});
	}

	dashboardPromoItems() {

		return [
			<DashboardPromo key="GitHub" title="Code browsing & usage examples" subtitle="Browse GitHub code with jump-to-definition links, and see everywhere a function is being used. The right usage example is worth 1,000 words of documentation." onClick={this.clickHandler.bind(this, 0)} isSelected={this.state.selectedItem === 0}/>,
			<DashboardPromo key="Chrome" title="Chrome extension" subtitle="Get Sourcegraph's code intelligence when you're on GitHub.com." onClick={this.clickHandler.bind(this, 1)} isSelected={this.state.selectedItem === 1}/>,
			<span key="cta" styleName="cta-box">
				<a href="join" onClick={() => EventLogger.logEventForPage("JoinCTAClicked", EventLocation.Dashboard, {PageLocation: "Promo"})}>
					<Button color="info" size="large">Get Started</Button>
				</a>
			</span>,
		];
	}

	render() {
		return (
			<div styleName="promo-container">
				<ul styleName="promo-list">
					{this.dashboardPromoItems()}
				</ul>

				{this.state.selectedItem === 0 && <img styleName="promo-img" src={`${context.assetsRoot}/img/Dashboard/Dashboard-Link-GitHub-Promo.svg`}></img>}
				{this.state.selectedItem === 1 && <img styleName="promo-img" src={`${context.assetsRoot}/img/Dashboard/Dashboard-Chrome-Extension-Promo.svg`}></img>}
			</div>
		);
	}
}

export default CSSModules(DashboardPromos, styles);

