import * as classNames from "classnames";
import * as React from "react";
import * as styles from "sourcegraph/components/styles/checkboxList.css";

interface Props {
	title: string;
	name: string;
	labels: string[];
	values?: string[];

	defaultValues: string[];
	className?: string;
	style?: React.CSSProperties;

	onChange: (list: string[]) => void;
}

export class CheckboxList extends React.Component<Props, {}> {
	private fieldset: HTMLFieldSetElement;

	selected(): void {
		let selected: any[] = [];
		for (const input of Array.from(this.fieldset.querySelectorAll("input"))) {
			if (input.checked) {
				selected.push(input.value);
			}
		}
		this.props.onChange(selected);
	}

	private isDefaultValue(s: string): boolean {
		return this.props.defaultValues && this.props.defaultValues.indexOf(s) !== -1;
	}

	render(): JSX.Element | null {
		const { className, title, name, labels, values, style } = this.props;
		let checkboxes: any[] = [];
		for (let i = 0; i < labels.length; i++) {
			const value = values ? values[i] : labels[i];
			checkboxes.push(<span className={styles.checkbox} key={value}><label><input type="checkbox" name={name} defaultValue={value} defaultChecked={this.isDefaultValue(value)} /> {labels[i]}</label></span>);
		}

		return (
			<fieldset ref={(c) => this.fieldset = c} style={style} className={classNames(className, styles.fieldset)} onChange={this.selected.bind(this)}>
				<legend>{title}</legend>
				{checkboxes}
			</fieldset>
		);
	}
}
