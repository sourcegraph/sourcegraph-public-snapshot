// tslint:disable

import * as React from "react";

import * as style from "sourcegraph/components/styles/loader.css";

class Loader extends React.Component<{}, any> {
	render(): JSX.Element | null {
		return (
			<div className={style.loader}>
				<span className={style.loader1}>●</span>
				<span className={style.loader2}>●</span>
				<span className={style.loader3}>●</span>
			</div>
		);
	}
}

export default Loader;
