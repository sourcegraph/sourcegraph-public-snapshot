// @flow

import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/BetaInterestForm.css";
import {Button, Input, CheckboxList} from "sourcegraph/components";
import Dispatcher from "sourcegraph/Dispatcher";
import * as UserActions from "sourcegraph/user/UserActions";
import base from "sourcegraph/components/styles/_base.css";
import {languages, editors} from "./HomeUtils";
import {langName} from "sourcegraph/Language";

type OnChangeListener = () => void;

class BetaInterestForm extends React.Component {

	static propTypes = {
		onSubmit: React.PropTypes.func,
		className: React.PropTypes.string,
		language: React.PropTypes.string,
	}

	static contextTypes = {
		eventLogger: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this._onChange = this._onChange.bind(this);
	}

	state = {
		submitted: false,
		formError: "",
		resp: null,
	};

	componentDidMount() {
		this._dispatcherToken = Dispatcher.Stores.register(this._onDispatch.bind(this));

		// Trigger _onChange now to save this.props.language if set.
		if (this.props.language) this._onChange();
	}

	componentWillUnmount() {
		Dispatcher.Stores.unregister(this._dispatcherToken);
	}

	_dispatcherToken: string;

	// TODO(slimsag): these should be 'element' type?
	_fullName: any;
	_email: any;
	_editors: any;
	_languages: any;
	_message: any;

	_onDispatch(action) {
		if (action instanceof UserActions.BetaSubscriptionCompleted) {
			this.setState({resp: action.resp});
		}
	}

	_onChange: OnChangeListener;
	_onChange() {
		window.localStorage["beta-interest-form"] = JSON.stringify({
			fullName: this._fullName["value"],
			email: this._email["value"],
			editors: this._editors.selected(),
			languages: this._languages.selected(),
			message: this._message["value"],
		});
	}

	_sendForm(ev) {
		ev.preventDefault();
		const name = this._fullName["value"];
		let firstName;
		let lastName;
		if (name) {
			const names = name.split(/\s+/);
			firstName = names[0];
			lastName = names.slice(1).join(" ");
		}

		if (this._editors.selected().length === 0) {
			this.setState({formError: "Please select at least one preferred editor."});
			return;
		}
		if (this._languages.selected().length === 0) {
			this.setState({formError: "Please select at least one preferred language."});
			return;
		}

		if (this.props.onSubmit) this.props.onSubmit();

		Dispatcher.Backends.dispatch(new UserActions.SubmitBetaSubscription(
			this._email["value"].trim(),
			firstName || "",
			lastName || "",
			this._languages.selected(),
			this._editors.selected(),
			this._message["value"].trim(),
		));
	}

	render() {
		if (this.state.resp && !this.state.resp.Error) {
			return (<p>Thank you for registering. You will hear from us soon.</p>);
		}
		let [className, language] = [this.props.className, this.props.language];

		let defaultFullName, defaultEmail, defaultMessage;
		let defaultEditors = [];
		let defaultLanguages = [];
		let ls = window.localStorage["beta-interest-form"];
		if (ls) {
			ls = JSON.parse(ls);
			defaultFullName = ls.fullName;
			defaultEmail = ls.email;
			defaultEditors = ls.editors;
			defaultLanguages = ls.languages;
			defaultMessage = ls.message;
		}

		if (language) defaultLanguages.push(langName(language));

		return (
			<form styleName="form" className={className} onSubmit={this._sendForm.bind(this)} onChange={this._onChange}>
					<div styleName="row">
						<Input domRef={(c) => this._fullName = c} block={true} type="text" name="fullName" placeholder="Name" required={true} defaultValue={defaultFullName} />
					</div>
					<div styleName="row">
						<Input domRef={(c) => this._email = c} block={true} type="email" name="email" placeholder="Email address" required={true} defaultValue={defaultEmail} />
					</div>
					<div styleName="row">
						<CheckboxList ref={(c) => this._editors = c} title="Preferred editors" name="editors" labels={editors} defaultValues={defaultEditors} />
					</div>
					<div styleName="row">
						<CheckboxList ref={(c) => this._languages = c} title="Preferred languages" name="languages" labels={languages} defaultValues={defaultLanguages} />
					</div>
					<div styleName="row">
						<textarea ref={(c) => this._message = c} styleName="textarea" name="message" placeholder="Other / comments" defaultValue={defaultMessage}></textarea>
					</div>
					<div styleName="row" className={base.pb4}>
						<Button block={true} type="submit" color="purple">Participate in the beta</Button>
					</div>
					<div styleName="row" className={base.pb4}>
						{this.state.formError && <strong>{this.state.formError}</strong>}
						{this.state.resp && this.state.resp.Error && <div>{this.state.resp.Error.body}</div>}
					</div>
			</form>
		);
	}
}

export default CSSModules(BetaInterestForm, styles);
