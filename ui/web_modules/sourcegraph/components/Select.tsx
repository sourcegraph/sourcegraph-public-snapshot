// tslint:disable

import * as React from "react";
import * as classNames from "classnames";

import * as base from "./styles/_base.css";
import * as styles from "./styles/select.css";
import {DownPointer, Alert} from "./symbols/index";

class Select extends React.Component<any, any> {
	static propTypes = {
		block: React.PropTypes.bool,
		className: React.PropTypes.string,
		children: React.PropTypes.any,
		label: React.PropTypes.string,
		placeholder: React.PropTypes.string,
		helperText: React.PropTypes.string,
		error: React.PropTypes.bool,
		errorText: React.PropTypes.string,
		style: React.PropTypes.object,
		defaultValue: React.PropTypes.string,
	};

	static defaultProps = {
		block: true,
	};

	render(): JSX.Element | null {
		const {style, block, className, placeholder, label, helperText, error, errorText, children, defaultValue} = this.props;
		return (
			<div className={className}>
				{label && <div>{label} <br /></div>}
				<select
					required={true}
					style={style}
					defaultValue={defaultValue}
					className={classNames(styles.select, error ? styles.border_red : styles.border_neutral, block && styles.block)}
					placeholder={placeholder ? placeholder : ""}>
					{children}
				</select>
				<DownPointer style={{marginLeft: "-28px"}} width={11} className={styles.icon} />
				{helperText && <em className={classNames(styles.small, styles.block, base.mt2)}>{helperText}</em>}
				{errorText &&
					<div className={classNames(styles.red, base.mv2)}>
						<Alert width={16} className={classNames(base.mr2, styles.red_fill)} style={{marginTop: "-4px"}} />
						This is an error message.
					</div>
				}
			</div>
		);
	}
}

export default Select;
