import interval from "./interval";
import {day} from "./duration";

export default interval(function(date) {
  date.setUTCHours(0, 0, 0, 0);
}, function(date, step) {
  date.setUTCDate(date.getUTCDate() + step);
}, function(start, end) {
  return (end - start) / day;
}, function(date) {
  return date.getUTCDate() - 1;
});
