// tslint:disable: typedef ordered-imports

import * as React from "react";

interface Props {
	className?: string;
	height?: number;
	width?: number;
	style?: any;
}

type State = any;

export class GitHubLogo extends React.Component<Props, State> {
	static defaultProps = {
		width: 16,
	};

	render(): JSX.Element | null {
		let style = this.props.style || {};
		return <svg className={this.props.className} width={`${this.props.width}px`} style={style} viewBox="0 0 61 60"><path d="M30.5 0C13.656 0 0 13.656 0 30.5c0 13.478 8.738 24.908 20.86 28.94 1.524.28 2.08-.663 2.08-1.467 0-.726-.028-3.13-.043-5.678-8.48 1.843-10.274-3.596-10.274-3.596-1.388-3.523-3.386-4.462-3.386-4.462-2.772-1.89.21-1.854.21-1.854 3.063.212 4.675 3.143 4.675 3.143 2.722 4.66 7.14 3.313 8.876 2.53.277-1.964 1.066-3.312 1.936-4.072-6.77-.77-13.893-3.388-13.893-15.075 0-3.332 1.19-6.05 3.138-8.185-.31-.775-1.36-3.878.3-8.076 0 0 2.56-.82 8.39 3.127 2.43-.68 5.04-1.015 7.632-1.026 2.59.012 5.2.35 7.636 1.03 5.82-3.95 8.38-3.127 8.38-3.127 1.67 4.202.62 7.3.306 8.072 1.955 2.135 3.135 4.853 3.135 8.185 0 11.717-7.133 14.297-13.928 15.053 1.097.946 2.07 2.8 2.07 5.644 0 4.077-.042 7.365-.042 8.37 0 .81.555 1.76 2.1 1.463C52.268 55.4 61 43.97 61 30.5 61 13.656 47.344 0 30.5 0z" fill="#7793AE" fillRule="evenodd" opacity=".37"/></svg>;
	}
}
