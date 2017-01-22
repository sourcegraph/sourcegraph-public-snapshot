import { TPromise } from "vs/base/common/winjs.base";
import * as tree from "vs/base/parts/tree/browser/tree";
import { ITextModelResolverService } from "vs/editor/common/services/resolverService";

import { FileReferences, OneReference, ReferencesModel } from "sourcegraph/workbench/info/referencesModel";

export class DataSource implements tree.IDataSource {

	constructor(
		@ITextModelResolverService private _textModelResolverService: ITextModelResolverService
	) {
		//
	}

	public getId(tree: tree.ITree, element: any): string | any {
		if (element instanceof ReferencesModel) {
			return "root";
		} else if (element instanceof FileReferences) {
			return (element).id;
		} else if (element instanceof OneReference) {
			return (element).id;
		}
	}

	public hasChildren(tree: tree.ITree, element: any): boolean | any {
		if (element instanceof ReferencesModel) {
			return true;
		}
		if (element instanceof FileReferences && !(element).failure) {
			if (element.children.length === 0) {
				return false;
			}
			tree.expand(element);
			return true;
		}
	}

	public getChildren(tree: tree.ITree, element: ReferencesModel | FileReferences): TPromise<any[]> {
		if (element instanceof ReferencesModel) {
			return TPromise.as(element.groups);
		} else if (element instanceof FileReferences) {
			return element.resolve(this._textModelResolverService).then(val => {
				if (element.failure) {
					// refresh the element on failure so that
					// we can update its rendering
					return tree.refresh(element).then(() => val.children);
				}
				return val.children;
			});
		} else {
			return TPromise.as([]);
		}
	}

	public getParent(tree: tree.ITree, element: any): TPromise<any> {
		let result: any = null;
		if (element instanceof FileReferences) {
			result = (element).parent;
		} else if (element instanceof OneReference) {
			result = (element).parent;
		}
		return TPromise.as(result);
	}
}
