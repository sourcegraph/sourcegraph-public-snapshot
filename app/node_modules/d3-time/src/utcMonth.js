import interval from "./interval";

export default interval(function(date) {
  date.setUTCHours(0, 0, 0, 0);
  date.setUTCDate(1);
}, function(date, step) {
  date.setUTCMonth(date.getUTCMonth() + step);
}, function(start, end) {
  return end.getUTCMonth() - start.getUTCMonth() + (end.getUTCFullYear() - start.getUTCFullYear()) * 12;
}, function(date) {
  return date.getUTCMonth();
});
