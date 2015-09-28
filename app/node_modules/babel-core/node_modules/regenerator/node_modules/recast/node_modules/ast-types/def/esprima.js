require("./es7");

var types = require("../lib/types");
var defaults = require("../lib/shared").defaults;
var def = types.Type.def;
var or = types.Type.or;

def("VariableDeclaration")
    .field("declarations", [or(
        def("VariableDeclarator"),
        def("Identifier") // Esprima deviation.
    )]);

def("Property")
    .field("value", or(
        def("Expression"),
        def("Pattern") // Esprima deviation.
    ));

def("ArrayPattern")
    .field("elements", [or(
        def("Pattern"),
        def("SpreadElement"),
        null
    )]);

def("ObjectPattern")
    .field("properties", [or(
        def("Property"),
        def("PropertyPattern"),
        def("SpreadPropertyPattern"),
        def("SpreadProperty") // Used by Esprima.
    )]);

def("NamedSpecifier")
    .bases("Specifier")
    // Note: this abstract type is intentionally not buildable.
    .field("id", def("Identifier"))
    .field("name", or(def("Identifier"), null), defaults["null"]);

// Like NamedSpecifier, except type:"ExportSpecifier" and buildable.
// export {<id [as name]>} [from ...];
def("ExportSpecifier")
    .bases("NamedSpecifier")
    .build("id", "name");

// export <*> from ...;
def("ExportBatchSpecifier")
    .bases("Specifier")
    .build();

// Like NamedSpecifier, except type:"ImportSpecifier" and buildable.
// import {<id [as name]>} from ...;
def("ImportSpecifier")
    .bases("NamedSpecifier")
    .build("id", "name");

// import <* as id> from ...;
def("ImportNamespaceSpecifier")
    .bases("Specifier")
    .build("id")
    .field("id", def("Identifier"));

// import <id> from ...;
def("ImportDefaultSpecifier")
    .bases("Specifier")
    .build("id")
    .field("id", def("Identifier"));

def("ExportDeclaration")
    .bases("Declaration")
    .build("default", "declaration", "specifiers", "source")
    .field("default", Boolean)
    .field("declaration", or(
        def("Declaration"),
        def("Expression"), // Implies default.
        null
    ))
    .field("specifiers", [or(
        def("ExportSpecifier"),
        def("ExportBatchSpecifier")
    )], defaults.emptyArray)
    .field("source", or(
        def("Literal"),
        null
    ), defaults["null"]);

def("ImportDeclaration")
    .bases("Declaration")
    .build("specifiers", "source")
    .field("specifiers", [or(
        def("ImportSpecifier"),
        def("ImportNamespaceSpecifier"),
        def("ImportDefaultSpecifier")
    )], defaults.emptyArray)
    .field("source", def("Literal"));
