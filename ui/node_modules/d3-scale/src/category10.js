import colors from "./colors";
import ordinal from "./ordinal";

var colors10 = colors("1f77b4ff7f0e2ca02cd627289467bd8c564be377c27f7f7fbcbd2217becf");

export default function() {
  return ordinal().range(colors10);
}
