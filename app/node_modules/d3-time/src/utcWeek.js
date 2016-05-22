import interval from "./interval";
import {week} from "./duration";

function utcWeekday(i) {
  return interval(function(date) {
    date.setUTCHours(0, 0, 0, 0);
    date.setUTCDate(date.getUTCDate() - (date.getUTCDay() + 7 - i) % 7);
  }, function(date, step) {
    date.setUTCDate(date.getUTCDate() + step * 7);
  }, function(start, end) {
    return (end - start) / week;
  });
}

export var utcSunday = utcWeekday(0);
export var utcMonday = utcWeekday(1);
export var utcTuesday = utcWeekday(2);
export var utcWednesday = utcWeekday(3);
export var utcThursday = utcWeekday(4);
export var utcFriday = utcWeekday(5);
export var utcSaturday = utcWeekday(6);
export default utcSunday;
