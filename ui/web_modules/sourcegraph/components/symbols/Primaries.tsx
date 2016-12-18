import * as React from "react";
import { Symbol } from "sourcegraph/components/symbols/Symbol";

interface Props {
	className?: string;
	width?: number; // appended by "px"
	style?: Object;
	color?: any;
}

export function Close(props: Props): React.ReactElement<Props> {
	return <Symbol {...props} viewBox="271 18 13 13"><path fillRule="evenodd" d="M283 30c0 .2-.3.3-.5.3s-.4 0-.5-.2l-4.7-4.7-4.7 4.8c-.2.2-.5.3-.8.3-.2 0-.4-.3-.5-.6 0-.2 0-.5.2-.7l4.7-4.7-4.7-4.7c-.3-.3-.3-.8 0-1 .3-.4.8-.4 1 0l4.8 4.6 4.7-4.7c.2-.2.5-.3.7-.2.3 0 .5.3.6.5 0 .3 0 .6-.2.8l-4.7 4.7L283 29c.2 0 .3.3.3.5s0 .4-.2.6z" /></Symbol>;
};
