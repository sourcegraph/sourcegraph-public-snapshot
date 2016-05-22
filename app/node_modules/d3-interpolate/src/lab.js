import {lab} from "d3-color";
import interpolateColor from "./color";

export default function interpolateLab(start, end) {
  var l = interpolateColor((start = lab(start)).l, (end = lab(end)).l),
      a = interpolateColor(start.a, end.a),
      b = interpolateColor(start.b, end.b),
      opacity = interpolateColor(start.opacity, end.opacity);
  return function(t) {
    start.l = l(t);
    start.a = a(t);
    start.b = b(t);
    start.opacity = opacity(t);
    return start + "";
  };
}
