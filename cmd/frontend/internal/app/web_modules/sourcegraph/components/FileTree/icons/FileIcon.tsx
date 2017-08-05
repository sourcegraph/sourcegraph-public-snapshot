import * as React from "react";
import * as IconBase from "react-icon-base";

export class FileIcon extends React.Component<any, {}> {
	render(): JSX.Element {
		return (<IconBase viewBox="0 0 40 40" {...this.props}>
			<g id="Octicons" stroke="none" strokeWidth="1" fill="none" fillRule="evenodd">
				<g id="file" fill="#000000">
					<path d="M6,5 L2,5 L2,4 L6,4 L6,5 L6,5 Z M2,8 L9,8 L9,7 L2,7 L2,8 L2,8 Z M2,10 L9,10 L9,9 L2,9 L2,10 L2,10 Z M2,12 L9,12 L9,11 L2,11 L2,12 L2,12 Z M12,4.5 L12,14 C12,14.55 11.55,15 11,15 L1,15 C0.45,15 0,14.55 0,14 L0,2 C0,1.45 0.45,1 1,1 L8.5,1 L12,4.5 L12,4.5 Z M11,5 L8,2 L1,2 L1,14 L11,14 L11,5 L11,5 Z" id="Shape"></path>
				</g>
			</g>
		</IconBase>);
	}
}
