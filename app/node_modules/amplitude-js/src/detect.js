/* jshint eqeqeq: false, eqnull: true, freeze: false, bitwise: false */
/* jshint curly: false, unused: false, immed: false, forin: false */

/**
 * Detect.js: User-Agent Parser
 * https://github.com/darcyclarke/Detect.js
 * Dual licensed under the MIT and GPL licenses.
 *
 * @version 2.2.1
 * @author Darcy Clarke
 * @url http://darcyclarke.me
 * @createdat Thu Feb 13 2014 11:36:42 GMT+0000 (WET)
 *
 * Based on UA-Parser (https://github.com/tobie/ua-parser) by Tobie Langel
 *
 * Example Usage:
 * var agentInfo = detect.parse(navigator.userAgent);
 * console.log(agentInfo.browser.family); // Chrome
 *
 */
var detect = (function(root, undefined) {
    // Shim Array.prototype.map if necessary
    // Production steps of ECMA-262, Edition 5, 15.4.4.19
    // Reference: http://es5.github.com/#x15.4.4.19
    if (!Array.prototype.map) {
        Array.prototype.map = function(callback, thisArg) {
            var T, A, k;
            if (this == null) {
                throw new TypeError(" this is null or not defined");
            }
            // 1. Let O be the result of calling ToObject passing the |this| value as the argument.
            var O = Object(this);
            // 2. Let lenValue be the result of calling the Get internal method of O with the argument "length".
            // 3. Let len be ToUint32(lenValue).
            var len = O.length >>> 0;
            // 4. If IsCallable(callback) is false, throw a TypeError exception.
            // See: http://es5.github.com/#x9.11
            if (typeof callback !== "function") {
                throw new TypeError(callback + " is not a function");
            }
            // 5. If thisArg was supplied, let T be thisArg; else let T be undefined.
            if (thisArg) {
                T = thisArg;
            }
            // 6. Let A be a new array created as if by the expression new Array(len) where Array is
            // the standard built-in constructor with that name and len is the value of len.
            A = new Array(len);
            // 7. Let k be 0
            k = 0;
            // 8. Repeat, while k < len
            while (k < len) {
                var kValue, mappedValue;
                // a. Let Pk be ToString(k).
                //   This is implicit for LHS operands of the in operator
                // b. Let kPresent be the result of calling the HasProperty internal method of O with argument Pk.
                //   This step can be combined with c
                // c. If kPresent is true, then
                if (k in O) {
                    // i. Let kValue be the result of calling the Get internal method of O with argument Pk.
                    kValue = O[k];
                    // ii. Let mappedValue be the result of calling the Call internal method of callback
                    // with T as the this value and argument list containing kValue, k, and O.
                    mappedValue = callback.call(T, kValue, k, O);
                    // iii. Call the DefineOwnProperty internal method of A with arguments
                    // Pk, Property Descriptor {Value: mappedValue, : true, Enumerable: true, Configurable: true},
                    // and false.
                    // In browsers that support Object.defineProperty, use the following:
                    // Object.defineProperty(A, Pk, { value: mappedValue, writable: true, enumerable: true, configurable: true });
                    // For best browser support, use the following:
                    A[k] = mappedValue;
                }
                // d. Increase k by 1.
                k++;
            }
            // 9. return A
            return A;
        };
    }
    // Detect
    var detect = root.detect = function() {
        // Context
        var _this = function() {};
        // Regexes
        var regexes = {
            browser_parsers: [ {
                regex: "(SeaMonkey|Camino)/(\\d+)\\.(\\d+)\\.?([ab]?\\d+[a-z]*)",
                family_replacement: "Camino"
            }, {
                regex: "(Fennec)/(\\d+)\\.(\\d+)\\.?([ab]?\\d+[a-z]*)",
                family_replacement: "Firefox Mobile"
            }, {
                regex: "(Fennec)/(\\d+)\\.(\\d+)(pre)",
                family_replacment: "Firefox Mobile"
            }, {
                regex: "(Fennec)/(\\d+)\\.(\\d+)",
                family_replacement: "Firefox Mobile"
            }, {
                regex: "Mobile.*(Firefox)/(\\d+)\\.(\\d+)",
                family_replacement: "Firefox"
            }, {
                regex: "(Namoroka|Shiretoko|Minefield)/(\\d+)\\.(\\d+)\\.(\\d+(?:pre)?)",
                family_replacement: "Firefox"
            }, {
                regex: "(Firefox)/(\\d+)\\.(\\d+)(a\\d+[a-z]*)",
                family_replacement: "Firefox"
            }, {
                regex: "(Firefox)/(\\d+)\\.(\\d+)(b\\d+[a-z]*)",
                family_replacement: "Firefox"
            }, {
                regex: "(Firefox)-(?:\\d+\\.\\d+)?/(\\d+)\\.(\\d+)(a\\d+[a-z]*)",
                family_replacement: "Firefox"
            }, {
                regex: "(Firefox)-(?:\\d+\\.\\d+)?/(\\d+)\\.(\\d+)(b\\d+[a-z]*)",
                family_replacement: "Firefox"
            }, {
                regex: "(Namoroka|Shiretoko|Minefield)/(\\d+)\\.(\\d+)([ab]\\d+[a-z]*)?",
                family_replacement: "Firefox"
            }, {
                regex: "(Navigator)/(\\d+)\\.(\\d+)\\.(\\d+)",
                family_replacement: "Netscape"
            }, {
                regex: "(Navigator)/(\\d+)\\.(\\d+)([ab]\\d+)",
                family_replacement: "Netscape"
            }, {
                regex: "(Netscape6)/(\\d+)\\.(\\d+)\\.(\\d+)",
                family_replacement: "Netscape"
            }, {
                regex: "(Opera Tablet).*Version/(\\d+)\\.(\\d+)(?:\\.(\\d+))?",
                family_replacement: "Opera Mobile",
                tablet: true
            }, {
                regex: "(Opera)/.+Opera Mobi.+Version/(\\d+)\\.(\\d+)",
                family_replacement: "Opera Mobile"
            }, {
                regex: "Opera Mobi",
                family_replacement: "Opera Mobile"
            }, {
                regex: "(Opera Mini)/(\\d+)\\.(\\d+)",
                family_replacement: "Opera Mini"
            }, {
                regex: "(Opera Mini)/att/(\\d+)\\.(\\d+)",
                family_replacement: "Opera Mini"
            }, {
                regex: "(Opera)/9.80.*Version/(\\d+)\\.(\\d+)(?:\\.(\\d+))?",
                family_replacement: "Opera"
            }, {
                regex: "(konqueror)/(\\d+)\\.(\\d+)\\.(\\d+)",
                family_replacement: "Konqueror"
            }, {
                regex: "(CrMo)/(\\d+)\\.(\\d+)\\.(\\d+)\\.(\\d+)",
                family_replacement: "Chrome Mobile"
            }, {
                regex: "(CriOS)/(\\d+)\\.(\\d+)\\.(\\d+)\\.(\\d+)",
                family_replacement: "Chrome Mobile"
            }, {
                regex: "(Chrome)/(\\d+)\\.(\\d+)\\.(\\d+)\\.(\\d+) Mobile",
                family_replacement: "Chrome Mobile"
            }, {
                regex: "(KDE)",
                family_replacement: "Konqueror"
            }, {
                regex: "(OmniWeb|Camino|Chrome|Netscape|Konqueror|Opera Mini|iCab)/(\\d+)\\.(\\d+)\\.(\\d+)"
            }, {
                regex: "(Camino|Chrome|Netscape|Opera Mini|Opera|Konqueror|iCab)/(\\d+)\\.(\\d+)"
            }, {
                regex: "(iCab) (\\d+)\\.(\\d+)\\.(\\d+)"
            }, {
                regex: "(iCab|Opera) (\\d+)\\.(\\d+)\\.?(\\d+)?"
            }, {
                regex: "(IEMobile)[ /](\\d+)\\.(\\d+)",
                family_replacement: "IE Mobile"
            }, {
                regex: "(MSIE) (\\d+)\\.(\\d+).*XBLWP7",
                family_replacement: "IE"
            }, {
                regex: "(Firefox)/(\\d+)\\.(\\d+)\\.(\\d+)"
            }, {
                regex: "(Firefox)/(\\d+)\\.(\\d+)(pre|[ab]\\d+[a-z]*)?"
            }, {
                regex: "(iPod).+Version/(\\d+)\\.(\\d+)\\.(\\d+)",
                family_replacement: "Mobile Safari",
                manufacturer: "Apple"
            }, {
                regex: "(iPod).*Version/(\\d+)\\.(\\d+)",
                family_replacement: "Mobile Safari",
                manufacturer: "Apple"
            }, {
                regex: "(iPod)",
                family_replacement: "Mobile Safari",
                manufacturer: "Apple"
            }, {
                regex: "(iPhone).*Version/(\\d+)\\.(\\d+)\\.(\\d+)",
                family_replacement: "Mobile Safari",
                manufacturer: "Apple"
            }, {
                regex: "(iPhone).*Version/(\\d+)\\.(\\d+)",
                family_replacement: "Mobile Safari",
                manufacturer: "Apple"
            }, {
                regex: "(iPhone)",
                family_replacement: "Mobile Safari",
                manufacturer: "Apple"
            }, {
                regex: "(iPad).*Version/(\\d+)\\.(\\d+)\\.(\\d+)",
                family_replacement: "Mobile Safari",
                tablet: true,
                manufacturer: "Apple"
            }, {
                regex: "(iPad).*Version/(\\d+)\\.(\\d+)",
                family_replacement: "Mobile Safari",
                tablet: true,
                manufacturer: "Apple"
            }, {
                regex: "(iPad)",
                family_replacement: "Mobile Safari",
                tablet: true,
                manufacturer: "Apple"
            }, {
                regex: "(PlayBook).+RIM Tablet OS (\\d+)\\.(\\d+)\\.(\\d+)",
                family_replacement: "Blackberry",
                tablet: true,
                manufacturer: "Nokia"
            }, {
                regex: "(Black[bB]erry).+Version/(\\d+)\\.(\\d+)\\.(\\d+)",
                family_replacement: "Blackberry",
                manufacturer: "RIM"
            }, {
                regex: "(Black[bB]erry)\\s?(\\d+)",
                family_replacement: "Blackberry",
                manufacturer: "RIM"
            }, {
                regex: "(OmniWeb)/v(\\d+)\\.(\\d+)"
            }, {
                regex: "(AppleWebKit)/(\\d+)\\.?(\\d+)?\\+ .* Version/\\d+\\.\\d+.\\d+ Safari/",
                family_replacement: "Safari"
            }, {
                regex: "(Version)/(\\d+)\\.(\\d+)(?:\\.(\\d+))?.*Safari/",
                family_replacement: "Safari"
            }, {
                regex: "(Safari)/\\d+"
            }, {
                regex: "Trident(.*)rv.(\\d+)\\.(\\d+)",
                family_replacement: "IE"
            }, {
                regex: "(MSIE) (\\d+)\\.(\\d+)",
                family_replacement: "IE"
            } ],
            os_parsers: [ {
                regex: "(Silk-Accelerated=[a-z]{4,5})",
                os_replacement: "Android"
            }, {
                regex: "(Windows Phone 6\\.5)",
                os_replacement: "Windows Phone"
            }, {
                regex: "(XBLWP7)",
                os_replacement: "Windows Phone"
            }, {
                regex: "(Windows Phone OS) (\\d+)\\.(\\d+)",
                os_replacement: "Windows Phone"
            }, {
                regex: "(Win)",
                os_replacement: "Windows"
            }, {
                regex: "(Mac) OS X (\\d+)[_.](\\d+)(?:[_.](\\d+))?",
                manufacturer: "Apple"
            }, {
                regex: "(?:PPC|Intel) (Mac OS X)",
                os_replacement: "Mac",
                manufacturer: "Apple"
            }, {
                regex: "(iPhone|iPod)",
                os_replacement: "iPhone",
                manufacturer: "Apple"
            }, {
                regex: "(iPad)",
                os_replacement: "iPad",
                tablet: true,
                manufacturer: "Apple"
            }, {
                regex: "(CrOS) [a-z0-9_]+ (\\d+)\\.(\\d+)(?:\\.(\\d+))?",
                os_replacement: "Chrome OS"
            }, {
                regex: "(Symbian[Oo][Ss])/(\\d+)\\.(\\d+)",
                os_replacement: "Symbian"
            }, {
                regex: "(Symbian/3).+NokiaBrowser/7\\.3",
                os_replacement: "Symbian"
            }, {
                regex: "(Symbian/3).+NokiaBrowser/7\\.4",
                os_replacement: "Symbian"
            }, {
                regex: "(Symbian/3)",
                os_replacement: "Symbian"
            }, {
                regex: "(Series 60|SymbOS|S60)",
                os_replacement: "Symbian"
            }, {
                regex: "Symbian [Oo][Ss]",
                os_replacement: "Symbian"
            }, {
                regex: "(Black[Bb]erry)[0-9a-z]+/(\\d+)\\.(\\d+)\\.(\\d+)(?:\\.(\\d+))?",
                os_replacement: "BlackBerry",
                manufacturer: "RIM"
            }, {
                regex: "(Black[Bb]erry).+Version/(\\d+)\\.(\\d+)\\.(\\d+)(?:\\.(\\d+))?",
                os_replacement: "BlackBerry",
                manufacturer: "RIM"
            }, {
                regex: "(RIM Tablet OS) (\\d+)\\.(\\d+)\\.(\\d+)",
                os_replacement: "BlackBerry",
                tablet: true,
                manufacturer: "RIM"
            }, {
                regex: "(Play[Bb]ook)",
                os_replacement: "BlackBerry",
                tablet: true,
                manufacturer: "RIM"
            }, {
                regex: "(Black[Bb]erry)",
                os_replacement: "Blackberry",
                manufacturer: "RIM"
            }, {
                regex: "(Linux)"
            }, {
                regex: "(Android)"
            } ],
            mobile_os_families: [ "Windows Phone", "Windows CE", "Symbian" ],
            device_parsers: [],
            mobile_browser_families: [ "Firefox Mobile", "Opera Mobile", "Opera Mini", "Mobile Safari", "webOS", "IE Mobile", "Playstation Portable", "Nokia", "Blackberry", "Palm", "Silk", "Android", "Maemo", "Obigo", "Netfront", "AvantGo", "Teleca", "SEMC-Browser", "Bolt", "Iris", "UP.Browser", "Symphony", "Minimo", "Bunjaloo", "Jasmine", "Dolfin", "Polaris", "BREW", "Chrome Mobile", "Chrome Mobile iOS", "UC Browser", "Tizen Browser" ]
        };
        // Parsers
        _this.parsers = [ "device_parsers", "browser_parsers", "os_parsers", "mobile_os_families", "mobile_browser_families" ];
        // Types
        _this.types = [ "browser", "os", "device" ];
        // Regular Expressions
        _this.regexes = regexes || function() {
            var results = {};
            _this.parsers.map(function(parser) {
                results[parser] = [];
            });
            return results;
        }();
        // Families
        _this.families = function() {
            var results = {};
            _this.types.map(function(type) {
                results[type] = [];
            });
            return results;
        }();
        // Utility Variables
        var ArrayProto = Array.prototype, ObjProto = Object.prototype, FuncProto = Function.prototype, nativeForEach = ArrayProto.forEach, nativeIndexOf = ArrayProto.indexOf;
        var push = ArrayProto.push, slice = ArrayProto.slice, hasOwnProperty = ObjProto.hasOwnProperty;

        // Find Utility
        var find = function(ua, obj) {
            var ret = {};
            for (var i = 0; i < obj.length; i++) {
                ret = obj[i](ua);
                if (ret) {
                    break;
                }
            }
            return ret;
        };
        // Remove Utility
        var remove = function(arr, props) {
            each(arr, function(obj) {
                each(props, function(prop) {
                    delete obj[prop];
                });
            });
        };
        // Has Utility
        var has = function(obj, key) {
            return obj != null && hasOwnProperty.call(obj, key);
        };
        // Each Utility
        var each = function(obj, iterator, context) {
            if (obj == null) return;
            if (nativeForEach && obj.forEach === nativeForEach) {
                obj.forEach(iterator, context);
            } else if (obj.length === +obj.length) {
                for (var i = 0, l = obj.length; i < l; i++) {
                    iterator.call(context, obj[i], i, obj);
                }
            } else {
                for (var key in obj) {
                    if (has(obj, key)) {
                        iterator.call(context, obj[key], key, obj);
                    }
                }
            }
        };
        // Extend Utiltiy
        var extend = function(obj) {
            each(slice.call(arguments, 1), function(source) {
                for (var prop in source) {
                    obj[prop] = source[prop];
                }
            });
            return obj;
        };
        // Check String Utility
        var check = function(str) {
            return !!(str && typeof str != "undefined" && str != null);
        };
        // To Version String Utility
        var toVersionString = function(obj) {
            var output = "";
            obj = obj || {};
            if (check(obj)) {
                if (check(obj.major)) {
                    output += obj.major;
                    if (check(obj.minor)) {
                        output += "." + obj.minor;
                        if (check(obj.patch)) {
                            output += "." + obj.patch;
                        }
                    }
                }
            }
            return output;
        };
        // To String Utility
        var toString = function(obj) {
            obj = obj || {};
            var suffix = toVersionString(obj);
            if (suffix) suffix = " " + suffix;
            return obj && check(obj.family) ? obj.family + suffix : "";
        };
        // Parse User-Agent String
        _this.parse = function(ua) {
            // Parsers Utility
            var parsers = function(type) {
                return _this.regexes[type + "_parsers"].map(function(obj) {
                    var regexp = new RegExp(obj.regex), rep = obj[(type === "browser" ? "family" : type) + "_replacement"], major_rep = obj.major_version_replacement;
                    function parser(ua) {
                        var m = ua.match(regexp);
                        if (!m) return null;
                        var ret = {};
                        ret.family = (rep ? rep.replace("$1", m[1]) : m[1]) || null;
                        ret.major = parseInt(major_rep ? major_rep : m[2]) || null;
                        ret.minor = m[3] ? parseInt(m[3]) : null;
                        ret.patch = m[4] ? parseInt(m[4]) : null;
                        ret.tablet = obj.tablet;
                        ret.man = obj.manufacturer || null;
                        return ret;
                    }
                    return parser;
                });
            };
            // User Agent
            var UserAgent = function() {};
            // Browsers Parsed
            var browser_parsers = parsers("browser");
            // Operating Systems Parsed
            var os_parsers = parsers("os");
            // Devices Parsed
            var device_parsers = parsers("device");
            // Set Agent
            var a = new UserAgent();
            // Remember the original user agent string
            a.source = ua;
            // Set Browser
            a.browser = find(ua, browser_parsers);
            if (check(a.browser)) {
                a.browser.name = toString(a.browser);
                a.browser.version = toVersionString(a.browser);
            } else {
                a.browser = {};
            }
            // Set OS
            a.os = find(ua, os_parsers);
            if (check(a.os)) {
                a.os.name = toString(a.os);
                a.os.version = toVersionString(a.os);
            } else {
                a.os = {};
            }
            // Set Device
            a.device = find(ua, device_parsers);
            if (check(a.device)) {
                a.device.name = toString(a.device);
                a.device.version = toVersionString(a.device);
            } else {
                a.device = {
                    tablet: false,
                    family: null
                };
            }
            // Determine Device Type
            var mobile_agents = {};
            var mobile_browser_families = _this.regexes.mobile_browser_families.map(function(str) {
                mobile_agents[str] = true;
            });
            var mobile_os_families = _this.regexes.mobile_os_families.map(function(str) {
                mobile_agents[str] = true;
            });
            // Is Spider
            if (a.browser.family === "Spider") {
                a.device.type = "Spider";
            } else if (a.browser.tablet || a.os.tablet || a.device.tablet) {
                a.device.type = "Tablet";
            } else if (mobile_agents.hasOwnProperty(a.browser.family)) {
                a.device.type = "Mobile";
            } else {
                a.device.type = "Desktop";
            }
            // Determine Device Manufacturer
            a.device.manufacturer = a.browser.man || a.os.man || a.device.man || null;
            // Cleanup Objects
            remove([ a.browser, a.os, a.device ], [ "tablet", "man" ]);
            // Return Agent
            return a;
        };
        // Return context
        return _this;
    }();
    return detect;
})(window);

module.exports = detect;
