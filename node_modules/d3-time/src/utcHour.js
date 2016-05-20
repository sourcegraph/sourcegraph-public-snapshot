import interval from "./interval";
import {hour} from "./duration";

export default interval(function(date) {
  date.setUTCMinutes(0, 0, 0);
}, function(date, step) {
  date.setTime(+date + step * hour);
}, function(start, end) {
  return (end - start) / hour;
}, function(date) {
  return date.getUTCHours();
});
