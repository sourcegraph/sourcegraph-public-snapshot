// @flow

import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/checkboxList.css";

class CheckboxList extends React.Component {
	static propTypes = {
		title: React.PropTypes.string.isRequired,
		name: React.PropTypes.string.isRequired,
		labels: React.PropTypes.arrayOf(React.PropTypes.string).isRequired,

		defaultValues: React.PropTypes.arrayOf(React.PropTypes.string),
		className: React.PropTypes.string,
	};

	// TODO(slimsag): this should be 'element' type?
	_fieldset: any;

	selected(): string[] {
		let selected = [];
		for (let input of this._fieldset.querySelectorAll("input")) {
			if (input.checked) selected.push(input.value);
		}
		return selected;
	}

	_isDefaultValue(s: string): bool {
		return this.props.defaultValues && this.props.defaultValues.indexOf(s) !== -1;
	}

	render() {
		const {className, title, name, labels} = this.props;
		let checkboxes = [];
		for (let label of labels) {
			checkboxes.push(<span styleName="checkbox" key={label}><label><input type="checkbox" name={name} defaultValue={label} defaultChecked={this._isDefaultValue(label)} /> {label}</label></span>);
		}

		return (
			<fieldset ref={(c) => this._fieldset = c} className={className} styleName="fieldset">
				<legend>{title}</legend>
				{checkboxes}
			</fieldset>
		);
	}
}

export default CSSModules(CheckboxList, styles);
