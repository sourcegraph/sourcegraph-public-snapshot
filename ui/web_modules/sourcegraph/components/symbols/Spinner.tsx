/* tslint:disable */
/* TypeScript types don't yet recognize animated SVG properties and attributes */

import * as React from "react";
import { Symbol } from "sourcegraph/components/symbols/Symbol";

interface Props {
	className?: string;
	width?: number; // appended by "px"
	style?: Object;
	color?: any;
}

// Spinner is a spinning loading indicator.
export function Spinner(props: Props): React.ReactElement<Props> {
	const svgShapes = `<path fill="none" className="bk" d="M0 0h100v100H0z"/><rect x="43" y="41" width="14" height="18" rx="7" ry="7" fill=${props.color} transform="translate(0 -37)"><animate attributeName="opacity" from="1" to="0" dur="1s" begin="0s" repeatCount="indefinite"/></rect><rect x="43" y="41" width="14" height="18" rx="7" ry="7" fill=${props.color} transform="rotate(45 94.663 68.5)"><animate attributeName="opacity" from="1" to="0" dur="1s" begin="0.125s" repeatCount="indefinite"/></rect><rect x="43" y="41" width="14" height="18" rx="7" ry="7" fill=${props.color} transform="rotate(90 68.5 68.5)"><animate attributeName="opacity" from="1" to="0" dur="1s" begin="0.25s" repeatCount="indefinite"/></rect><rect x="43" y="41" width="14" height="18" rx="7" ry="7" fill=${props.color} transform="rotate(135 57.663 68.5)"><animate attributeName="opacity" from="1" to="0" dur="1s" begin="0.375s" repeatCount="indefinite"/></rect><rect x="43" y="41" width="14" height="18" rx="7" ry="7" fill=${props.color} transform="rotate(180 50 68.5)"><animate attributeName="opacity" from="1" to="0" dur="1s" begin="0.5s" repeatCount="indefinite"/></rect><rect x="43" y="41" width="14" height="18" rx="7" ry="7" fill=${props.color} transform="rotate(-135 42.337 68.5)"><animate attributeName="opacity" from="1" to="0" dur="1s" begin="0.625s" repeatCount="indefinite"/></rect><rect x="43" y="41" width="14" height="18" rx="7" ry="7" fill=${props.color} transform="rotate(-90 31.5 68.5)"><animate attributeName="opacity" from="1" to="0" dur="1s" begin="0.75s" repeatCount="indefinite"/></rect><rect x="43" y="41" width="14" height="18" rx="7" ry="7" fill=${props.color} transform="rotate(-45 5.337 68.5)"><animate attributeName="opacity" from="1" to="0" dur="1s" begin="0.875s" repeatCount="indefinite"/></rect>`;

	return <Symbol {...props} viewBox="0 0 100 100" className="uil-default"><g dangerouslySetInnerHTML={{ __html: svgShapes }} /></Symbol>;
};
