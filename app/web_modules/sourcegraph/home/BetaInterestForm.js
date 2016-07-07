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
import GitHubAuthButton from "sourcegraph/components/GitHubAuthButton";

type OnChangeListener = () => void;

class BetaInterestForm extends React.Component {

	static propTypes = {
		onSubmit: React.PropTypes.func,
		className: React.PropTypes.string,
		language: React.PropTypes.string,
		loginReturnTo: React.PropTypes.string,
	}

	static contextTypes = {
		user: React.PropTypes.object,
		signedIn: React.PropTypes.bool.isRequired,
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

		Dispatcher.Backends.dispatch(new UserActions.SubmitBetaSubscription(
			"",
			firstName || "",
			lastName || "",
			this._languages.selected(),
			this._editors.selected(),
			this._message["value"].trim(),
		));
	}

	render() {
		if (this.state.resp && !this.state.resp.Error) {
			// Display a "Close" button if there is an onSubmit handler.
			return (<span>
				<p>Success! Return to this page any time to update your favorite editors / languages!</p>
				<p>We'll contact you at <strong>{this.state.resp.EmailAddress}</strong> once a beta has begun.</p>
				{this.props.onSubmit && <Button block={true} type="submit" color="purple" onClick={this.props.onSubmit}>Close</Button>}
			</span>);
		}

		if (!this.context.signedIn) {
			return (<div styleName="cta">
				<p styleName="p">You must sign in to continue.</p>
				<GitHubAuthButton returnTo={this.props.loginReturnTo} color="blue" className={base.mr3}>
					<strong>Sign in with GitHub</strong>
				</GitHubAuthButton>
			</div>);
		}


		let [className, language] = [this.props.className, this.props.language];
		let betaRegistered = this.context.user && this.context.user.BetaRegistered;

		let defaultFullName, defaultMessage;
		let defaultEditors = [];
		let defaultLanguages = [];
		let ls = window.localStorage["beta-interest-form"];
		if (ls) {
			ls = JSON.parse(ls);
			defaultFullName = ls.fullName;
			defaultEditors = ls.editors;
			defaultLanguages = ls.languages;
			defaultMessage = ls.message;
		}

		if (language) defaultLanguages.push(langName(language));

		return (
			<div>
				{betaRegistered && <span>
					<p>You've already registered. We'll contact you once a beta matching your interests has begun.</p>
					<p>Feel free to update your favorite editors / languages using the form below.</p>
				</span>}
				<form styleName="form" className={className} onSubmit={this._sendForm.bind(this)} onChange={this._onChange}>
						<div styleName="row">
							<Input domRef={(c) => this._fullName = c} block={true} type="text" name="fullName" placeholder="Name" required={true} defaultValue={defaultFullName} />
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
							<Button block={true} type="submit" color="purple">{betaRegistered ? "Update my interests" : "Participate in the beta"}</Button>
						</div>
						<div styleName="row" className={base.pb4}>
							{this.state.formError && <strong>{this.state.formError}</strong>}
							{this.state.resp && this.state.resp.Error && <div>{this.state.resp.Error.body}</div>}
						</div>
				</form>
			</div>
		);
	}
}

export default CSSModules(BetaInterestForm, styles);
