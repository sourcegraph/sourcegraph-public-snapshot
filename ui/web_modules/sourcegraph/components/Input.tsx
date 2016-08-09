// tslint:disable

import * as React from "react";

import * as styles from "./styles/input.css";
import * as base from "./styles/_base.css";
import {Alert} from "./symbols/index";
import * as classNames from "classnames";

type Props = {
	block?: boolean,
	className?: string,
	// domRef is like ref, but it is called with the <input> DOM element,
	// not this pure wrapper component. <Input domRef={...}> is equivalent
	// to <input ref={...}>.
	domRef?: () => void,
	label?: string,
	placeholder?: string,
	helperText?: string,
	error?: boolean,
	errorText?: string,
	style?: any,

	[key: string]: any,
};

export class Input extends React.Component<Props, any> {
	render(): JSX.Element | null {
		const {style, domRef, block, className, placeholder, label, helperText, error, errorText} = this.props;
		return (
			<div className={className} style={block ? {width: "100%"} : {}}>
				{label && <div className={base.mb2}>{label} <br /></div>}
				<input
					{...this.props}
					style={style} ref={domRef}
					className={classNames(styles.input, block && styles.block, error ? styles.border_red : styles.border_neutral)}
					placeholder={placeholder ? placeholder : ""} />
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
