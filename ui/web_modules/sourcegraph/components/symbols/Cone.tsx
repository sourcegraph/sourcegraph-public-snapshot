// tslint:disable: typedef ordered-imports

import * as React from "react";

interface Props {
	className?: string;
	width?: number; // appended by "px"
	style?: any;
}

type State = any;

export class Cone extends React.Component<Props, State> {
	static defaultProps = {
		width: 16,
	};

	render(): JSX.Element | null {
		return <svg viewBox="32 992 16 17" className={this.props.className} width={`${this.props.width}px`} style={this.props.style} xmlns="http://www.w3.org/2000/svg"><path d="M32 1007.5c0 .5.4.8 1 .8h14c.6 0 1-.3 1-.8s-.4-1-1-1h-1.7l-4.5-14c0-.3-.4-.5-.8-.5s-.7.2-.8.6l-4.5 14H33c-.6 0-1 .4-1 1zm5-2.6l.7-2.4h4.6l.8 2.3h-6zm1.6-5.5l.8-2.7h1.2l.8 2.7h-2.8z" fill-rule="evenodd"/></svg>;
	}
}
