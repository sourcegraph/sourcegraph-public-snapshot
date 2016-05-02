import React from "react";

import CSSModules from "react-css-modules";
import styles from "./App.css";

@CSSModules(styles)
export default class SearchInput extends React.Component {
	static propTypes = {
		onSubmit: React.PropTypes.func.isRequired,
		placeholder: React.PropTypes.string,
	};

	constructor(props) {
		super(props);
	}

	handleSubmit = (e) => {
		const text = e.target.value.trim();
		if (e.which === 13) {
			this.props.onSubmit(text);
		}
	};

	render() {
		return (
			<input styleName="input"
				type="text"
				autoFocus={true}
				placeholder={this.props.placeholder}
				onKeyDown={this.handleSubmit}
				className="sg-input" />
		);
	}
}
