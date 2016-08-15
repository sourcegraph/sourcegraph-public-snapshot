// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as styles from "sourcegraph/components/styles/toggleSwitch.css";

interface Props {
	defaultChecked?: boolean;
	onChange: (checked: boolean) => void;
}

interface State {
	checked: boolean;
}

export class ToggleSwitch extends React.Component<Props, State> {
	static defaultProps = {
		defaultChecked: false,
	};

	state: State = {
		checked: false,
	};

	constructor(props: Props) {
		super(props);
		this.state = {
			checked: props.defaultChecked || false,
		};
	}

	_toggle() {
		this.setState({checked: !this.state.checked}, () => this.props.onChange(this.state.checked));
	}

	render(): JSX.Element | null {
		return (
			<div className={styles.toggle} onClick={this._toggle.bind(this)}>
				<input type="checkbox" name="toggle" className={styles.toggle_checkbox} checked={this.state.checked} readOnly={true}/>
				<label className={styles.toggle_label}>
						<span className={styles.toggle_inner}></span>
						<span className={styles.toggle_switch}></span>
				</label>
			</div>
		);
	}
}
