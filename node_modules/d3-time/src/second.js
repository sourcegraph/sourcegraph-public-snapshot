import interval from "./interval";
import {second} from "./duration";

export default interval(function(date) {
  date.setTime(Math.floor(date / second) * second);
}, function(date, step) {
  date.setTime(+date + step * second);
}, function(start, end) {
  return (end - start) / second;
}, function(date) {
  return date.getUTCSeconds();
});
