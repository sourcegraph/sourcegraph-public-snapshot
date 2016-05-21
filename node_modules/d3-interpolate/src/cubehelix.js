import {cubehelix} from "d3-color";
import interpolateColor, {hue as interpolateHue} from "./color";

export default (function gamma(y) {
  y = +y;

  function interpolateCubehelix(start, end) {
    var h = interpolateHue((start = cubehelix(start)).h, (end = cubehelix(end)).h),
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

  interpolateCubehelix.gamma = gamma;

  return interpolateCubehelix;
})(1);
