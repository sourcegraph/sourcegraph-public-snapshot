// tslint:disable

import * as React from "react";

type Props = {
	className?: string,
	width?: number, // appended by "px"
};

export class Repository extends React.Component<Props, any> {
	static defaultProps = {
		width: 16,
	};

	render(): JSX.Element | null {
		return <svg viewBox="0 0 24 24" width={this.props.width} className={this.props.className} style={{verticalAlign: "middle"}}><defs><path id="a" d="M0 3c0-1.7 1.3-3 3-3h18c1.7 0 3 1.3 3 3v18c0 1.7-1.3 3-3 3H3c-1.7 0-3-1.3-3-3V3z"/></defs><g fill="none" fillRule="evenodd"><path stroke="#D5E5F2" strokeOpacity=".6" d="M12.5 4.5v78.2" strokeLinecap="square"/><g><mask id="b" fill="#fff"><use xlinkHref="#a" /></mask><use fill="#D5E5F2" xlinkHref="#a"/><path fill="#7793AE" fillOpacity=".5" d="M-5 7h33v1H-5zM-5 16h33v1H-5z" mask="url(#b)"/><ellipse cx="4" cy="4" fill="#7793AE" mask="url(#b)" rx="1" ry="1"/><ellipse cx="4" cy="12" fill="#7793AE" mask="url(#b)" rx="1" ry="1"/><ellipse cx="4" cy="20" fill="#7793AE" mask="url(#b)" rx="1" ry="1"/><ellipse cx="7" cy="4" fill="#7793AE" mask="url(#b)" rx="1" ry="1"/><ellipse cx="7" cy="12" fill="#7793AE" mask="url(#b)" rx="1" ry="1"/><ellipse cx="7" cy="20" fill="#7793AE" mask="url(#b)" rx="1" ry="1"/></g></g></svg>;
	}
}
