import interval from "./interval";
import {day, minute} from "./duration";

export default interval(function(date) {
  date.setHours(0, 0, 0, 0);
}, function(date, step) {
  date.setDate(date.getDate() + step);
}, function(start, end) {
  return (end - start - (end.getTimezoneOffset() - start.getTimezoneOffset()) * minute) / day;
}, function(date) {
  return date.getDate() - 1;
});
