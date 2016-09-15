import * as React from "react";
import {Symbol} from "sourcegraph/components/symbols/Symbol";

interface Props {
	className?: string;
	width?: number; // appended by "px"
	style?: Object;
	color?: any;
}

export const Dismiss = (props: Props) => <Symbol	{...props} viewBox="453 12 20 21"><path d="M463 12c-5.5 0-10 4.5-10 10s4.5 10 10 10 10-4.5 10-10-4.5-10-10-10zm0 18.14c-4.5 0-8.14-3.65-8.14-8.14 0-4.5 3.65-8.14 8.14-8.14 4.5 0 8.14 3.65 8.14 8.14 0 4.5-3.65 8.14-8.14 8.14zm3-5.15c-.37.35-.96.35-1.32 0L463 23.3l-1.68 1.7c-.36.36-.95.36-1.3 0-.37-.37-.37-.96 0-1.32l1.67-1.68-1.7-1.68c-.35-.36-.35-.95 0-1.3.37-.37.96-.37 1.32 0l1.68 1.67 1.68-1.7c.36-.35.95-.35 1.3 0 .37.37.38.96 0 1.32L464.32 22l1.7 1.68c.36.36.35.95 0 1.3z" fillRule="evenodd"/></Symbol>;
