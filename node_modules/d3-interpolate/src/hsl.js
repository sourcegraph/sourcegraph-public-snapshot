import {hsl} from "d3-color";
import interpolateColor, {hue as interpolateHue} from "./color";

export default function interpolateHsl(start, end) {
  var h = interpolateHue((start = hsl(start)).h, (end = hsl(end)).h),
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
