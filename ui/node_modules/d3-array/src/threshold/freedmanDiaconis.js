import ascending from "../ascending";
import quantile from "../quantile";

export default function(values, min, max) {
  values.sort(ascending);
  return Math.ceil((max - min) / (2 * (quantile(values, 0.75) - quantile(values, 0.25)) * Math.pow(values.length, -1 / 3)));
}
