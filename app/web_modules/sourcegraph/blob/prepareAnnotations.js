import {sortAnns} from "sourcegraph/blob/Annotations";

// prepareAnnotations should be called on annotations received from the server
// to prepare them in ways described below for presentation in the UI.
export default function prepareAnnotations(anns) {
	// Ensure that syntax highlighting is the innermost annotation so
	// that the CSS colors are applied (otherwise ref links appear in
	// the normal link color).
	anns.forEach((a) => {
		if (!a.URL) a.WantInner = 1;
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
					ann.URLs = (ann.URLs || [ann.URL]).concat(ann2.URL);

					// Sort for determinism.
					ann.URLs.sort((a, b) => {
						if (a < b) return -1;
						if (a > b) return 1;
						return 0;
					});

					delete ann.URL;
					anns.splice(j, 1); // Delete the coincident ref.
				}
			} else {
				break;
			}
		}
	}

	return anns;
}
