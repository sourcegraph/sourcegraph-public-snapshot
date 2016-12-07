import { after, before, media, merge } from "glamor";

export const breakpoints = {
	notSm: "screen and (min-width: 48em)",
	sm: "screen and (max-width: 48em)",
	md: "screen and (min-width: 48em) and (max-width: 64em)",
	lg: "screen and (min-width: 64em)",
};

const hide = {
	notSm: media(breakpoints.notSm, { display: "none" }),
	sm: media(breakpoints.sm, { display: "none !important" }),
	md: media(breakpoints.md, { display: "none !important" }),
	lg: media(breakpoints.lg, { display: "none !important" }),
};

export const clearFix = merge(
	{ zoom: "1" },
	before({
		content: "\"\"",
		display: "table",
	}),
	after({
		content: "\"\"",
		display: "table",
		clear: "both",
	})
);

export const container = {
	margin: "auto",
	maxWidth: 1280,
};

const editorToolbarHeight = 65;

export const layout = {
	breakpoints,
	clearFix,
	container,
	editorToolbarHeight,
	hide,
};
