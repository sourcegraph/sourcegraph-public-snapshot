// @flow

import React from "react";
import Helmet from "react-helmet";
import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";
import {locationForSearch} from "sourcegraph/search/routes";
import GlobalSearchInput from "sourcegraph/search/GlobalSearchInput";
import {Button, Logo} from "sourcegraph/components";
import {PlayIcon} from "sourcegraph/components/Icons";
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
						<strong>Code faster, together.</strong><br/>Search instantly across all the code you use and write. See where and how code is being used, with real usage examples. Integrates with the tools you love.
					</h2>
					<GlobalSearchInput
						name="q"
						size="large"
						value={this.props.location.query.q || ""}
						autoFocus={true}
						onChange={this._handleInput} />
					<div styleName="actions">
						<Button styleName="search-button" type="button" color="blue">Search code</Button>
						<Button styleName="action-button" type="button" color="blue" outline={true}>
							<PlayIcon styleName="action-icon" /> Watch the video
						</Button>
					</div>
					<SearchSettings showAlerts={false} styleName="search-settings" />
				</div>
			</div>
		);
	}
}

export default CSSModules(DashboardContainer, styles, {allowMultiple: true});
