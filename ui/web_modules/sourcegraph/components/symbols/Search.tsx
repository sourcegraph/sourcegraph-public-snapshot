// tslint:disable

import * as React from "react";

type Props = {
	className?: string,
	height?: number,
	width?: number,
	style?: any,
};

class Search extends React.Component<Props, any> {
	static defaultProps = {
		width: 16,
	};

	render(): JSX.Element | null {
		let style = this.props.style || {};
		style.verticalAlign = "middle";
		return <svg className={this.props.className} width={`${this.props.width}px`} style={style} viewBox="0 0 17 17" xmlns="http://www.w3.org/2000/svg"><path d="M10.6 0c1.2 0 2.2.3 3.2 1 1 .4 1.8 1.2 2.3 2.2.7 1 1 2 1 3.2 0 1-.3 2.2-1 3.2-.4 1-1.2 1.7-2.2 2.3-1 .5-2 .8-3.2.8-1.2 0-2.4-.4-3.4-1l-4.8 4.8c-.3.3-.6.4-1 .4s-.7 0-1-.4c-.3-.3-.4-.6-.4-1s0-.7.4-1l4.8-4.8c-.6-1-1-2.2-1-3.4 0-1.2.3-2.2 1-3.2.5-1 1.2-1.8 2.2-2.3 1-.7 2-1 3.2-1zm0 10.6c.6 0 1 0 1.7-.3.5-.2 1-.5 1.3-1 .4-.3.7-.8 1-1.3.2-.5.3-1 .3-1.6 0-.6-.2-1-.5-1.7-.2-.5-.5-1-1-1.3-.3-.4-.7-.7-1.2-1-.6-.2-1-.3-1.7-.3-.6 0-1 .2-1.6.5-.5.2-1 .5-1.4 1-.4.3-.7.7-1 1.2l-.2 1.7c0 .6 0 1 .3 1.6.2.5.5 1 1 1.4.3.4.8.7 1.3 1l1.6.2z"/></svg>;
	}
}

export default Search;
