import * as React from "react";
import * as IconBase from "react-icon-base";

export class DirectoryIcon extends React.Component<any, {}> {
	render(): JSX.Element {
		return (<IconBase viewBox="0 0 40 40" {...this.props}>
			<g id="Octicons" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
				<g id="file-directory" fill="#000000">
					<path d="M13,4 L7,4 L7,3 C7,2.34 6.69,2 6,2 L1,2 C0.45,2 0,2.45 0,3 L0,13 C0,13.55 0.45,14 1,14 L13,14 C13.55,14 14,13.55 14,13 L14,5 C14,4.45 13.55,4 13,4 L13,4 Z M6,4 L1,4 L1,3 L6,3 L6,4 L6,4 Z" id="Shape"></path>
				</g>
			</g>
		</IconBase>);
	}
}
