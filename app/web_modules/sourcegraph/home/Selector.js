import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/InterestForm.css";

class Selector extends React.Component {

	static propTypes = {
		valueArray: React.PropTypes.array.isRequired,
		title: React.PropTypes.string.isRequired,
	};

	render() {
		let options = [];
		options.push(<option key="none" value="" disabled="true">{this.props.title}</option>);
		for (let [key, value] of this.props.valueArray.entries()) {
			options.push(<option key={key}>{value}</option>);
		}
		return (
			<select styleName="input-select" required={true} defaultValue="">
				{options}
			</select>);
	}
}

export default CSSModules(Selector, styles);
