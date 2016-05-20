import {rgb} from "d3-color";
import {gamma as interpolateGamma} from "./color";

export default (function gamma(y) {
  var interpolateColor = interpolateGamma(y);

  function interpolateRgb(start, end) {
    var r = interpolateColor((start = rgb(start)).r, (end = rgb(end)).r),
        g = interpolateColor(start.g, end.g),
        b = interpolateColor(start.b, end.b),
        opacity = interpolateColor(start.opacity, end.opacity);
    return function(t) {
      start.r = r(t);
      start.g = g(t);
      start.b = b(t);
      start.opacity = opacity(t);
      return start + "";
    };
  }

  interpolateRgb.gamma = gamma;

  return interpolateRgb;
})(1);
