// tslint:disable: typedef ordered-imports

import * as React from "react";

import * as styles from "sourcegraph/components/styles/input.css";
import * as base from "sourcegraph/components/styles/_base.css";
import {Alert} from "sourcegraph/components/symbols";
import * as classNames from "classnames";

export interface Props extends React.HTMLAttributes<HTMLInputElement> {
	block?: boolean;
	className?: string;
	// domRef is like ref, but it is called with the <input> DOM element,
	// not this pure wrapper component. <Input domRef={...}> is equivalent
	// to <input ref={...}>.
	domRef?: (c: HTMLInputElement) => void;
	label?: string;
	placeholder?: string;
	helperText?: string;
	error?: boolean;
	errorText?: string;
	style?: any;
}

type State = any;

export class Input extends React.Component<Props, State> {
	render(): JSX.Element | null {
		const {style, domRef, block, className, placeholder, label, helperText, error, errorText} = this.props;
		const other = Object.assign({}, this.props);
		delete other.block;
		delete other.className;
		delete other.domRef;
		delete other.label;
		delete other.placeholder;
		delete other.helperText;
		delete other.error;
		delete other.errorText;
		delete other.style;
		return (
			<div className={className} style={block ? {width: "100%"} : {}}>
				{label && <div className={base.mb2}>{label} <br /></div>}
				<input
					{...other}
					style={style} ref={domRef}
					className={classNames(styles.input, block ? styles.block : null, error ? styles.border_red : styles.border_neutral)}
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
