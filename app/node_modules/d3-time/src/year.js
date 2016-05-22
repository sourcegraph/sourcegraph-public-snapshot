import interval from "./interval";

export default interval(function(date) {
  date.setHours(0, 0, 0, 0);
  date.setMonth(0, 1);
}, function(date, step) {
  date.setFullYear(date.getFullYear() + step);
}, function(start, end) {
  return end.getFullYear() - start.getFullYear();
}, function(date) {
  return date.getFullYear();
});
