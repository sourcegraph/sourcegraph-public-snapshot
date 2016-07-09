// @flow

import React from "react";

import CSSModules from "react-css-modules";
import styles from "./styles/input.css";
import base from "./styles/_base.css";
import {Alert} from "./symbols";

class Input extends React.Component {
	static propTypes = {
		block: React.PropTypes.bool,
		className: React.PropTypes.string,
		// domRef is like ref, but it is called with the <input> DOM element,
		// not this pure wrapper component. <Input domRef={...}> is equivalent
		// to <input ref={...}>.
		domRef: React.PropTypes.func,
		label: React.PropTypes.string,
		placeholder: React.PropTypes.string,
		helperText: React.PropTypes.string,
		error: React.PropTypes.bool,
		errorText: React.PropTypes.string,
		style: React.PropTypes.object,
	};

	render() {
		const {style, domRef, block, className, placeholder, label, helperText, error, errorText} = this.props;
		return (
			<div className={className}>
				{label && <div className={base.mb2}>{label} <br /></div>}
				<input
					{...this.props}
					style={style} ref={domRef}
					styleName={`input ${block ? "block" : ""} ${error ? "border-red" : "border-neutral"}`}
					placeholder={placeholder ? placeholder : ""} />
				{helperText && <em styleName="small block" className={base.mt2}>{helperText}</em>}
				{errorText &&
					<div styleName="red" className={base.mv2}>
						<Alert width={16} className={base.mr2} style={{marginTop: "-4px"}} styleName="red-fill" />
						This is an error message.
					</div>
				}
			</div>
		);
	}
}

export default CSSModules(Input, styles, {allowMultiple: true});
