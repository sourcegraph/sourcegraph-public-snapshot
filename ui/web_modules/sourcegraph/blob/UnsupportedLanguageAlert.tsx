import * as React from "react";
import { Label } from "sourcegraph/components";

interface Props {
	ext: string | null;
	style?: React.CSSProperties;
}

const sx = { fontWeight: "bold" };

export function UnsupportedLanguageAlert({ ext, style }: Props): JSX.Element {
	const text = ext ? `.${ext} files not supported` : "This file is not supported";
	return <Label color="yellow" style={Object.assign({}, sx, style)} text={text} icon="Warning" />;
};
