import value from "./value";

// TODO sparse arrays?
export default function(a, b) {
  var x = [],
      c = [],
      na = a ? a.length : 0,
      nb = b ? b.length : 0,
      n0 = Math.min(na, nb),
      i;

  for (i = 0; i < n0; ++i) x.push(value(a[i], b[i]));
  for (; i < na; ++i) c[i] = a[i];
  for (; i < nb; ++i) c[i] = b[i];

  return function(t) {
    for (i = 0; i < n0; ++i) c[i] = x[i](t);
    return c;
  };
}
