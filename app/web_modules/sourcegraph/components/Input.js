// @flow

import React from "react";

import CSSModules from "react-css-modules";
import styles from "./styles/input.css";
import base from "./styles/_base.css";

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
				{label} <br/>
				<input
					style={style} ref={domRef}
					styleName={`input ${block ? "block" : ""} ${error ? "border-red" : "border-neutral"}`}
					placeholder={placeholder ? placeholder : ""} />
				{helperText && <em styleName="small block" className={base.mt2}>{helperText}</em>}
				{errorText &&
					<div styleName="red" className={base.mv2}>
						<svg xmlns="http://www.w3.org/2000/svg" width="16" height="14" styleName="v-mid" className={base.mr2} style={{marginTop: "-4px"}}><path fill="#F56869" fill-rule="evenodd" d="M15.2 11.6L9.2 1C9 1 8.8 1 8.7.8 8.2 0 7.2 0 6.6.7l-.3.4-6 10.6-.2.4c-.2 1 .3 1.7 1.2 2h12.4c.7 0 1.3-.2 1.6-.8.3-.5.3-1 0-1.6zm-6.8 0c-.2.3-.4.4-.7.4-.3 0-.5 0-.8-.3-.3-.2-.4-.5-.4-.7 0-.3 0-.6.3-.8.2-.2.4-.3.7-.3.3 0 .5 0 .7.2.2.2.3.5.3.8 0 .2 0 .5-.3.7zm.3-6l-.2.8-.3 1.2L8 9.3h-.5c0-.7-.2-1.3-.3-1.7 0-.5-.2-1-.3-1.2l-.3-.8V5c0-.3 0-.6.2-.8 0-.2.4-.3.7-.3.3 0 .5 0 .7.2.2.2.3.5.3.7v.5z"/></svg>
						This is an error message.
					</div>
				}
			</div>
		);
	}
}

export default CSSModules(Input, styles, {allowMultiple: true});
