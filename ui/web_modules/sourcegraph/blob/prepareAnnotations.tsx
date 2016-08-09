// tslint:disable

import {Annotation} from "sourcegraph/blob/Annotations";
import {sortAnns} from "sourcegraph/blob/Annotations";

// prepareAnnotations should be called on annotations added on the client side
// to prepare them in ways described below for presentation in the UI.
//
// NOTE: This logic must be kept in sync with annotations.Prepare in Go.
export function prepareAnnotations(anns: Array<Annotation>): Array<Annotation> {
	if (anns.length === 0) {
		return anns;
	}

	// Ensure that syntax highlighting is the innermost annotation so
	// that the CSS colors are applied (otherwise ref links appear in
	// the normal link color).
	anns.forEach((a) => {
		if (!a.URL && (!a.URLs || a.URLs.length === 0)) a.WantInner = 1;
		if (!a.StartByte) a.StartByte = 0;
	});

	sortAnns(anns);

	// Condense coincident refs ("multiple defs", such as an embedded Go
	// field's ref to both the field def and the type def).
	for (let i = 0; i < anns.length; i++) {
		const ann = anns[i];
		for (let j = i + 1; j < anns.length; j++) {
			const ann2 = anns[j];
			if (ann.StartByte === ann2.StartByte && ann.EndByte === ann2.EndByte) {
				if ((ann.URLs || ann.URL) && ann2.URL) {
					if (ann.URLs) ann.URLs = ann.URLs.concat(ann2.URL);
					else if (ann.URL) ann.URLs = [ann.URL].concat(ann2.URL);
					else ann.URLs = ([] as any[]).concat(ann2.URL);

					// Sort for determinism.
					ann.URLs.sort((a, b) => {
						if (a < b) return -1;
						if (a > b) return 1;
						return 0;
					});

					delete ann.URL;
					anns.splice(j, 1); // Delete the coincident ref.
					j--;
				}
			} else {
				break;
			}
		}
	}

	return anns;
}
