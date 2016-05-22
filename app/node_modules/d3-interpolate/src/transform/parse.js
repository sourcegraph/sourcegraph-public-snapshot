import decompose, {identity} from "./decompose";

var cssNode,
    cssRoot,
    cssView,
    svgNode;

export function parseCss(value) {
  if (value === "none") return identity;
  if (!cssNode) cssNode = document.createElement("DIV"), cssRoot = document.documentElement, cssView = document.defaultView;
  cssNode.style.transform = value;
  value = cssView.getComputedStyle(cssRoot.appendChild(cssNode), null).getPropertyValue("transform");
  cssRoot.removeChild(cssNode);
  var m = value.slice(7, -1).split(",");
  return decompose(+m[0], +m[1], +m[2], +m[3], +m[4], +m[5]);
}

export function parseSvg(value) {
  if (!svgNode) svgNode = document.createElementNS("http://www.w3.org/2000/svg", "g");
  svgNode.setAttribute("transform", value == null ? "" : value);
  var m = svgNode.transform.baseVal.consolidate().matrix;
  return decompose(m.a, m.b, m.c, m.d, m.e, m.f);
}
