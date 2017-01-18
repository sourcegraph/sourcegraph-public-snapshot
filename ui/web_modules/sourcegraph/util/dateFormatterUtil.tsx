import * as moment from "moment";

moment.updateLocale("en", {
	relativeTime: {
		future: "in %s",
		past: "%s ago",
		s: "s",
		m: "a minute",
		mm: "%dm",
		h: "an hour",
		hh: "%dh",
		d: "a day",
		dd: "%dd",
		M: "a month",
		MM: "%dmo",
		y: "a year",
		yy: "%dy"
	}
});

const StandardDateFormat: string = "YYYY-MM-DDThh:mmTZD";

export function DateMoment(date: string, format?: string): moment.Moment {
	return moment(date, format || StandardDateFormat);
}

/**
	* TimeFromNow returns a human readable string given an input date.
	* A common way of displaying time is handled by moment#fromNow.
	* This is sometimes called timeago or relative time.
	* @param {string} date Date to be formatted
	* @param {string} format Input date format. e.g: YYY-MM-DDThh:mmTZD
 */
export function TimeFromNow(date: string, format?: string): string {
	return DateMoment(date, format).fromNow();
}

export function TimeFromNowUntil(date: string, dayToChange: number, format?: string): string {
	const dateMoment = DateMoment(date, format);
	const diffInDays = moment().diff(dateMoment, "day");
	if (diffInDays < dayToChange) {
		return TimeFromNow(date);
	}

	return dateMoment.format("MMM D, YYYY");
};
