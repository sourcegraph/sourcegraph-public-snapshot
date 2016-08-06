// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";
import * as styles from "./styles/checkboxList.css";

class CheckboxList extends React.Component<any, any> {
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
		let selected: any[] = [];
		for (let input of this._fieldset.querySelectorAll("input")) {
			if (input.checked) selected.push(input.value);
		}
		return selected;
	}

	_isDefaultValue(s: string): boolean {
		return this.props.defaultValues && this.props.defaultValues.indexOf(s) !== -1;
	}

	render(): JSX.Element | null {
		const {className, title, name, labels} = this.props;
		let checkboxes: any[] = [];
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
