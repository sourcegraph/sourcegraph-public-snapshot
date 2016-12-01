import * as isEmpty from "lodash/isEmpty";
import { IKeyboardEvent } from "vs/base/browser/keyboardEvent";
import { TPromise } from "vs/base/common/winjs.base";
import { IDataSource, ITree } from "vs/base/parts/tree/browser/tree";
import { DefaultController, ICancelableEvent, LegacyRenderer } from "vs/base/parts/tree/browser/treeDefaults";
import { Tree } from "vs/base/parts/tree/browser/treeImpl";

export type Node = {
	label: string;
	id: string;
	children: Node[];
	parent?: Node;
}

export function makeTree(files: GQL.IFile[]): Node {
	let root = { label: "", id: "", children: [] };

	files.forEach(file => {
		const parts = file.name.split("/");
		merge(root, parts);
	});

	return root;
}

export function nodePathFromPath(root: Node, path: string): Node[] {
	const parts = path.split("/");
	let cur = root;
	let nodes = [root];
	parts.forEach(name => {
		const next = cur.children.find(x => x.label === name);
		if (!next) {
			console.error(`${path} not found in tree `, root);
			return [];
		}
		cur = next;
		nodes.push(cur);
	});
	return nodes;
}

function merge(root: Node, parts: string[]): void {
	let cur = root;
	parts.forEach(name => {
		// O(files in dir)
		let next = cur.children.find(x => x.label === name);
		if (!next) {
			next = { label: name, id: cur.id + "/" + name, children: [], parent: cur };
			next.toString = () => name;
			cur.children.push(next);
		}
		cur = next;
	});
}

export class FileTreeDataSource implements IDataSource {

	getId(tree: ITree, node: Node): string {
		return node.id;
	}

	hasChildren(tree: ITree, node: Node): boolean {
		return !isEmpty(node.children);
	}

	getChildren(tree: ITree, node: Node): TPromise<Node[]> {
		return TPromise.wrap(node.children);
	}

	getParent(tree: ITree, node: Node): TPromise<Node | undefined> {
		return TPromise.wrap(node.parent);
	}

}

// Controller overrides the default left click behavior to push a new URL using
// React Router, when the target element is a file.
export class Controller extends DefaultController {

	defaultLeftClick: (tree: Tree, element: Node, eventish: ICancelableEvent, origin?: string) => boolean;
	defaultOnEnter: (tree: Tree, event: IKeyboardEvent) => boolean;
	selectElement: (node: Node) => boolean;

	constructor(selectElement: (node: Node) => boolean) {
		super();
		this.defaultLeftClick = super.onLeftClick;
		this.defaultOnEnter = super.onEnter;
		this.selectElement = selectElement;
	}

	onLeftClick(tree: Tree, element: Node, eventish: ICancelableEvent, origin: string = "mouse"): boolean {
		this.selectElement(element);
		return this.defaultLeftClick(tree, element, eventish, origin);
	}

	onEnter(tree: Tree, event: IKeyboardEvent): boolean {
		if (!this.defaultOnEnter(tree, event)) {
			return false;
		}
		const sel = tree.getSelection();
		return sel.length === 1 && this.selectElement(sel[0]);
	}

}

export class Renderer extends LegacyRenderer {
	getHeight(): number {
		return 30;
	}
}
