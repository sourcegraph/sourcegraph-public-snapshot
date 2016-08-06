// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";
import * as styles from "./styles/toggleSwitch.css";

class ToggleSwitch extends React.Component<any, any> {
	static propTypes = {
		defaultChecked: React.PropTypes.bool,
		onChange: React.PropTypes.func,
	};

	static defaultProps = {
		defaultChecked: false,
	};

	constructor(props) {
		super(props);
		this.state = {
			checked: props.defaultChecked,
		};
	}

	state: {
		checked: boolean;
	} = {
		checked: false,
	};

	_toggle() {
		this.setState({checked: !this.state.checked}, () => this.props.onChange(this.state.checked));
	}

	render(): JSX.Element | null {
		return (
			<div styleName="toggle" onClick={this._toggle.bind(this)}>
				<input type="checkbox" name="toggle" styleName="toggle_checkbox" checked={this.state.checked} readOnly={true}/>
				<label styleName="toggle_label">
						<span styleName="toggle_inner"></span>
						<span styleName="toggle_switch"></span>
				</label>
			</div>
		);
	}
}

export default CSSModules(ToggleSwitch, styles, {allowMultiple: true});
