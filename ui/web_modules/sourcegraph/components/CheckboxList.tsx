// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as styles from "./styles/checkboxList.css";
import * as classNames from "classnames";

interface Props {
	title: string;
	name: string;
	labels: string[];

	defaultValues: string[];
	className?: string;
}

export class CheckboxList extends React.Component<Props, any> {
	// TODO(slimsag): this should be 'element' type?
	_fieldset: any;

	selected(): string[] {
		let selected: any[] = [];
		for (let input of this._fieldset.querySelectorAll("input")) {
			if (input.checked) {
				selected.push(input.value);
			}
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
			checkboxes.push(<span className={styles.checkbox} key={label}><label><input type="checkbox" name={name} defaultValue={label} defaultChecked={this._isDefaultValue(label)} /> {label}</label></span>);
		}

		return (
			<fieldset ref={(c) => this._fieldset = c} className={classNames(className, styles.fieldset)}>
				<legend>{title}</legend>
				{checkboxes}
			</fieldset>
		);
	}
}
