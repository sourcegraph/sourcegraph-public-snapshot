import interval from "./interval";
import {minute} from "./duration";

export default interval(function(date) {
  date.setTime(Math.floor(date / minute) * minute);
}, function(date, step) {
  date.setTime(+date + step * minute);
}, function(start, end) {
  return (end - start) / minute;
}, function(date) {
  return date.getMinutes();
});
