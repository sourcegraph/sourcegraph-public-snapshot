import * as React from "react";
import {Symbol} from "sourcegraph/components/symbols/Symbol";

interface Props {
	className?: string;
	width?: number; // appended by "px"
	style?: React.CSSProperties;
	color?: string;
}

export const Check = (props: Props) => <Symbol {...props} viewBox="33 207 17 16"><path d="M40.66 217.73c.3 0 .6-.12.83-.34l7.65-7.52c.46-.46.47-1.2 0-1.67-.44-.45-1.2-.46-1.65 0l-7.2 7.05h.7l-2.2-2.17c-.46-.45-1.2-.45-1.66.02-.45.46-.45 1.2.02 1.66l2.68 2.63c.22.2.52.33.82.33zm6.95-2.6c0-.65-.52-1.18-1.17-1.18-.65 0-1.17.53-1.17 1.18 0 2.53-2.1 4.6-4.7 4.6s-4.7-2.07-4.7-4.6 2.1-4.6 4.7-4.6c.63 0 1.25.13 1.83.38.6.26 1.28-.03 1.53-.63s-.04-1.3-.64-1.54c-.87-.36-1.8-.54-2.75-.54-3.88 0-7.05 3.1-7.05 6.93 0 3.83 3.17 6.94 7.05 6.94 3.9 0 7.06-3.1 7.06-6.94z" fillRule="evenodd"/></Symbol>;
