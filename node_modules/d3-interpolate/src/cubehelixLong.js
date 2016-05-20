import {cubehelix} from "d3-color";
import interpolateColor from "./color";

export default (function gamma(y) {
  y = +y;

  function interpolateCubehelixLong(start, end) {
    var h = interpolateColor((start = cubehelix(start)).h, (end = cubehelix(end)).h),
        s = interpolateColor(start.s, end.s),
        l = interpolateColor(start.l, end.l),
        opacity = interpolateColor(start.opacity, end.opacity);
    return function(t) {
      start.h = h(t);
      start.s = s(t);
      start.l = l(Math.pow(t, y));
      start.opacity = opacity(t);
      return start + "";
    };
  }

  interpolateCubehelixLong.gamma = gamma;

  return interpolateCubehelixLong;
})(1);
