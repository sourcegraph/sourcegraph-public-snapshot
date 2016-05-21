import {color} from "d3-color";
import rgb from "./rgb";
import array from "./array";
import number from "./number";
import object from "./object";
import string from "./string";
import constant from "./constant";

export default function(a, b) {
  var t = typeof b, c;
  return b == null || t === "boolean" ? constant(b)
      : (t === "number" ? number
      : t === "string" ? ((c = color(b)) ? (b = c, rgb) : string)
      : b instanceof color ? rgb
      : Array.isArray(b) ? array
      : object)(a, b);
}
