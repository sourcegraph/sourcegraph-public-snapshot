import * as differenceInDays from "date-fns/difference_in_days";
import * as distanceInWordsToNow from "date-fns/distance_in_words_to_now";
import * as format from "date-fns/format";

/**
	* TimeFromNow returns a human readable string given an input date.
	* A common way of displaying time is handled by moment#fromNow.
	* This is sometimes called timeago or relative time.
	* @param {string} date Date to be formatted
	* @param {string} format Input date format. e.g: YYY-MM-DDThh:mmTZD
 */
export function timeFromNow(date: string): string {
	return `${distanceInWordsToNow(date)} ago`;
}

export function timeFromNowUntil(date: string, dayToChange: number, dateFormat?: string): string {
	const daysFromNow = differenceInDays(Date.now(), date);
	if (daysFromNow < dayToChange) {
		return timeFromNow(date);
	}

	return format(date, dateFormat || "MMM D, YYYY");
};
