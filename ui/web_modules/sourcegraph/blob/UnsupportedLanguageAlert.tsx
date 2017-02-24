import * as React from "react";
import { Label } from "sourcegraph/components";

interface Props {
	ext: string | null;
	style?: React.CSSProperties;
	inBeta?: boolean;
}

const sx = { fontWeight: "bold" };

export function UnsupportedLanguageAlert({ ext, style, inBeta }: Props): JSX.Element {
	let text;
	if (inBeta) {
		text = ext ? `.${ext} support is in beta` : "Support for this file is in beta";
	} else {
		text = ext ? `.${ext} files not supported` : "This file is not supported";
	}
	return <Label color={inBeta ? "green" : "yellow"} style={Object.assign({}, sx, style)} text={text} icon="Warning" />;
};
