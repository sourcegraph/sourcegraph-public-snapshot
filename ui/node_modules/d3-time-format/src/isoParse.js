import locale from "./locale/en-US";
import {isoSpecifier} from "./isoFormat";

function parseIsoNative(string) {
  var date = new Date(string);
  return isNaN(date) ? null : date;
}

var parseIso = +new Date("2000-01-01T00:00:00.000Z")
    ? parseIsoNative
    : locale.utcParse(isoSpecifier);

export default parseIso;
