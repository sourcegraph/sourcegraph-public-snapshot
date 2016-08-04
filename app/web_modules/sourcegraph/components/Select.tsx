// tslint:disable

import * as React from "react";

import CSSModules from "react-css-modules";
import base from "./styles/_base.css";
import styles from "./styles/select.css";
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
					styleName={`select ${error ? "border_red" : "border_neutral"} ${block ? "block" : ""} `}
					placeholder={placeholder ? placeholder : ""}>
					{children}
				</select>
				<DownPointer style={{marginLeft: "-28px"}} width={11} styleName="icon" />
				{helperText && <em styleName="small block" className={base.mt2}>{helperText}</em>}
				{errorText &&
					<div styleName="red" className={base.mv2}>
						<Alert width={16} className={base.mr2} style={{marginTop: "-4px"}} styleName="red_fill" />
						This is an error message.
					</div>
				}
			</div>
		);
	}
}

export default CSSModules(Select, styles, {allowMultiple: true});
