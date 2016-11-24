import * as React from "react";
import { Symbol } from "sourcegraph/components/symbols/Symbol";

interface Props {
	className?: string;
	width?: number; // appended by "px"
	style?: Object;
	color?: any;
}

export function Check(props: Props): React.ReactElement<Props> {
	return <Symbol {...props} viewBox="670 182 20 15"><path fillRule="evenodd" d="M670 190l2-2 5 5 11-11 2 2-13 13" /></Symbol>;
};

export function CheckCircle(props: Props): React.ReactElement<Props> {
	return <Symbol {...props} viewBox="620 182 20 20"><path fillRule="evenodd" d="M622.93 199.07c3.904 3.906 10.236 3.906 14.14 0 3.906-3.904 3.906-10.236 0-14.14-3.904-3.906-10.236-3.906-14.14 0-3.906 3.904-3.906 10.236 0 14.14zm12.727-1.413c3.124-3.124 3.124-8.19 0-11.314-3.124-3.124-8.19-3.124-11.314 0-3.124 3.124-3.124 8.19 0 11.314 3.124 3.124 8.19 3.124 11.314 0zM624 192l2-2 3 3 5-5 2 2-7 7-5-5z" /></Symbol>;
};

export function ChevronDown(props: Props): React.ReactElement<Props> {
	return <Symbol {...props} viewBox="720 182 12 8"><path fillRule="evenodd" d="M725.293 188.95l.707.707 5.657-5.657-1.414-1.414-4.243 4.242-4.243-4.242-1.414 1.414" /></Symbol>;
};

export function ChevronRight(props: Props): React.ReactElement<Props> {
	return <Symbol {...props} viewBox="92 5 6 9"><path fillRule="evenodd" d="M97 10l.7-.6-4-4-1 1 3 3-3 3 1 1" /></Symbol>;
};

export function Close(props: Props): React.ReactElement<Props> {
	return <Symbol {...props} viewBox="410 232 18 18"><path fillRule="evenodd" d="M419 239.586l-7.07-7.07-1.415 1.413 7.07 7.07-7.07 7.07 1.414 1.415 7.07-7.07 7.07 7.07 1.415-1.414-7.07-7.07 7.07-7.07-1.414-1.415" /></Symbol>;
};

export function Flag(props: Props): React.ReactElement<Props> {
	return <Symbol {...props} viewBox="484 332 20 20"><path d="M491.667 344H486v8h-2v-20h12l.333 2H504l-3 6 3 6h-12" fillRule="evenodd" /></Symbol>;
};

export function UserAdd(props: Props): React.ReactElement<Props> {
	return <Symbol {...props} viewBox="0 0 20 20"><path fillRule="evenodd" d="M2 6H0v2h2v2h2V8h2V6H4V4H2v2zm7 0c0-1.7 1.3-3 3-3s3 1.3 3 3v2c0 1.7-1.3 3-3 3S9 9.7 9 8V6zm11 9c-2.4-1.2-5-2-8-2s-5.6.8-8 2v3h16v-3z" /></Symbol>;
};

export function Link(props: Props): React.ReactElement<Props> {
	return <Symbol {...props} viewBox="0 0 20 20"><path fillRule="evenodd" d="M14 11v-1c0-2.8-2.2-5-5-5H5c-2.8 0-5 2.2-5 5s2.2 5 5 5h1c-.5-.6-1-1.3-1.3-2-1.5-.2-2.7-1.5-2.7-3 0-1.7 1.3-3 3-3h4c1.7 0 3 1.3 3 3 0 .4 0 .7-.2 1h2zM6 9v1c0 2.8 2.2 5 5 5h4c2.8 0 5-2.2 5-5s-2.2-5-5-5h-1c.5.6 1 1.3 1.3 2 1.5.2 2.7 1.5 2.7 3 0 1.7-1.3 3-3 3h-4c-1.7 0-3-1.3-3-3 0-.4 0-.7.2-1h-2z" /></Symbol>;
};
