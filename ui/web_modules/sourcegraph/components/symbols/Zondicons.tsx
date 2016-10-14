import * as React from "react";
import {Symbol} from "sourcegraph/components/symbols/Symbol";

interface Props {
	className?: string;
	width?: number; // appended by "px"
	style?: Object;
	color?: any;
}

export function Check(props: Props): React.ReactElement<Props> {
	return <Symbol {...props} viewBox="670 182 20 15"><path fillRule="evenodd" d="M670 190l2-2 5 5 11-11 2 2-13 13"/></Symbol>;
};

export function CheckCircle(props: Props): React.ReactElement<Props> {
	return <Symbol {...props} viewBox="620 182 20 20"><path fillRule="evenodd" d="M622.93 199.07c3.904 3.906 10.236 3.906 14.14 0 3.906-3.904 3.906-10.236 0-14.14-3.904-3.906-10.236-3.906-14.14 0-3.906 3.904-3.906 10.236 0 14.14zm12.727-1.413c3.124-3.124 3.124-8.19 0-11.314-3.124-3.124-8.19-3.124-11.314 0-3.124 3.124-3.124 8.19 0 11.314 3.124 3.124 8.19 3.124 11.314 0zM624 192l2-2 3 3 5-5 2 2-7 7-5-5z"/></Symbol>;
};

export function ChevronDown(props: Props): React.ReactElement<Props> {
	return <Symbol {...props} viewBox="720 182 12 8"><path fillRule="evenodd" d="M725.293 188.95l.707.707 5.657-5.657-1.414-1.414-4.243 4.242-4.243-4.242-1.414 1.414"/></Symbol>;
};

export function Close(props: Props): React.ReactElement<Props> {
	return <Symbol {...props} viewBox="410 232 18 18"><path fillRule="evenodd" d="M419 239.586l-7.07-7.07-1.415 1.413 7.07 7.07-7.07 7.07 1.414 1.415 7.07-7.07 7.07 7.07 1.415-1.414-7.07-7.07 7.07-7.07-1.414-1.415"/></Symbol>;
};

export function Flag(props: Props): React.ReactElement<Props> {
	return <Symbol {...props} viewBox="484 332 20 20"><path d="M491.667 344H486v8h-2v-20h12l.333 2H504l-3 6 3 6h-12" fillRule="evenodd"/></Symbol>;
};
