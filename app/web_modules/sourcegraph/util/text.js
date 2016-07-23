// @flow

// getSnippets transforms txt into a string that showcases the parts
// of txt that match the regex described by matchTerms.
// The returned string conisists of the first "initSize" chars of txt
// (rounded to the location of the next word), followed by the sections of
// of the rest of txt that match matchTerms, padded with the surrouding
// padSize chars (again rounded to the location of the next word), and separated
// by an ellipsis.
export function getSnippets(matchTerms: string, txt: string, initSize: number, padSize: number): string {
	if (!matchTerms || !txt) {
		return txt;
	}
	let lastEndIndex = 0;
	let matcher = new RegExp(matchTerms, "ig");
	let pivot = initSize;
	if (initSize > 0) {
		let nextWordIndex = txt.indexOf(" ", initSize);
		pivot = (nextWordIndex === -1)? txt.length: nextWordIndex + 1;
	}
	// indicies is an array of tuples with the substring, the leftmost character position of the substring,
	// and the rightmost character position of the substring.
	let indicies = [[txt.substring(0, pivot), 0, pivot]];
	let rest = txt.substring(pivot);
	let result = matcher.exec(rest);
	while (result) {
		let leftPadIndex = Math.max(lastEndIndex, leftIndexOf(rest, " ", result.index - padSize - 1));
		let rightPadIndex = rest.indexOf(" ", matcher.lastIndex + padSize) + 1;
		if (rightPadIndex === 0) {
			rightPadIndex = rest.length;
		}
		indicies.push([rest.substring(leftPadIndex, rightPadIndex), pivot + leftPadIndex, pivot + rightPadIndex]);
		lastEndIndex = rightPadIndex;
		result = matcher.exec(rest);
	}
	let snip = indicies.reduce((acc, elem, i, arr) => {
		let [lastStr, lastRight] = acc;
		let [currStr, currLeft, currRight] = elem;
		// If two snippets are separated by more than a space, put an ellipsis between them.
		return [`${lastStr}${(currLeft - lastRight > 1)? "...": " "}${currStr.trim()}`, currRight];
	}, ["", 0])[0];
	if (lastEndIndex !== rest.length) {
		snip += "...";
	}
	snip = snip.trim();
	if (!initSize && indicies.length > 1 && indicies[1][1] !== 0) {
		snip = `...${snip}`;
	}
	return snip;
}

// leftIndexOf is like indexOf for strings, but instead
// it searches backwards from the "start" position.
export function leftIndexOf(str: string, searchTerm: string, start:number = 0): number {
	let reverseStr = str.split("").reverse().join("");
	let reverseStart = Math.abs(str.length - 1 - start);
	let reverseTerm = searchTerm.split("").reverse().join("");
	let revRes = reverseStr.indexOf(reverseTerm, reverseStart);
	if (revRes === -1) {
		return -1;
	}
	return Math.abs(str.length - 1 - revRes) - (searchTerm.length - 1);
}

// escapeRegExp escapes user input "str", and ensures that eveything in "str" is used
// as a literal string when feeding the output to a RegExp constructor.
// Taken from: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Regular_Expressions
export function escapeRegExp(str: string): string {
	return str.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

