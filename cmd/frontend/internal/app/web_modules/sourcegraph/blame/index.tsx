import "app/blame/dom";
import { fetchBlameFile } from "sourcegraph/backend";
import { addHunks, BlameContext, setBlame, store } from "sourcegraph/blame/store";
import * as types from "sourcegraph/util/types";

export function triggerBlame(ctx: BlameContext): void {
	setBlame({ ...store.getValue(), context: ctx, displayLoading: false });

	// Fetch the data.
	fetchBlameFile(ctx.repoURI, ctx.rev, ctx.path, ctx.line, ctx.line).then((hunks: types.Hunk[]) => {
		if (!hunks) {
			return;
		}
		addHunks(ctx, hunks);
	});

	// After 250ms, if there is no data, the component will display a loading
	// indicator.
	setTimeout(() => {
		setBlame({ ...store.getValue(), displayLoading: true });
	}, 250);
}
