import * as React from "react";
import {Symbol} from "sourcegraph/components/symbols/Symbol";

interface Props {
	className?: string;
	width?: number; // appended by "px"
	style?: Object;
	color?: any;
}

export const Search = (props: Props) => <Symbol {...props} viewBox="16 14 15 15"><path d="M30.438 27.362l-2.608-2.608c.815-1.093 1.296-2.447 1.296-3.91 0-3.62-2.944-6.564-6.563-6.564-3.62 0-6.563 2.944-6.563 6.563 0 3.62 2.944 6.563 6.563 6.563 1.464 0 2.818-.48 3.91-1.296l2.61 2.608c.186.188.43.282.677.282.246 0 .49-.094.678-.282.376-.374.376-.982 0-1.356zm-7.875-1.876c-2.56 0-4.643-2.083-4.643-4.643 0-2.56 2.083-4.643 4.643-4.643 2.56 0 4.643 2.083 4.643 4.643 0 2.56-2.083 4.643-4.643 4.643z"  fillRule="even-odd"/></Symbol>;
