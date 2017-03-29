import { after, before, media, merge } from "glamor";
import { Features } from "sourcegraph/util/features";

export const breakpoints = {
	notSm: "screen and (min-width: 48em)",
	sm: "screen and (max-width: 48em)",
	md: "screen and (min-width: 48em) and (max-width: 64em)",
	lg: "screen and (min-width: 64em)",
};

export const hide = {
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

export const center = { margin: "auto" };
export const container = {
	lg: { ...center, maxWidth: 1280 },
	sm: { ...center, maxWidth: 920 },
};

export const editorToolbarHeight = 45;
export const editorCommitInfoBarHeight = 38;
export const EDITOR_TITLE_HEIGHT = Features.commitInfoBar.isEnabled()
	? editorToolbarHeight + editorCommitInfoBarHeight
	: editorToolbarHeight;

export const flexItem = {
	autoSize: { flex: "1 1 auto" },
	autoGrow: { flex: "1 0 auto" },
	autoShrink: { flex: "0 1 auto" },
};
