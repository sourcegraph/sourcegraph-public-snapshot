import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/Tools.css";
import base from "sourcegraph/components/styles/_base.css";
import {Heading, Panel, Button} from "sourcegraph/components";
import {CloseIcon} from "sourcegraph/components/Icons";
import Modal from "sourcegraph/components/Modal";

class ToolComponent extends React.Component {

	static propTypes = {
		location: React.PropTypes.object.isRequired,
		supportedTool: React.PropTypes.object.isRequired,
	};

	static contextTypes = {
		siteConfig: React.PropTypes.object.isRequired,
		eventLogger: React.PropTypes.object.isRequired,
		router: React.PropTypes.object.isRequired,
	};

	_dismissModal() {
		this.context.eventLogger.logEvent("ToolBackButtonClicked", {toolType: this.props.location.query.tool});
		this.context.router.replace({...this.props.location, query: ""});
	}

	render() {
		return (
			<Modal onDismiss={this._dismissModal.bind(this)}>
					<div styleName="tool-item">
						<Panel hoverLevel="high">
							<span styleName="panel-cta">
							<Button onClick={this._dismissModal.bind(this)} color="white">
								<CloseIcon className={base.pt2} />
							</Button>
							</span>
							<div styleName="flex-container">
								<span><img styleName="tool-img" src={`${this.context.siteConfig.assetsRoot}${this.props.supportedTool.hero.img}`}></img></span>
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
							{this.props.supportedTool.secondaryButton}
							{this.props.supportedTool.gif && <div styleName="tool-gif-container">
								<img styleName="tool-gif" src={`${this.context.siteConfig.assetsRoot}${this.props.supportedTool.gif}`}></img>
							</div>}
						</Panel>
					</div>
			</Modal>
		);
	}
}

export default CSSModules(ToolComponent, styles);
