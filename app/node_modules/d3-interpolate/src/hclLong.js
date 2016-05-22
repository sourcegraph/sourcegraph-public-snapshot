import {hcl} from "d3-color";
import interpolateColor from "./color";

export default function interpolateHclLong(start, end) {
  var h = interpolateColor((start = hcl(start)).h, (end = hcl(end)).h),
      c = interpolateColor(start.c, end.c),
      l = interpolateColor(start.l, end.l),
      opacity = interpolateColor(start.opacity, end.opacity);
  return function(t) {
    start.h = h(t);
    start.c = c(t);
    start.l = l(t);
    start.opacity = opacity(t);
    return start + "";
  };
}
