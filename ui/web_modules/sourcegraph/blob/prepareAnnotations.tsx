// tslint:disable: typedef ordered-imports

import {Annotation} from "sourcegraph/blob/Annotations";
import {sortAnns} from "sourcegraph/blob/Annotations";
import * as cloneDeep from "lodash/cloneDeep";

// prepareAnnotations should be called on annotations added on the client side
// to prepare them in ways described below for presentation in the UI.
export function prepareAnnotations(anns: Annotation[]): Annotation[] {
	if (anns.length === 0) {
		return anns;
	}

	anns.forEach((a) => {
		if (!a.StartByte) {
			a.StartByte = 0;
		}
	});

	// Ensure that syntax highlighting is the innermost annotation so
	// that the CSS colors are applied (otherwise ref links appear in
	// the normal link color).
	// For each annotation, we create a clone that has its WantInner 
	// field set to zero. This ensures (after sorting), that each token
	// has two annotations:
	// - first: an annotation that sets the color for the syntax highlighting
	// - second: an annotation that will eventually be transformed into a
	// jump-to-def link. The link's color is the same as the color set by 
	// the first annotation. 
	let clones = cloneDeep(anns);
	clones.forEach(a => a.WantInner = 0);
	let wrappedAnns = anns.concat(clones);
	sortAnns(wrappedAnns);
	return wrappedAnns;
}
