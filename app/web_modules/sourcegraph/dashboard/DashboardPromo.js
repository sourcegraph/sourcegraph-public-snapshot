
import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";

class DashboardPromo extends React.Component {

	constructor(props) {
		super(props);
		this.hoverEvent = this.hoverEvent.bind(this);
	}

	state = {
		isSelected: false,
		hover_flag: false,
	};

	hoverEvent() {
		this.setState({hover_flag: !this.state.hover_flag});
	}

	render() {
		let liStyle = {
			background: "#fff",
			color: "#000",
		};

		if (this.props.isSelected) {
			liStyle["background"] = "rgba(0, 116, 217, .1)";
			liStyle["color"] = "rgba(0, 116, 217,1)";
		} else if (this.state.hover_flag) {
			liStyle["background"] = "#eee";
		}

		return (
			<li styleName="promo"
				onClick={this.props.onClick}
				onMouseEnter={this.hoverEvent}
				onMouseLeave={this.hoverEvent}
				style={liStyle}>
					<p styleName="promo-header">{this.props.title}</p>
					<p styleName="promo-subtext">{this.props.subtitle}</p>
			</li>
		);
	}
}

DashboardPromo.propTypes = {
	title: React.PropTypes.string.isRequired,
	subtitle: React.PropTypes.string.isRequired,
	onClick: React.PropTypes.func.isRequired,
	isSelected: React.PropTypes.bool,
};

export default CSSModules(DashboardPromo, styles);
