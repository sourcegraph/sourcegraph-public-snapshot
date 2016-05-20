import interval from "./interval";

export default interval(function(date) {
  date.setHours(0, 0, 0, 0);
  date.setDate(1);
}, function(date, step) {
  date.setMonth(date.getMonth() + step);
}, function(start, end) {
  return end.getMonth() - start.getMonth() + (end.getFullYear() - start.getFullYear()) * 12;
}, function(date) {
  return date.getMonth();
});
