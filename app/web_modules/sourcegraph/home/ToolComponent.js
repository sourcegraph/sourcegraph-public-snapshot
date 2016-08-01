import * as React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/Tools.css";
import base from "sourcegraph/components/styles/_base.css";
import {Heading, Panel, Button} from "sourcegraph/components";
import {CloseIcon} from "sourcegraph/components/Icons";
import Modal from "sourcegraph/components/Modal";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {TriangleRightIcon, TriangleDownIcon} from "sourcegraph/components/Icons";
import BetaInterestForm from "./BetaInterestForm";

class ToolComponent extends React.Component {

	static propTypes = {
		location: React.PropTypes.object.isRequired,
		supportedTool: React.PropTypes.object.isRequired,
		formVisible: React.PropTypes.bool,
	};

	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
		router: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this.state = {
			formVisible: this.props.location.query.expanded === "true",
			formError: "none",
			submitted: window.localStorage["email_subscription_submitted"] === "true",
		};
	}

	_dismissModal() {
		this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_TOOLS, AnalyticsConstants.ACTION_CLOSE, "ToolBackButtonClicked", {toolType: this.props.supportedTool.hero.title, page_name: this.props.location.pathname});
		let urlWithOutQueryParams = this.props.location.pathname.split("?")[0];
		this.context.router.replace({...urlWithOutQueryParams, pathname: urlWithOutQueryParams.split("/").slice(0, -1).join("/")});
	}

	_toggleView() {
		this.context.eventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ENGAGEMENT, AnalyticsConstants.ACTION_TOGGLE, "EmailSubscriptionViewToggled", {visible: !this.state.formVisible});
		this.context.router.replace({...this.props.location, query: {...this.props.location.query, expanded: !this.state.formVisible}});
		this.setState({formVisible: !this.state.formVisible});
	}

	_getVisibility() {
		return this.state.formVisible;
	}

	_optionalFormView() {
		if (!this.props.supportedTool.interestForm) {
			return <div/>;
		}

		return (<div>
			<div styleName="dont-see-div">
				<a styleName="dont-see-link" onClick={this._toggleView.bind(this)}>{this.state.formVisible ? <TriangleDownIcon /> : <TriangleRightIcon />}{this.props.supportedTool.interestForm.title}</a>
			</div>
			<div styleName={`beta-container ${this.state.formVisible ? "visible" : "invisible"}`}>
				<BetaInterestForm />
			</div>
		</div>);
	}

	render() {
		return (
			<Modal onDismiss={this._dismissModal.bind(this)}>
				<Panel styleName="tool-item" hoverLevel="high">
					<div styleName="panel-cta">
						<Button onClick={this._dismissModal.bind(this)} color="white">
							<CloseIcon className={base.pt2} />
						</Button>
					</div>
					<div styleName="flex-container">
						{this.props.supportedTool.hero.img ? <span styleName="tool-img-container"><img styleName="large-img" src={`${this.context.siteConfig.assetsRoot}${this.props.supportedTool.hero.img}`}></img></span> : <div styleName="tool-img-container"/>}
						<div>
							<Heading align="left" level="2" className={base.pt5}>{this.props.supportedTool.hero.title}</Heading>
							<div styleName="tool-item-paragraph">
								<b>{this.props.supportedTool.hero.subtitle}</b>
								<br/><br/>
								{this.props.supportedTool.hero.paragraph}
							</div>
						</div>
					</div>
					<div styleName="button-container">{this.props.supportedTool.primaryButton()}</div>
					{this._optionalFormView()}
					{this.props.supportedTool.secondaryButton}
					{this.props.supportedTool.youtube && <div styleName="tool-gif-container">
						<iframe width="420" height="315" src="https://www.youtube.com/embed/ssON7dfaDZo" frameBorder="0" allowFullScreen="true"></iframe>
					</div>}
				</Panel>
			</Modal>
		);
	}
}

export default CSSModules(ToolComponent, styles, {allowMultiple: true});
