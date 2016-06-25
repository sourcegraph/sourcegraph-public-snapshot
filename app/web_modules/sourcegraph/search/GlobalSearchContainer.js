import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/GlobalSearch.css";
import base from "sourcegraph/components/styles/_base.css";
import GlobalSearch from "sourcegraph/search/GlobalSearch";
import {Panel} from "sourcegraph/components";

class GlobalSearchContainer extends React.Component {

	static propTypes = {
		location: React.PropTypes.object.isRequired,
	}

	render() {
		return (
			<div styleName="bg">
				<div styleName="container-fixed" className={base.mt5}>
					<Panel hoverLevel="low" className={`${base.mb4} ${base.pb4} ${base.ph4} ${base.pt3}`}>
						<GlobalSearch query={this.props.location.query.q || ""} location={this.props.location} />
					</Panel>
				</div>
			</div>
		);
	}
}

export default CSSModules(GlobalSearchContainer, styles);
