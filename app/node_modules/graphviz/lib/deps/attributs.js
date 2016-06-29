var Hash = require('./core_ext/hash').Hash,
    util = require('util');

var attrs = {
  "Damping" :            { "usage" : "G",    "type" : "double" },
  "K" :                  { "usage" : "GC",   "type" : "double" },
  "URL" :                { "usage" : "ENGC", "type" : "escString" },
  "area" :               { "usage" : "NC",   "type" : "double" },
  "arrowhead" :          { "usage" : "E",    "type" : "arrowType" },
  "arrowsize" :          { "usage" : "E",    "type" : "double" },
  "arrowtail" :          { "usage" : "E",    "type" : "arrowType" },
  "aspect" :             { "usage" : "G",    "type" : "aspectType" },
  "bb" :                 { "usage" : "G",    "type" : "rect" },
  "bgcolor" :            { "usage" : "GC",   "type" : "color" },
  "center" :             { "usage" : "G",    "type" : "bool" },
  "charset" :            { "usage" : "G",    "type" : "string" },
  "clusterrank" :        { "usage" : "G",    "type" : "clusterMode" },
  "color" :              { "usage" : "ENC",  "type" : "color" },
  "colorscheme" :        { "usage" : "ENCG", "type" : "string" },
  "comment" :            { "usage" : "ENG",  "type" : "string" },
  "compound" :           { "usage" : "G",    "type" : "bool" },
  "concentrate" :        { "usage" : "G",    "type" : "bool" },
  "constraint" :         { "usage" : "E",    "type" : "bool" },
  "decorate" :           { "usage" : "E",    "type" : "bool" },
  "defaultdist" :        { "usage" : "G",    "type" : "double" },
  "dim" :                { "usage" : "G",    "type" : "int" },
  "dimen" :              { "usage" : "G",    "type" : "int" },
  "dir" :                { "usage" : "E",    "type" : "dirType" },
  "diredgeconstraints" : { "usage" : "G",    "type" : "string" },
  "distortion" :         { "usage" : "N",    "type" : "double" },
  "dpi" :                { "usage" : "G",    "type" : "double" },
  "edgeURL" :            { "usage" : "E",    "type" : "escString" },
  "edgehref" :           { "usage" : "E",    "type" : "escString" },
  "edgetarget" :         { "usage" : "E",    "type" : "escString" },
  "edgetooltip" :        { "usage" : "E",    "type" : "escString" },
  "epsilon" :            { "usage" : "G",    "type" : "double" },
  "esep" :               { "usage" : "G",    "type" : "double" },
  "fillcolor" :          { "usage" : "NEC",  "type" : "color" },
  "fixedsize" :          { "usage" : "N",    "type" : "bool" },
  "fontcolor" :          { "usage" : "ENGC", "type" : "color" },
  "fontname" :           { "usage" : "ENGC", "type" : "string" },
  "fontnames" :          { "usage" : "G",    "type" : "string" },
  "fontpath" :           { "usage" : "G",    "type" : "string" },
  "fontsize" :           { "usage" : "ENGC", "type" : "double" },
  "group" :              { "usage" : "N",    "type" : "string" },
  "headURL" :            { "usage" : "E",    "type" : "escString" },
  "headclip" :           { "usage" : "E",    "type" : "bool" },
  "headhref" :           { "usage" : "E",    "type" : "escString" },
  "headlabel" :          { "usage" : "E",    "type" : "lblString" },
  "headport" :           { "usage" : "E",    "type" : "portPos" },
  "headtarget" :         { "usage" : "E",    "type" : "escString" },
  "headtooltip" :        { "usage" : "E",    "type" : "escString" },
  "height" :             { "usage" : "N",    "type" : "double" },
  "href" :               { "usage" : "ENGC", "type" : "escString" },
  "id" :                 { "usage" : "GNE",  "type" : "lblString" },
  "image" :              { "usage" : "N",    "type" : "string" },
  "imagepath" :          { "usage" : "G",    "type" : "string" },
  "imagescale" :         { "usage" : "N",    "type" : "string" },
  "label" :              { "usage" : "ENGC", "type" : "lblString" },
  "labelURL" :           { "usage" : "E",    "type" : "escString" },
  "labelangle" :         { "usage" : "E",    "type" : "double" },
  "labeldistance" :      { "usage" : "E",    "type" : "double" },
  "labelfloat" :         { "usage" : "E",    "type" : "bool" },
  "labelfontcolor" :     { "usage" : "E",    "type" : "color" },
  "labelfontname" :      { "usage" : "E",    "type" : "string" },
  "labelfontsize" :      { "usage" : "E",    "type" : "double" },
  "labelhref" :          { "usage" : "E",    "type" : "escString" },
  "labeljust" :          { "usage" : "GC",   "type" : "string" },
  "labelloc" :           { "usage" : "NGC",  "type" : "string" },
  "labeltarget" :        { "usage" : "E",    "type" : "escString" },
  "labeltooltip" :       { "usage" : "E",    "type" : "escString" },
  "landscape" :          { "usage" : "G",    "type" : "bool" },
  "layer" :              { "usage" : "ENC",  "type" : "layerRange" },
  "layerlistsep" :       { "usage" : "G",    "type" : "string" },
  "layers" :             { "usage" : "G",    "type" : "layerList" },
  "layerselect" :        { "usage" : "G",    "type" : "layerRange" },
  "layersep" :           { "usage" : "G",    "type" : "string" },
  "layout" :             { "usage" : "G",    "type" : "string" },
  "len" :                { "usage" : "E",    "type" : "double" },
  "levels" :             { "usage" : "G",    "type" : "int" },
  "levelsgap" :          { "usage" : "G",    "type" : "double" },
  "lhead" :              { "usage" : "E",    "type" : "string" },
  "lheight" :            { "usage" : "GC",   "type" : "double" },
  "lp" :                 { "usage" : "EGC",  "type" : "point" },
  "ltail" :              { "usage" : "E",    "type" : "string" },
  "lwidth" :             { "usage" : "GC",   "type" : "double" },
  "margin" :             { "usage" : "NGC",  "type" : "pointf" },
  "maxiter" :            { "usage" : "G",    "type" : "int" },
  "mclimit" :            { "usage" : "G",    "type" : "double" },
  "mindist" :            { "usage" : "G",    "type" : "double" },
  "minlen" :             { "usage" : "E",    "type" : "int" },
  "mode" :               { "usage" : "G",    "type" : "string" },
  "model" :              { "usage" : "G",    "type" : "string" },
  "mosek" :              { "usage" : "G",    "type" : "bool" },
  "nodesep" :            { "usage" : "G",    "type" : "double" },
  "nojustify" :          { "usage" : "GCNE", "type" : "bool" },
  "normalize" :          { "usage" : "G",    "type" : "bool" },
  "nslimit" :            { "usage" : "G",    "type" : "double" },
  "nslimit1" :           { "usage" : "G",    "type" : "double" },
  "ordering" :           { "usage" : "GN",   "type" : "string" },
  "orientation" :        { "usage" : "GN",   "type" : "string" },
  "outputorder" :        { "usage" : "G",    "type" : "outputMode" },
  "overlap" :            { "usage" : "G",    "type" : "string" },
  "overlap_scaling" :    { "usage" : "G",    "type" : "double" },
  "pack" :               { "usage" : "G",    "type" : "int" },
  "packmode" :           { "usage" : "G",    "type" : "packMode" },
  "pad" :                { "usage" : "G",    "type" : "pointf" },
  "page" :               { "usage" : "G",    "type" : "pointf" },
  "pagedir" :            { "usage" : "G",    "type" : "pagedir" },
  "pencolor" :           { "usage" : "C",    "type" : "color" },
  "penwidth" :           { "usage" : "CNE",  "type" : "double" },
  "peripheries" :        { "usage" : "NC",   "type" : "int" },
  "pin" :                { "usage" : "N",    "type" : "bool" },
  "pos" :                { "usage" : "EN",   "type" : "point" },
  "quadtree" :           { "usage" : "G",    "type" : "quadType" },
  "quantum" :            { "usage" : "G",    "type" : "double" },
  "rank" :               { "usage" : "S",    "type" : "rankType" },
  "rankdir" :            { "usage" : "G",    "type" : "rankdir" },
  "ranksep" :            { "usage" : "G",    "type" : "double" },
  "ratio" :              { "usage" : "G",    "type" : "string" },
  "rects" :              { "usage" : "N",    "type" : "rect" },
  "regular" :            { "usage" : "N",    "type" : "bool" },
  "remincross" :         { "usage" : "G",    "type" : "bool" },
  "repulsiveforce" :     { "usage" : "G",    "type" : "double" },
  "resolution" :         { "usage" : "G",    "type" : "double" },
  "root" :               { "usage" : "GN",   "type" : "string" },
  "rotate" :             { "usage" : "G",    "type" : "int" },
  "rotation" :           { "usage" : "G",    "type" : "double" },
  "samehead" :           { "usage" : "E",    "type" : "string" },
  "sametail" :           { "usage" : "E",    "type" : "string" },
  "samplepoints" :       { "usage" : "N",    "type" : "int" },
  "scale" :              { "usage" : "G",    "type" : "double" },
  "searchsize" :         { "usage" : "G",    "type" : "int" },
  "sep" :                { "usage" : "G",    "type" : "double" },
  "shape" :              { "usage" : "N",    "type" : "shape" },
  "shapefile" :          { "usage" : "N",    "type" : "string" },
  "showboxes" :          { "usage" : "ENG",  "type" : "int" },
  "sides" :              { "usage" : "N",    "type" : "int" },
  "size" :               { "usage" : "G",    "type" : "pointf" },
  "skew" :               { "usage" : "N",    "type" : "double" },
  "smoothing" :          { "usage" : "G",    "type" : "smoothType" },
  "sortv" :              { "usage" : "GCN",  "type" : "int" },
  "splines" :            { "usage" : "G",    "type" : "string" },
  "start" :              { "usage" : "G",    "type" : "startType" },
  "style" :              { "usage" : "ENCG", "type" : "style" },
  "stylesheet" :         { "usage" : "G",    "type" : "string" },
  "tailURL" :            { "usage" : "E",    "type" : "escString" },
  "tailclip" :           { "usage" : "E",    "type" : "bool" },
  "tailhref" :           { "usage" : "E",    "type" : "escString" },
  "taillabel" :          { "usage" : "E",    "type" : "lblString" },
  "tailport" :           { "usage" : "E",    "type" : "portPos" },
  "tailtarget" :         { "usage" : "E",    "type" : "escString" },
  "tailtooltip" :        { "usage" : "E",    "type" : "escString" },
  "target" :             { "usage" : "ENGC", "type" : "escString" },
  "tooltip" :            { "usage" : "NEC",  "type" : "escString" },
  "truecolor" :          { "usage" : "G",    "type" : "bool" },
  "vertices" :           { "usage" : "N",    "type" : "pointfList" },
  "viewport" :           { "usage" : "G",    "type" : "viewPort" },
  "voro_margin" :        { "usage" : "G",    "type" : "double" },
  "weight" :             { "usage" : "E",    "type" : "double" },
  "width" :              { "usage" : "N",    "type" : "double" },
  "xlabel" :             { "usage" : "EN",   "type" : "lblString" },
  "z" :                  { "usage" : "N",    "type" : "double" }
};

