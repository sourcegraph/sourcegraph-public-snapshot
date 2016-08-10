// tslint:disable: typedef ordered-imports

import * as utf8 from "utf8";

export type Annotation = {
	StartByte: number;
	EndByte: number;
	URL?: string;
	URLs?: string[];
	Class?: string;
	WantInner?: number;
};

// sortAnns sorts annotations *in place* by start position, breaking ties by preferring
// longer annotations.
//
// This must be 100% deterministic because it is run on both the client and server,
// and the order must be the same.
export function sortAnns(anns) {
	return anns.sort((a, b) => {
		if (a.StartByte < b.StartByte || (a.StartByte === b.StartByte && a.EndByte > b.EndByte)) {
			return -1;
		} else if (a.StartByte === b.StartByte && a.EndByte === b.EndByte) {
			return (a.WantInner || 0) - (b.WantInner || 0);
		}
		return 1;
	});
}

// annotate annotates text with the specified annotations. It calls render
// for each annotation. If annotations are overlapping, render may be called
// more or less than once for a single annotation. (It tries to merge simple
// annotations and repeats annotations if nesting makes it necessary.)
//
// Assumes that anns is sorted (using this module's sortAnns function).
//
// Adapted from https://github.com/sourcegraph/annotate/blob/master/annotate.go.
export function annotate(text: string, startByte: number, anns: Annotation[], render) {

	let utf = utf8.encode(text);

	let out: any = [[]];

	// Keep a stack of open annotations (i.e., that have been opened and not
	// yet closed.
	let open: any = [];

	for (let b0 = 0; b0 < utf.length; b0++) {
		let b = b0 + startByte;
		// Open annotations that begin here.
		for (let i = 0; i < anns.length; i++) {
			let a = anns[i];
			if ((b === startByte && a.StartByte < startByte) || a.StartByte === b) {
				// if (a.StartByte < b) throw new Error("start byte out of bounds");
				if (a.EndByte === b) {
					out[0].push(render(a));
				} else {
					// Put this annotation on the stack of annotations that will need
					// to be closed. We remove it from anns at the end of the loop
					// (to avoid modifying anns while we're iterating over it).
					out.unshift([]);
					open.push(a);
				}
			} else if (a.StartByte > b) {
				// Skip past all annotations that we opened (we already put them on the
				// stack of annotations that will need to be closed).
				anns = anns.slice(i);
				break;
			}
		}

		// Just append to the existing string if the last item is a string.
		if (typeof utf[b0] === "undefined") {
			throw new Error("undefined text");
		}
		if (typeof out[0][out[0].length - 1] === "string") {
			out[0][out[0].length - 1] += utf[b0];
		} else {
			out[0].push(utf[b0]);
		}

		// Close annotations that end after this byte, handling overlapping
		// elements as described below. Elements of open are ordered by their
		// annotation start position.

		// We need to close all annotatations ending after this byte, as well as
		// annotations that overlap this annotation's end and should reopen
		// after it closes.
		let toClose: any[] = [];

		// Find annotations ending after this byte.
		let minStart = 0; // start of the leftmost annotation closing here
		for (let i = open.length - 1; i >= 0; i--) {
			let a = open[i];
			if (a.EndByte === b + 1) {
				toClose.push(a);
				if (minStart === 0 || a.StartByte < minStart) {
					minStart = a.StartByte;
				}
				open.splice(i, 1);
			}
		}

		// Find annotations that overlap annotations closing after this and
		// that should reopen after it closes.
		if (toClose.length) {
			for (let i = open.length - 1; i >= 0; i--) {
				let a = open[i];
				if (a.StartByte > minStart) {
					let v = render(a, out.shift());
					out[0].push(v);
				}
			}
		}

		if (toClose.length) {
			for (let i = 0; i < toClose.length; i++) {
				let a = toClose[i];
				let v = render(a, out.shift());
				out[0].push(v);
			}
		}

		if (toClose.length) {
			for (let i = 0; i < open.length; i++) {
				let a = open[i];
				if (a.StartByte > minStart) {
					out.unshift([]);
				}
			}
		}
	}

	if (open.length) {
		// throw new Error("end byte out of bounds");
		// Clean up by closing unclosed annotations, in the order they
		// would have been closed in.
		while (open.length > 0) {
			let a = open.pop();
			let c = out.shift();
			if (c && (typeof c !== "object" || c.length > 0)) {
				let v = render(a, c);
				out[0].push(v);
			}
		}
	}
	if (out.length !== 1) {
		throw new Error("remaining stack");
	}

	return out[0];
}
