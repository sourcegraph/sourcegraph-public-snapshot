export {default as timeInterval} from "./src/interval";

import millisecond from "./src/millisecond";
import second from "./src/second";
import minute from "./src/minute";
import hour from "./src/hour";
import day from "./src/day";
import {default as week, sunday, monday, tuesday, wednesday, thursday, friday, saturday} from "./src/week";
import month from "./src/month";
import year from "./src/year";

import utcMinute from "./src/utcMinute";
import utcHour from "./src/utcHour";
import utcDay from "./src/utcDay";
import {default as utcWeek, utcSunday, utcMonday, utcTuesday, utcWednesday, utcThursday, utcFriday, utcSaturday} from "./src/utcWeek";
import utcMonth from "./src/utcMonth";
import utcYear from "./src/utcYear";

export var timeMilliseconds = millisecond.range;
export var timeSeconds = second.range;
export var timeMinutes = minute.range;
export var timeHours = hour.range;
export var timeDays = day.range;
export var timeSundays = sunday.range;
export var timeMondays = monday.range;
export var timeTuesdays = tuesday.range;
export var timeWednesdays = wednesday.range;
export var timeThursdays = thursday.range;
export var timeFridays = friday.range;
export var timeSaturdays = saturday.range;
export var timeWeeks = week.range;
export var timeMonths = month.range;
export var timeYears = year.range;

export var utcMillisecond = millisecond;
export var utcMilliseconds = timeMilliseconds;
export var utcSecond = second;
export var utcSeconds = timeSeconds;
export var utcMinutes = utcMinute.range;
export var utcHours = utcHour.range;
export var utcDays = utcDay.range;
export var utcSundays = utcSunday.range;
export var utcMondays = utcMonday.range;
export var utcTuesdays = utcTuesday.range;
export var utcWednesdays = utcWednesday.range;
export var utcThursdays = utcThursday.range;
export var utcFridays = utcFriday.range;
export var utcSaturdays = utcSaturday.range;
export var utcWeeks = utcWeek.range;
export var utcMonths = utcMonth.range;
export var utcYears = utcYear.range;

export {
  millisecond as timeMillisecond,
  second as timeSecond,
  minute as timeMinute,
  hour as timeHour,
  day as timeDay,
  sunday as timeSunday,
  monday as timeMonday,
  tuesday as timeTuesday,
  wednesday as timeWednesday,
  thursday as timeThursday,
  friday as timeFriday,
  saturday as timeSaturday,
  week as timeWeek,
  month as timeMonth,
  year as timeYear,
  utcSecond,
  utcMinute,
  utcHour,
  utcDay,
  utcSunday,
  utcMonday,
  utcTuesday,
  utcWednesday,
  utcThursday,
  utcFriday,
  utcSaturday,
  utcWeek,
  utcMonth,
  utcYear
};