var gType = {
  "E" : "edge",
  "N" : "node",
  "G" : "graph",
  "C" : "cluster"
};

var quotedTypes = [
  "escString",
  "rect",
  "color",
  "colorList",
  "string",
  "lblString",
  "portPos",
  "point",
  "pointf",
  "pointfList",
  "splineType",
  "style",
  "viewPort"
];

function mustBeQuoted(data) {
  return(quotedTypes.indexOf(attrs[data].type) !== -1);
}

function quoteMe(attr, value) {
  if(mustBeQuoted(attr)) {
    return('"' + value + '"');
  } else {
    return(value);
  }
}

function validateAttribut(name, type) {
  if(attrs[name]) {
    return(attrs[name].usage.indexOf(type) > -1);
  } else {
    return(false);
  }
}

exports.isValid = function(name, type) {
	return validateAttribut(name, type);
};

var Attributs = exports.Attributs = function(t) {
  this._type = t;
  this.attributs = new Hash();
};

Attributs.prototype.length = function() {
  return(this.attributs.length);
};

Attributs.prototype.set = function(name, value) {
  if(validateAttribut(name, this._type) === false) {
    util.debug("Warning : Invalid attribut `" + name + "' for a " + gType[this._type]);
    // throw "Invalid attribut `"+name+"' for a "+gType[this._type]
  }
  this.attributs.setItem(name, value);
};

Attributs.prototype.get = function(name) {
  return this.attributs.items[name];
};

Attributs.prototype.to_dot = function(link) {
  var attrsOutput = "",
		sep = "";
  
  if(this.attributs.length > 0) {
    attrsOutput = attrsOutput + " [ ";
    for(var name in this.attributs.items) {
      attrsOutput = attrsOutput + sep + name + " = " + quoteMe(name, this.attributs.items[name]);
      sep = ", ";
    }
    attrsOutput = attrsOutput + " ]";
  }
  
  return attrsOutput;
};
