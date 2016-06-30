import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/InterestForm.css";
import {Button} from "sourcegraph/components";
import Selector from "./Selector";
import Dispatcher from "sourcegraph/Dispatcher";
import * as UserActions from "sourcegraph/user/UserActions";
import base from "sourcegraph/components/styles/_base.css";
import {languages, editors} from "./HomeUtils";
import {langName} from "sourcegraph/Language";

class InterestForm extends React.Component {

	static propTypes = {
		onSubmit: React.PropTypes.func.isRequired,
		rowClass: React.PropTypes.string,
		language: React.PropTypes.string,
	}

	constructor(props) {
		super(props);
		this.state = {
			submitted: false,
			formError: null,
		};
	}

	_sendForm(ev) {
		ev.preventDefault();
		const name = ev.currentTarget[0]["value"];
		let firstName = null;
		let lastName = null;
		if (name) {
			const names = name.split(/\s+/);
			firstName = names[0];
			lastName = names.slice(1).join(" ");
		}

		this.props.onSubmit();
		Dispatcher.Backends.dispatch(new UserActions.SubmitEmailSubscription(
			ev.currentTarget[1]["value"].trim(),
			firstName,
			lastName,
			ev.currentTarget[3]["value"].trim(),
			ev.currentTarget[2]["value"].trim(),
			ev.currentTarget[4]["value"].trim(),
		));
	}

	render() {
		return (
			<form onSubmit={this._sendForm.bind(this)}>
				{this.state.formError && <div>{this.state.formError}</div>}
				<div styleName="container">
					<div styleName="table-row" className={this.props.rowClass}>
						<span styleName="full-input">
							<input styleName="input-field" type="text" name="firstName" placeholder="Name" required={true}/>
						</span>
					</div>
					<div styleName="table-row" className={this.props.rowClass}>
						<span styleName="full-input">
							<input styleName="input-field" type="email" name="emailAddress" placeholder="Email address" required={true}/>
						</span>
					</div>
					<div styleName="table-row" className={this.props.rowClass}>
						<span styleName="full-input">
							<Selector title="Preferred editor" valueArray={editors} />
						</span>
					</div>
					<div styleName="table-row" className={this.props.rowClass}>
						<span styleName="full-input">
							<Selector title="Preferred language" valueArray={languages} defaultValue={this.props.language ? langName(this.props.language) : null} />
						</span>
					</div>
					<div styleName="table-row" className={this.props.rowClass}>
						<span styleName="full-input">
							<textarea styleName="input-field" name="message" placeholder="Other / comments"></textarea>
						</span>
					</div>
					<div styleName="table-row" className={`${base.pb4} ${this.props.rowClass || ""}`}>
						<span styleName="full-input">
							<Button styleName="button" type="submit" color="purple">Get early access</Button>
						</span>
					</div>
				</div>
			</form>
		);
	}
}

export default CSSModules(InterestForm, styles);
