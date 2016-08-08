import curveLinear from "./curve/linear";
import curveRadial from "./curve/radial";
import area from "./area";

export default function() {
  var a = area(),
      c = a.curve;

  a.angle = a.x, delete a.x;
  a.startAngle = a.x0, delete a.x0;
  a.endAngle = a.x1, delete a.x1;
  a.radius = a.y, delete a.y;
  a.innerRadius = a.y0, delete a.y0;
  a.outerRadius = a.y1, delete a.y1;

  a.curve = function(_) {
    return arguments.length ? c(curveRadial(_)) : c()._curve;
  };

  return a.curve(curveLinear);
}
