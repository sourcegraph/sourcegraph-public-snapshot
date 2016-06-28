var lang, isJavaObj,
    hasOwn = Object.prototype.hasOwnProperty;
/**
 * @license Copyright (c) 2010-2014, The Dojo Foundation All Rights Reserved.
 * Available via the MIT or new BSD license.
 * see: http://github.com/jrburke/requirejs for details
 */

function hasProp(obj, prop) {
    return hasOwn.call(obj, prop);
}

isJavaObj = function () {
    return false;
};

if (typeof java !== 'undefined' && java.lang && java.lang.Object) {
    isJavaObj = function (obj) {
        return obj instanceof java.lang.Object;
    };
}

lang = {
    backSlashRegExp: /\\/g,
    ostring: Object.prototype.toString,

    isArray: Array.isArray || function (it) {
        return lang.ostring.call(it) === "[object Array]";
    },

    isFunction: function(it) {
        return lang.ostring.call(it) === "[object Function]";
    },

    isRegExp: function(it) {
        return it && it instanceof RegExp;
    },

    hasProp: hasProp,

    //returns true if the object does not have an own property prop,
    //or if it does, it is a falsy value.
    falseProp: function (obj, prop) {
        return !hasProp(obj, prop) || !obj[prop];
    },

    //gets own property value for given prop on object
    getOwn: function (obj, prop) {
        return hasProp(obj, prop) && obj[prop];
    },

    _mixin: function(dest, source, override){
        var name;
        for (name in source) {
            if(source.hasOwnProperty(name) &&
                (override || !dest.hasOwnProperty(name))) {
                dest[name] = source[name];
            }
        }

        return dest; // Object
    },

    /**
     * mixin({}, obj1, obj2) is allowed. If the last argument is a boolean,
     * then the source objects properties are force copied over to dest.
     */
    mixin: function(dest){
        var parameters = Array.prototype.slice.call(arguments),
            override, i, l;

        if (!dest) { dest = {}; }

        if (parameters.length > 2 && typeof arguments[parameters.length-1] === 'boolean') {
            override = parameters.pop();
        }

        for (i = 1, l = parameters.length; i < l; i++) {
            lang._mixin(dest, parameters[i], override);
        }
        return dest; // Object
    },


    /**
     * Does a type of deep copy. Do not give it anything fancy, best
     * for basic object copies of objects that also work well as
     * JSON-serialized things, or has properties pointing to functions.
     * For non-array/object values, just returns the same object.
     * @param  {Object} obj      copy properties from this object
     * @param  {Object} [result] optional result object to use
     * @return {Object}
     */
    deeplikeCopy: function (obj) {
        var type, result;

        if (lang.isArray(obj)) {
            result = [];
            obj.forEach(function(value) {
                result.push(lang.deeplikeCopy(value));
            });
            return result;
        }

        type = typeof obj;
        if (obj === null || obj === undefined || type === 'boolean' ||
            type === 'string' || type === 'number' || lang.isFunction(obj) ||
            lang.isRegExp(obj)|| isJavaObj(obj)) {
            return obj;
        }

        //Anything else is an object, hopefully.
        result = {};
        lang.eachProp(obj, function(value, key) {
            result[key] = lang.deeplikeCopy(value);
        });
        return result;
    },

    delegate: (function () {
        // boodman/crockford delegation w/ cornford optimization
        function TMP() {}
        return function (obj, props) {
            TMP.prototype = obj;
            var tmp = new TMP();
            TMP.prototype = null;
            if (props) {
                lang.mixin(tmp, props);
            }
            return tmp; // Object
        };
    }()),

    /**
     * Helper function for iterating over an array. If the func returns
     * a true value, it will break out of the loop.
     */
    each: function each(ary, func) {
        if (ary) {
            var i;
            for (i = 0; i < ary.length; i += 1) {
                if (func(ary[i], i, ary)) {
                    break;
                }
            }
        }
    },

    /**
     * Cycles over properties in an object and calls a function for each
     * property value. If the function returns a truthy value, then the
     * iteration is stopped.
     */
    eachProp: function eachProp(obj, func) {
        var prop;
        for (prop in obj) {
            if (hasProp(obj, prop)) {
                if (func(obj[prop], prop)) {
                    break;
                }
            }
        }
    },

    //Similar to Function.prototype.bind, but the "this" object is specified
    //first, since it is easier to read/figure out what "this" will be.
    bind: function bind(obj, fn) {
        return function () {
            return fn.apply(obj, arguments);
        };
    },

    //Escapes a content string to be be a string that has characters escaped
    //for inclusion as part of a JS string.
    jsEscape: function (content) {
        return content.replace(/(["'\\])/g, '\\$1')
            .replace(/[\f]/g, "\\f")
            .replace(/[\b]/g, "\\b")
            .replace(/[\n]/g, "\\n")
            .replace(/[\t]/g, "\\t")
            .replace(/[\r]/g, "\\r");
    }
};
module.exports = lang;
