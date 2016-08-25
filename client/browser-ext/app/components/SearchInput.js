import React from "react";

import CSSModules from "react-css-modules";
import styles from "./App.css";
import EventLogger from "../analytics/EventLogger"

@CSSModules(styles)
export default class SearchInput extends React.Component {
	static propTypes = {
		onSubmit: React.PropTypes.func.isRequired,
		onChange: React.PropTypes.func.isRequired,
		placeholder: React.PropTypes.string,
	};

	constructor(props) {
		super(props);
		this.state = {
			value: "",
		}
	}

	handleSubmit = (e) => {
		const text = e.target.value.trim();
		if (e.which === 13) {
			EventLogger.logEventForCategory("Repo", "Click", "ForceSubmitGitHubSearchQuery", {query: text});
			this.props.onSubmit(text);
		}
	};

	handleChange = (e) => {
		const text = e.target.value.trim();
		EventLogger.logEventForCategory("Repo", "Success", "UpdateGitHubSearchQuery", {query: text});
		this.setState({value: text}, () => {
			if (this.props.onChange) this.props.onChange(text);
		})
	};

	render() {
		return (
			<input styleName="input"
				value={this.value}
				type="text"
				autoFocus={true}
				placeholder={this.props.placeholder}
				onKeyDown={this.handleSubmit}
				onChange={this.handleChange}
				className="sg-input" />
		);
	}
}
