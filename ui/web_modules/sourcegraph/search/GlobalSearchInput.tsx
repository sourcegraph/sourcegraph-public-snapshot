import * as classNames from "classnames";
import * as React from "react";
import {Input, Props as InputProps} from "sourcegraph/components/Input";
import {Search} from "sourcegraph/components/symbols";
import * as styles from "sourcegraph/search/styles/GlobalSearchInput.css";

interface Props extends InputProps {
	query: string;
}

export function GlobalSearchInput(props: Props): JSX.Element {
	const inputProps = Object.assign({}, props);
	delete inputProps.query; // TODO(john): this is gross
	return <div className={classNames(styles.flex_fill, styles.relative)}>
		<Search width={16} style={{top: "11px", left: "10px"}} className={classNames(styles.absolute, styles.cool_mid_gray_fill)} />
		<Input
			{...inputProps}
			id="e2etest-search-input"
			type="text"
			block={true}
			autoCorrect="off"
			autoCapitalize="off"
			spellCheck={false}
			autoComplete="off"
			defaultValue={props.query}
			style={{textIndent: "18px"}} />
	</div>;
}
