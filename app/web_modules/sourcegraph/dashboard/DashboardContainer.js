import React from "react";
import Helmet from "react-helmet";
import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";
import {urlToSearch} from "sourcegraph/search/routes";

class DashboardContainer extends React.Component {
	static propTypes = {
		location: React.PropTypes.object.isRequired,
	};

	static contextTypes = {
		signedIn: React.PropTypes.bool.isRequired,
		router: React.PropTypes.object.isRequired,
	};

	componentWillMount() {
		this.context.router.replace(urlToSearch());
	}

	componentWillReceiveProps(nextProps, nextContext) {
		this.context.router.replace(urlToSearch());
	}

	render() {
		return (
			<div styleName="flex-fill">
				<Helmet title="Home" />
			</div>
		);
	}
}

export default CSSModules(DashboardContainer, styles);
