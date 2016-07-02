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
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

class BetaInterestForm extends React.Component {

	static propTypes = {
		onSubmit: React.PropTypes.func,
		formClass: React.PropTypes.string,
		language: React.PropTypes.string,
	}

	static contextTypes = {
		eventLogger: React.PropTypes.object.isRequired,
	};

	state = {
		submitted: false,
		formError: "",
		resp: null,
	};

	componentDidMount() {
		this._dispatcherToken = Dispatcher.Stores.register(this._onDispatch.bind(this));
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

		this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ENGAGEMENT, AnalyticsConstants.ACTION_SUCCESS, "SubmitBetaSubscription");
	}

	render() {
		if (this.state.resp && !this.state.resp.Error) {
			return (<p>Thank you for registering. You will hear from us soon.</p>);
		}
		let [formClass, language] = [this.props.formClass, this.props.language];

		return (
			<form styleName="form" className={formClass} onSubmit={this._sendForm.bind(this)}>
					<div styleName="row">
						<Input domRef={(c) => this._fullName = c} block={true} type="text" name="fullName" placeholder="Name" required={true} />
					</div>
					<div styleName="row">
						<Input domRef={(c) => this._email = c} block={true} type="email" name="email" placeholder="Email address" required={true} />
					</div>
					<div styleName="row">
						<CheckboxList ref={(c) => this._editors = c} title="Preferred editors" name="editors" labels={editors} />
					</div>
					<div styleName="row">
						<CheckboxList ref={(c) => this._languages = c} title="Preferred languages" name="languages" labels={languages} defaultValue={language ? langName(language) : null} />
					</div>
					<div styleName="row">
						<textarea ref={(c) => this._message = c} styleName="textarea" name="message" placeholder="Other / comments"></textarea>
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
