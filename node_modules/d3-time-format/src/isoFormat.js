import locale from "./locale/en-US";

export var isoSpecifier = "%Y-%m-%dT%H:%M:%S.%LZ";

function formatIsoNative(date) {
  return date.toISOString();
}

var formatIso = Date.prototype.toISOString
    ? formatIsoNative
    : locale.utcFormat(isoSpecifier);

export default formatIso;
