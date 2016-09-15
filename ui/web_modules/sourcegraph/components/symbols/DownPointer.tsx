import * as React from "react";
import {Symbol} from "sourcegraph/components/symbols/Symbol";

interface Props {
	className?: string;
	width?: number; // appended by "px"
	style?: Object;
	color?: any;
}

export const DownPointer = (props: Props) => <Symbol {...props} viewBox="0 0 10 6"><path fillRule="evenodd" d="M9.702 1.712l-4.056 4c-.393.388-1.026.383-1.414-.01l-3.944-4C-.1 1.31-.095.676.298.288.69-.1 1.324-.095 1.712.298l3.944 4-1.414-.01 4.056-4C8.69-.1 9.324-.095 9.712.298c.388.393.383 1.026-.01 1.414z"/></Symbol>;
