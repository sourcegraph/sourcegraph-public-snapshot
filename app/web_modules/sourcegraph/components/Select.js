// @flow

import React from "react";

import CSSModules from "react-css-modules";
import styles from "./styles/select.css";
import base from "./styles/_base.css";

class Select extends React.Component {
	static propTypes = {
		block: React.PropTypes.bool,
		className: React.PropTypes.string,
		children: React.PropTypes.any,
		label: React.PropTypes.string,
		placeholder: React.PropTypes.string,
		helperText: React.PropTypes.string,
		error: React.PropTypes.bool,
		errorText: React.PropTypes.string,
		style: React.PropTypes.object,
		defaultValue: React.PropTypes.string,
	};

	render() {
		const {style, block, className, placeholder, label, helperText, error, errorText, children, defaultValue} = this.props;
		return (
			<div className={className}>
				{label && <div>{label} <br /></div>}
				<select
					required={true}
					style={style}
					defaultValue={defaultValue}
					styleName={`select ${error ? "border-red" : "border-neutral"}`}
					placeholder={placeholder ? placeholder : ""}>
					{children}
				</select>
			</div>
		);
	}
}

export default CSSModules(Select, styles, {allowMultiple: true});
