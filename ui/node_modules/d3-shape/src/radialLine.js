import curveLinear from "./curve/linear";
import curveRadial from "./curve/radial";
import line from "./line";

export default function() {
  var l = line(),
      c = l.curve;

  l.angle = l.x, delete l.x;
  l.radius = l.y, delete l.y;

  l.curve = function(_) {
    return arguments.length ? c(curveRadial(_)) : c()._curve;
  };

  return l.curve(curveLinear);
}
