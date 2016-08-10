// tslint:disable: typedef ordered-imports

// LICENSE
//
//   This software is dual-licensed to the public domain and under the following
//   license: you are granted a perpetual, irrevocable license to copy, modify,
//   publish, and distribute this file as you see fit.
//
//   Original from https://github.com/forrestthewoods/lib_fts

// Returns [score, str], Mimics the behavior of Sublime Text search.
// Higher score means better match. Value has no intrinsic meaning.
// Range varies with pattern. Can only compare scores with same search pattern.
export function fuzzy_score(pattern: string, str: string): Array<any> {
	// Score consts
	let adjacencyBonus = 5; // bonus for adjacent matches
	let separatorBonus = 10; // bonus if match occurs after a separator
	let camelBonus = 10; // bonus if match is uppercase and prev is lower
	let leadingLetterPenalty = -3; // penalty applied for every letter in str before the first match
	let maxLeadingLetterPenalty = -9; // maximum penalty for leading letters
	let unmatchedLetterPenalty = -1; // penalty for every letter that doesn't matter

	// Loop letiables
	let score = 0;
	let patternIdx = 0;
	let patternLength = pattern.length;
	let strIdx = 0;
	let strLength = str.length;
	let prevMatched = false;
	let prevLower = false;
	let prevSeparator = true; // true so if first letter match gets separator bonus

	// Use "best" matched letter if multiple string letters match the pattern
	let bestLetter: string | null = null;
	let bestLower: string | null = null;
	let bestLetterIdx: number | null = null;
	let bestLetterScore = 0;

	let matchedIndices: (number | null)[] = [];

	// Loop over strings
	while (strIdx !== strLength) {
		let patternChar = patternIdx !== patternLength ? pattern.charAt(patternIdx) : null;
		let strChar = str.charAt(strIdx);

		let patternLower = patternChar !== null ? patternChar.toLowerCase() : null;
		let strLower = strChar.toLowerCase();
		let strUpper = strChar.toUpperCase();

		let nextMatch = patternChar && patternLower === strLower;
		let rematch = bestLetter && bestLower === strLower;

		let advanced = nextMatch && bestLetter;
		let patternRepeat = bestLetter && patternChar && bestLower === patternLower;
		if (advanced || patternRepeat) {
			score += bestLetterScore;
			matchedIndices.push(bestLetterIdx);
			bestLetter = null;
			bestLower = null;
			bestLetterIdx = null;
			bestLetterScore = 0;
		}

		if (nextMatch || rematch) {
			let newScore = 0;

			// Apply penalty for each letter before the first pattern match
			// Note: std::max because penalties are negative values. So max is smallest penalty.
			if (patternIdx === 0) {
				let penalty = Math.max(strIdx * leadingLetterPenalty, maxLeadingLetterPenalty);
				score += penalty;
			}

			// Apply bonus for consecutive bonuses
			if (prevMatched) {
				newScore += adjacencyBonus;
			}

			// Apply bonus for matches after a separator
			if (prevSeparator) {
				newScore += separatorBonus;
			}

			// Apply bonus across camel case boundaries. Includes "clever" isLetter check.
			if (prevLower && strChar === strUpper && strLower !== strUpper) {
				newScore += camelBonus;
			}

			// Update patter index IFF the next pattern letter was matched
			if (nextMatch) {
				++patternIdx;
			}

			// Update best letter in str which may be for a "next" letter or a "rematch"
			if (newScore >= bestLetterScore) {

				// Apply penalty for now skipped letter
				if (bestLetter !== null) {
					score += unmatchedLetterPenalty;
				}

				bestLetter = strChar;
				bestLower = bestLetter.toLowerCase();
				bestLetterIdx = strIdx;
				bestLetterScore = newScore;
			}

			prevMatched = true;
		} else {
			score += unmatchedLetterPenalty;
			prevMatched = false;
		}

		// Includes "clever" isLetter check.
		prevLower = strChar === strLower && strLower !== strUpper;
		prevSeparator = strChar === "_" || strChar === " ";

		++strIdx;
	}

	// Apply score for last match
	if (bestLetter) {
		score += bestLetterScore;
		matchedIndices.push(bestLetterIdx);
	}

	return [score, str];
}
