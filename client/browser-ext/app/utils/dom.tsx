
/**
 * Inserts an element after the reference node.
 * @param el The element to be rendered.
 * @param referenceNode The node to render the element after.
 */
export function insertAfter(el: HTMLElement, referenceNode: Node): void {
	if (referenceNode.parentNode) {
		referenceNode.parentNode.insertBefore(el, referenceNode.nextSibling);
	}
}
