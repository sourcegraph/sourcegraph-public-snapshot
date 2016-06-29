// @flow weak

import Store from "sourcegraph/Store";
import Dispatcher from "sourcegraph/Dispatcher";
import deepFreeze from "sourcegraph/util/deepFreeze";
import * as DeltaActions from "sourcegraph/delta/DeltaActions";
import "sourcegraph/delta/DeltaBackend";
import type {DeltaFiles} from "sourcegraph/delta";

function keyFor(baseRepo: number, baseRev: string, headRepo: number, headRev: string): string {
	return `${baseRepo}#${baseRev}#${headRepo}#${headRev}`;
}

export class DeltaStore extends Store {
	reset(data?: {files: {content: DeltaFiles}}) {
		this.files = deepFreeze({
			content: data && data.files ? data.files.content : {},
			get(baseRepo: number, baseRev: string, headRepo: number, headRev: string) {
				return this.content[keyFor(baseRepo, baseRev, headRepo, headRev)] || null;
			},
		});
	}

	toJSON(): any {
		return {
			files: this.files,
		};
	}

	__onDispatch(action) {
		if (action instanceof DeltaActions.FetchedFiles) {
			this.files = deepFreeze({...this.files,
				content: {...this.files.content,
					[keyFor(action.baseRepo, action.baseRev, action.headRepo, action.headRev)]: action.data,
				},
			});
			this.__emitChange();
			return;
		}
	}
}

export default new DeltaStore(Dispatcher.Stores);
