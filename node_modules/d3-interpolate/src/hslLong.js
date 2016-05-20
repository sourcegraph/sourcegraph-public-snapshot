import {hsl} from "d3-color";
import interpolateColor from "./color";

export default function interpolateHslLong(start, end) {
  var h = interpolateColor((start = hsl(start)).h, (end = hsl(end)).h),
      s = interpolateColor(start.s, end.s),
      l = interpolateColor(start.l, end.l),
      opacity = interpolateColor(start.opacity, end.opacity);
  return function(t) {
    start.h = h(t);
    start.s = s(t);
    start.l = l(t);
    start.opacity = opacity(t);
    return start + "";
  };
}
