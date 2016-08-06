// tslint:disable

import * as React from "react";

import CSSModules from "react-css-modules";
import * as styles from "./styles/input.css";
import * as base from "./styles/_base.css";
import {Alert} from "./symbols/index";

class Input extends React.Component<any, any> {
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

	render(): JSX.Element | null {
		const {style, domRef, block, className, placeholder, label, helperText, error, errorText} = this.props;
		return (
			<div className={className} style={block ? {width: "100%"} : {}}>
				{label && <div className={base.mb2}>{label} <br /></div>}
				<input
					{...this.props}
					style={style} ref={domRef}
					className={`${styles.input} ${block ? styles.block : ""} ${error ? styles.border_red : styles.border_neutral}`}
					placeholder={placeholder ? placeholder : ""} />
				{helperText && <em className={`${styles.small} ${styles.block} ${base.mt2}`}>{helperText}</em>}
				{errorText &&
					<div className={`${styles.red} ${base.mv2}`}>
						<Alert width={16} className={`${base.mr2} ${styles.red_fill}`} style={{marginTop: "-4px"}} />
						This is an error message.
					</div>
				}
			</div>
		);
	}
}

export default CSSModules(Input, styles, {allowMultiple: true});
