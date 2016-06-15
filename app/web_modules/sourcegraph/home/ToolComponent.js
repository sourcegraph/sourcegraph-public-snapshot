import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/Tools.css";
import base from "sourcegraph/components/styles/_base.css";
import {Heading, Panel, Button} from "sourcegraph/components";
import {CloseIcon} from "sourcegraph/components/Icons";
import Modal from "sourcegraph/components/Modal";
import {TriangleRightIcon, TriangleDownIcon, CheckIcon} from "sourcegraph/components/Icons";
import InterestForm from "./InterestForm";

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
		this.context.eventLogger.logEvent("ToolBackButtonClicked", {toolType: this.props.location.pathname.split("/").slice(-1)[0]});
		this.context.router.replace({...this.props.location, pathname: this.props.location.pathname.split("/").slice(0, -1).join("/")});
	}

	_toggleView() {
		this.setState({formVisible: !this.state.formVisible});
		this.context.router.replace({...this.props.location, query: {...this.props.location.query, expanded: this.state.formVisible}});
		this.context.eventLogger.logEvent("EmailSubscriptionViewToggled", {visible: this.state.formVisible});
	}

	_getVisibility() {
		return this.state.formVisible;
	}

	_submitInterestForm() {
		this.context.eventLogger.logEvent("SubmitEmailSubscription", {page_name: this.props.supportedTool.hero.title});
		window.localStorage["email_subscription_submitted"] = "true";
		this.setState({
			submitted: true,
		});
	}

	_optionalFormView() {
		if (!this.props.supportedTool.interestForm) {
			return <div/>;
		}

		if (!this.state.submitted) {
			return (<div>
				<div styleName="dont-see-div">
					<a styleName="dont-see-link" onClick={this._toggleView.bind(this)}>{this.state.formVisible ? <TriangleDownIcon /> : <TriangleRightIcon />}{this.props.supportedTool.interestForm.title}</a>
				</div>
				<div className={base.mb5} styleName={this.state.formVisible ? "visible" : "invisible"}>
					<InterestForm onSubmit={this._submitInterestForm.bind(this)} />
				</div>
			</div>);
		}

		return (
			<div className={base.mb3}>
				<CheckIcon styleName="check-icon" /><span>{this.props.supportedTool.interestForm.submittedTitle}</span>
			</div>
		);
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
						<span>{this.props.supportedTool.hero.img ? <img styleName="tool-img" src={`${this.context.siteConfig.assetsRoot}${this.props.supportedTool.hero.img}`}></img> : <div styleName="tool-img"></div>}</span>
						<div>
							<Heading align="left" level="2" className={base.pt5}>{this.props.supportedTool.hero.title}</Heading>
							<div styleName="tool-item-paragraph">
								<b>{this.props.supportedTool.hero.subtitle}</b>
								<br/><br/>
								{this.props.supportedTool.hero.paragraph}
							</div>
						</div>
					</div>
					<div styleName="button-container">{this.props.supportedTool.primaryButton}</div>
					{this._optionalFormView()}
					{this.props.supportedTool.secondaryButton}
					{this.props.supportedTool.gif && <div styleName="tool-gif-container">
						<img styleName="tool-gif" src={`${this.context.siteConfig.assetsRoot}${this.props.supportedTool.gif}`}></img>
					</div>}
				</Panel>
			</Modal>
		);
	}
}

export default CSSModules(ToolComponent, styles, {allowMultiple: true});
