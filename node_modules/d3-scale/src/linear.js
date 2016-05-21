import {ticks, tickStep} from "d3-array";
import {interpolateNumber as reinterpolate} from "d3-interpolate";
import {default as continuous, copy, deinterpolateLinear as deinterpolate} from "./continuous";
import tickFormat from "./tickFormat";

export function linearish(scale) {
  var domain = scale.domain;

  scale.ticks = function(count) {
    var d = domain();
    return ticks(d[0], d[d.length - 1], count == null ? 10 : count);
  };

  scale.tickFormat = function(count, specifier) {
    return tickFormat(domain(), count, specifier);
  };

  scale.nice = function(count) {
    var d = domain(),
        i = d.length - 1,
        n = count == null ? 10 : count,
        start = d[0],
        stop = d[i],
        step = tickStep(start, stop, n);

    if (step) {
      step = tickStep(Math.floor(start / step) * step, Math.ceil(stop / step) * step, n);
      d[0] = Math.floor(start / step) * step;
      d[i] = Math.ceil(stop / step) * step;
      domain(d);
    }

    return scale;
  };

  return scale;
}

export default function linear() {
  var scale = continuous(deinterpolate, reinterpolate);

  scale.copy = function() {
    return copy(scale, linear());
  };

  return linearish(scale);
}
