// @flow

import React from "react";
import Helmet from "react-helmet";
import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";
import {locationForSearch} from "sourcegraph/search/routes";
import GlobalSearchInput from "sourcegraph/search/GlobalSearchInput";
import {Button, Logo} from "sourcegraph/components";
import {CloudDownloadIcon, PlayIcon} from "sourcegraph/components/Icons";
import SearchSettings from "sourcegraph/search/SearchSettings";

class DashboardContainer extends React.Component {
	static propTypes = {
		location: React.PropTypes.object.isRequired,
	};

	static contextTypes = {
		signedIn: React.PropTypes.bool.isRequired,
		router: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this._handleInput = this._handleInput.bind(this);
	}

	_handleInput: Function;

	_handleInput(ev: KeyboardEvent) {
		if (!(ev.currentTarget instanceof HTMLInputElement)) return;
		if (ev.currentTarget.value) {
			this.context.router.replace(locationForSearch(this.props.location, ev.currentTarget.value, true, false));
		}
	}

	render() {
		return (
			<div>
				<Helmet title="Home" />
				<div styleName="home-container">
					<Logo type="logotype" styleName="logo" />
					<h2 styleName="description">
						<strong>Instant&nbsp;usage&nbsp;examples and other&nbsp;helpful&nbsp;info as&nbsp;you&nbsp;code,</strong> automatically&nbsp;drawn&nbsp;from public&nbsp;and&nbsp;(your&nbsp;own)&nbsp;private&nbsp;code.
					</h2>
					<GlobalSearchInput
						name="q"
						size="large"
						value={this.props.location.query.q || ""}
						autoFocus={true}
						onChange={this._handleInput} />
					<div styleName="search-actions">
						<Button styleName="search-button" type="button" color="blue">Find usage examples</Button>
					</div>
					{this.context.signedIn && <SearchSettings showAlerts={false} location={this.props.location} styleName="search-settings" />}

					<div styleName="user-actions">
						{!this.context.signedIn && <Button styleName="action-link" type="button" color="blue" outline={true}>Sign in</Button>}
						<Button styleName="action-link" type="button" color="blue" outline={true}><CloudDownloadIcon styleName="action-icon" /> Download the app</Button>
						<Button styleName="action-link" type="button" color="blue" outline={true}><PlayIcon styleName="action-icon" /> Watch the video</Button>
					</div>
				</div>
			</div>
		);
	}
}

export default CSSModules(DashboardContainer, styles, {allowMultiple: true});
