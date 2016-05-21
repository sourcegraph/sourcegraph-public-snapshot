var rad2deg = 180 / Math.PI;

export var identity = {
  translateX: 0,
  translateY: 0,
  rotate: 0,
  skewX: 0,
  scaleX: 1,
  scaleY: 1
};

export default function(a, b, c, d, e, f) {
  if (a * d === b * c) return null;

  var scaleX = Math.sqrt(a * a + b * b);
  a /= scaleX, b /= scaleX;

  var skewX = a * c + b * d;
  c -= a * skewX, d -= b * skewX;

  var scaleY = Math.sqrt(c * c + d * d);
  c /= scaleY, d /= scaleY, skewX /= scaleY;

  if (a * d < b * c) a = -a, b = -b, skewX = -skewX, scaleX = -scaleX;

  return {
    translateX: e,
    translateY: f,
    rotate: Math.atan2(b, a) * rad2deg,
    skewX: Math.atan(skewX) * rad2deg,
    scaleX: scaleX,
    scaleY: scaleY
  };
}
