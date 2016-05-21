import {cubehelix} from "d3-color";
import {interpolateCubehelixLong} from "d3-interpolate";
import sequential from "./sequential";

export function warm() {
  return sequential(interpolateCubehelixLong(cubehelix(-100, 0.75, 0.35), cubehelix(80, 1.50, 0.8)));
}

export function cool() {
  return sequential(interpolateCubehelixLong(cubehelix(260, 0.75, 0.35), cubehelix(80, 1.50, 0.8)));
}

export default function() {
  var rainbow = cubehelix();
  return sequential(function(t) {
    if (t < 0 || t > 1) t -= Math.floor(t);
    var ts = Math.abs(t - 0.5);
    rainbow.h = 360 * t - 100;
    rainbow.s = 1.5 - 1.5 * ts;
    rainbow.l = 0.8 - 0.9 * ts;
    return rainbow + "";
  });
}
