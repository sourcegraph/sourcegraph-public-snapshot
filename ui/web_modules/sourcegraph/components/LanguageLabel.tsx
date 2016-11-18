import * as React from "react";
import {colors, whitespace} from "sourcegraph/components/utils";

interface Props {
	language: string;
	style?: React.CSSProperties;
}

const langColors = {
	"C": "#A8B9CB",
	"Go": "#76D0F2",
	"HTML": "#F96316",
	"JavaScript": "#FEC570",
	"Python": "#417DAC",
	"TypeScript": "#0D79C9",
};

export function LanguageLabel({language, style}: Props): JSX.Element {
	return <div style={style}>
		<span style={{
			display: "inline-block",
			backgroundColor: language ? langColors[language] :  colors.blue(),
			width: 7,
			height: 7,
			borderRadius: "50%",
			marginRight: whitespace[2],
		}} />
		{language}
	</div>;
};
