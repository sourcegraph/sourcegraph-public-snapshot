import * as React from "react";
import { Symbol } from "sourcegraph/components/symbols/Symbol";

interface Props {
	className?: string;
	width?: number; // appended by "px"
	style?: Object;
	color?: any;
}

export const Cone = (props: Props) => <Symbol {...props} viewBox="32 992 16 17"><path d="M32 1007.488c0 .47.38.852.85.852h14.3c.47 0 .85-.382.85-.852s-.38-.85-.85-.85h-1.873l-4.474-14.07c-.12-.333-.432-.568-.803-.568-.37 0-.684.235-.804.565l-4.473 14.073H32.85c-.47 0-.85.382-.85.85zm4.936-2.552l.715-2.383h4.7l.714 2.383h-6.128zm1.634-5.447l.817-2.724h1.226l.817 2.723h-2.86z" fillRule="evenodd" /></Symbol>;
