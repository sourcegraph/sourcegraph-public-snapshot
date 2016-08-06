// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";
import * as styles from "./styles/checklistItem.css";
import * as base from "./styles/_base.css";
import Icon from "./Icon";
import Button from "./Button";

class ChecklistItem extends React.Component<any, any> {
	static propTypes = {
		className: React.PropTypes.string,
		children: React.PropTypes.any,
		complete: React.PropTypes.bool,
		actionText: React.PropTypes.string, // Button text
		actionOnClick: React.PropTypes.func,
	};

	render(): JSX.Element | null {
		const {className, children, complete, actionText, actionOnClick} = this.props;
		return (
			<div className={`${className} ${styles.item}`}>
				<div className={complete ? styles.check_complete : styles.check_incomplete}>
					{complete && <Icon icon="check-green" width="50%" className={styles.check} />}
				</div>
				<div className={complete ? styles.content_complete : styles.content}>{children}</div>
				{actionText && !complete && <div className={styles.buttonContainer}>
					<Button color="green" onClick={actionOnClick} className={base.ph2}>{actionText}</Button>
				</div>}
			</div>
		);
	}
}

export default CSSModules(ChecklistItem, styles, {allowMultiple: true});
