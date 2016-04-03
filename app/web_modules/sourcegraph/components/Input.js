import React from "react";

import Component from "sourcegraph/Component";

import CSSModules from "react-css-modules";
import styles from "./styles/input.css";

class Input extends Component {
	constructor(props) {
		super(props);
		this.state = {
			value: "",
			valid: true,
		};
		this._handleChange = this._handleChange.bind(this);
		this._handleBlur = this._handleBlur.bind(this);
		this.getValue = this.getValue.bind(this);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		if (state.type === "email") {
			state.isValid = props.validate ?
				(value) => this.refs.input.checkValidity() && props.validate(value) :
				(value) => this.refs.input.checkValidity();
		} else {
			state.isValid = props.validate ? props.validate : (value) => true;
		}
	}

	_handleChange(e) {
		if (e.target) {
			const oldValue = this.state.value;
			const newValue = e.target.value;
			let valid = true;
			if (!this.state.valid) {
				// If error state triggered, update state.valid onChange.
				valid = this.state.isValid(newValue);
			}
			this.setState({
				value: newValue,
				valid: valid,
			}, () => this.state.onChange && this.state.onChange(newValue, oldValue));
		}
	}

	_handleBlur() {
		this.setState({
			valid: this.state.isValid(this.state.value),
		});
	}

	getValue() {
		return this.state.value;
	}

	render() {
		let style = this.state.block ? "block" : "input";
		style = `${style} ${this.state.valid ? "normal" : "error"}`;
		return (
			<input styleName={style}
				ref="input"
				type={this.state.type}
				value={this.state.value}
				placeholder={this.state.placeholder}
				autoFocus={this.state.autoFocus}
				onChange={this._handleChange}
				onBlur={this._handleBlur}
				disabled={this.state.disabled} /> // TODO: styles for disabled input.
		);
	}
}

Input.propTypes = {
	type: React.PropTypes.string.isRequired, // "text", "email", "password"
	autoFocus: React.PropTypes.bool, // "text", "email", "password"
	block: React.PropTypes.bool, // display:inline-block by default
	disabled: React.PropTypes.bool,
	placeholder: React.PropTypes.string,
	onChange: React.PropTypes.func,
	validate: React.PropTypes.func,
};

export default CSSModules(Input, styles, {allowMultiple: true});
