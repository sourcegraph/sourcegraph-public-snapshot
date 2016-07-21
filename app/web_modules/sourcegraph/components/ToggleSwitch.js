// @flow

import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/toggleSwitch.css";

class ToggleSwitch extends React.Component {
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
		this._toggle = this._toggle.bind(this);
	}

	_toggle() {
		this.setState({checked: !this.state.checked}, () => this.props.onChange(this.state.checked));
	}

	render() {
		return (
			<div styleName="toggle" onClick={this._toggle}>
				<input type="checkbox" name="toggle" styleName="toggle-checkbox" checked={this.state.checked}/>
				<label styleName="toggle-label">
						<span styleName="toggle-inner"></span>
						<span styleName="toggle-switch"></span>
				</label>
			</div>
		);
	}
}

export default CSSModules(ToggleSwitch, styles, {allowMultiple: true});
