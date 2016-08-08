export default function(interpolate, n) {
  var samples = new Array(n);
  for (var i = 0; i < n; ++i) samples[i] = interpolate(i / (n - 1));
  return samples;
}
