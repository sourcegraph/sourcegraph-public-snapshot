import React from "react";

import CSSModules from "react-css-modules";
import styles from "./App.css";

@CSSModules(styles)
export default class SearchInput extends React.Component {
	static propTypes = {
		onSubmit: React.PropTypes.func.isRequired,
		text: React.PropTypes.string,
		placeholder: React.PropTypes.string,
	};

	constructor(props, context) {
		super(props, context);
		this.state = {
			text: this.props.text || "",
		};
	}

	handleSubmit = (e) => {
		const text = e.target.value.trim();
		if (e.which === 13) {
			this.props.onSubmit(text);
			if (this.props.newTodo) {
				this.setState({text: ""});
			}
		}
	};

	handleChange = (e) => {
		this.setState({text: e.target.value});
	};

	render() {
		return (
			<input styleName="input"
				type="text"
				autoFocus={true}
				placeholder={this.props.placeholder}
				value={this.state.text}
				onChange={this.handleChange}
				onKeyDown={this.handleSubmit} />
		);
	}
}
