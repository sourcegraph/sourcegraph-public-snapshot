import React from "react";

import Component from "sourcegraph/Component";

import CSSModules from "react-css-modules";
import styles from "./styles/input.css";

class Input extends Component {
	constructor(props) {
		super(props);
		this.state = {
			value: props.defaultValue || "",
			valid: true,
		};
		this._handleChange = this._handleChange.bind(this);
		this._handleBlur = this._handleBlur.bind(this);
		this._handleKeyPress = this._handleKeyPress.bind(this);
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
				valid: valid,
				value: newValue,
			}, () => this.state.onChange && this.state.onChange(newValue, oldValue));
		}
	}

	_handleBlur() {
		this.setState({
			valid: this.state.isValid(this.state.value),
		}, () => this.state.onBlur && this.state.onBlur());
	}

	_handleKeyPress(e) {
		if (e && e.key === "Enter") {
			if (this.state.onSubmit) this.state.onSubmit();
		}
	}

	getValue() {
		return this.state.value;
	}

	focus() {
		this.refs.input.focus();
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
				onFocus={() => this.state.onFocus && this.state.onFocus()}
				onChange={this._handleChange}
				onKeyPress={this._handleKeyPress}
				onBlur={this._handleBlur}
				disabled={this.state.disabled} /> // TODO: styles for disabled input.
		);
	}
}

Input.propTypes = {
	type: React.PropTypes.string.isRequired, // "text", "email", "password"
	defaultValue: React.PropTypes.string,
	autoFocus: React.PropTypes.bool,
	block: React.PropTypes.bool, // display:inline-block by default
	disabled: React.PropTypes.bool,
	placeholder: React.PropTypes.string,
	onChange: React.PropTypes.func,
	onFocus: React.PropTypes.func,
	onBlur: React.PropTypes.func,
	onSubmit: React.PropTypes.func, // called when "Enter" key pressed
	validate: React.PropTypes.func,
};

export default CSSModules(Input, styles, {allowMultiple: true});
