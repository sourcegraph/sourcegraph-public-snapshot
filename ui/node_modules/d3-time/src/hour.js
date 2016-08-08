import interval from "./interval";
import {hour, minute} from "./duration";

export default interval(function(date) {
  var offset = date.getTimezoneOffset() * minute % hour;
  if (offset < 0) offset += hour;
  date.setTime(Math.floor((+date - offset) / hour) * hour + offset);
}, function(date, step) {
  date.setTime(+date + step * hour);
}, function(start, end) {
  return (end - start) / hour;
}, function(date) {
  return date.getHours();
});
