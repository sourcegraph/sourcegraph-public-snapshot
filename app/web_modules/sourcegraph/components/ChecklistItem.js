import * as React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/checklistItem.css";
import base from "./styles/_base.css";
import Icon from "./Icon";
import Button from "./Button";

class ChecklistItem extends React.Component {
	static propTypes = {
		className: React.PropTypes.string,
		children: React.PropTypes.any,
		complete: React.PropTypes.bool,
		actionText: React.PropTypes.string, // Button text
		actionOnClick: React.PropTypes.func,
	};

	render() {
		const {className, children, complete, actionText, actionOnClick} = this.props;
		return (
			<div className={className} styleName="item">
				<div styleName={`check_${complete ? "complete" : "incomplete"}`}>
					{complete && <Icon icon="check-green" width="50%" styleName="check" />}
				</div>
				<div styleName={`content${complete ? "_complete" : ""}`}>{children}</div>
				{actionText && !complete && <div styleName="buttonContainer">
					<Button color="green" onClick={actionOnClick} className={base.ph2}>{actionText}</Button>
				</div>}
			</div>
		);
	}
}

export default CSSModules(ChecklistItem, styles, {allowMultiple: true});
