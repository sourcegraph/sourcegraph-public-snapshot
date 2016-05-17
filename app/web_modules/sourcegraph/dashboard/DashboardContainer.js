import React from "react";
import Helmet from "react-helmet";
import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";
import AnonymousLandingPage from "./AnonymousLandingPage";
import HomeSearchContainer from "sourcegraph/home/HomeSearchContainer";

class DashboardContainer extends React.Component {
	static propTypes = {
		location: React.PropTypes.object.isRequired,
	}

	static contextTypes = {
		signedIn: React.PropTypes.bool.isRequired,
	};

	render() {
		return (
			<div>
				<Helmet title="Home" />
				{!this.context.signedIn && <AnonymousLandingPage location={this.props.location}/>}
				{this.context.signedIn && <div styleName="container">
					<HomeSearchContainer location={this.props.location}/>
				</div>}
			</div>
		);
	}
}

export default CSSModules(DashboardContainer, styles);
