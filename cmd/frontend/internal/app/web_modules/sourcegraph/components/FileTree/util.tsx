export interface TreeNode {
	text: string;
	children: any[] | undefined;
	state: NodeState | undefined;
	id?: string;
}

interface NodeState {
	selected?: boolean;
	opened?: boolean;
}

/**
 * buildFileTree is responsible for taking the graphql response from listFiles and building it into
 * the tree structure. It takes an optional parameter of a path that should be the selected node.
 * @param data Array of file names in the form of {name: "path/to/file" }
 * @param selectedPath The path to the current selected file
 */
export function buildFileTree(data: any[], selectedPath?: string): any[] {
	const input = data.map(file => {
		return `${file.name}`;
	});

	const output = [];
	let k: number = 0;
	let cursor: any | undefined;
	for (let i = 0; i < input.length; i++) {
		const chain: any[] = input[i].split("/");
		let currentNode: TreeNode[] | undefined = output;
		for (let j = 0; j < chain.length; j++) {
			const wantedNode: string = chain[j];
			const lastNode = currentNode;
			if (currentNode) {
				for (k = 0; k < currentNode.length; k++) {
					if (currentNode[k].text === wantedNode && currentNode[k].children !== undefined) {
						currentNode = currentNode[k].children;
						break;
					}
				}
			}
			// If we couldn't find an item in this list of children
			// that has the right name, create one:
			if (lastNode === currentNode) {
				if (chain[chain.length - 1] === wantedNode) {
					const newNode = currentNode![k] = { text: wantedNode, children: undefined, id: input[i], state: {} };
					if (selectedPath && input[i] === selectedPath) {
						cursor = newNode;
						currentNode![k].state = {
							selected: true,
							opened: true,
						};
					}
					currentNode = newNode.children;
				} else {
					const newNode = currentNode![k] = { text: wantedNode, children: [], state: {} };
					currentNode = newNode.children;
				}
			}
		}
	}
	return output;
}
