// soo many different forms via http://requirejs.org/docs/api.html

// Simple Name/Value Pairs

define({
    color: "black",
    size: "unisize"
});

// Definition Functions
define(function () {
    //Do setup work here

    return {
        color: "black",
        size: "unisize"
    }
});

// Definition Functions with Dependencies
define(["./cart", "./inventory"], function(cart, inventory) {
        //return an object to define the "my/shirt" module.
        return {
            color: "blue",
            size: "large",
            addToCart: function() {
                inventory.decrement(this);
                cart.add(this);
            }
        }
    }
);

// Define a Module as a Function
define(["my/cart", "my/inventory"],
    function(cart, inventory) {
        //return a function to define "foo/title".
        //It gets or sets the window title.
        return function(title) {
            return title ? (window.title = title) :
                   inventory.storeName + ' ' + cart.name;
        }
    }
);

// Define a Module with Simplified CommonJS Wrapper
define(function(require, exports, module) {
        var a = require('a'),
            b = require('b');

        //Return the module value
        return function () {};
    }
);

// Define a Module with a Name
define("foo/title",
    ["my/cart", "my/inventory"],
    function(cart, inventory) {
        //Define foo/title object in here.
   }
);

// Relative module names inside define()

define(["require", "./relative/name"], function(require) {
    var mod = require("./relative/name");
});

// shortened syntax that is available for use with translating CommonJS modules

define(function(require) {
    var mod = require("./relative/name");
});

// Specify a JSONP Service Dependency
require(["http://example.com/api/data.json?callback=define"],
    function (data) {
        //The data object will be the API response for the
        //JSONP data call.
        console.log(data);
    }
);

// Note: Generate URLs relative to module doesn't work
define(["require"], function(require) {
    var cssUrl = require.toUrl("./style.css");
});









