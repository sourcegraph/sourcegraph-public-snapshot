import {after, before, merge} from "glamor";

export const breakpoints = {
	"not-sm": "screen and (min-width: 48em)",
	"sm": "screen and (max-width: 48em)",
	"md": "screen and (min-width: 48em) and (max-width: 64em)",
	"lg": "screen and (min-width: 64em)",
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

export const layout = {
	breakpoints,
	clearFix,
};
