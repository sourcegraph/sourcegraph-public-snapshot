import * as React from "react";
import {Symbol} from "sourcegraph/components/symbols/Symbol";

interface Props {
	className?: string;
	width?: number; // appended by "px"
	style?: Object;
	color?: any;
}

export const DownMenu = (props: Props) => <Symbol {...props} viewBox="153 18 11 6"><path d="M159.2 23.3l3.6-3.4c.6-.8.2-2-.7-2h-7c-1 0-1.3 1.2-.7 1.8l3.5 3.5c.4.4 1 .4 1.4 0z" fillRule="evenodd"/></Symbol>;
