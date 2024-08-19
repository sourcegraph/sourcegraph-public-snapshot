var __webpack_modules__ = {
    5756: (module, __unused_webpack_exports, __webpack_require__) => {
        "use strict";
        module = __webpack_require__.nmd(module);
        const wrapAnsi16 = (fn, offset) => (...args) => {
            const code = fn(...args);
            return `[${code + offset}m`;
        };
        const wrapAnsi256 = (fn, offset) => (...args) => {
            const code = fn(...args);
            return `[${38 + offset};5;${code}m`;
        };
        const wrapAnsi16m = (fn, offset) => (...args) => {
            const rgb = fn(...args);
            return `[${38 + offset};2;${rgb[0]};${rgb[1]};${rgb[2]}m`;
        };
        const ansi2ansi = n => n;
        const rgb2rgb = (r, g, b) => [ r, g, b ];
        const setLazyProperty = (object, property, get) => {
            Object.defineProperty(object, property, {
                get: () => {
                    const value = get();
                    Object.defineProperty(object, property, {
                        value,
                        enumerable: true,
                        configurable: true
                    });
                    return value;
                },
                enumerable: true,
                configurable: true
            });
        };
        let colorConvert;
        const makeDynamicStyles = (wrap, targetSpace, identity, isBackground) => {
            if (colorConvert === undefined) {
                colorConvert = __webpack_require__(9208);
            }
            const offset = isBackground ? 10 : 0;
            const styles = {};
            for (const [sourceSpace, suite] of Object.entries(colorConvert)) {
                const name = sourceSpace === "ansi16" ? "ansi" : sourceSpace;
                if (sourceSpace === targetSpace) {
                    styles[name] = wrap(identity, offset);
                } else if (typeof suite === "object") {
                    styles[name] = wrap(suite[targetSpace], offset);
                }
            }
            return styles;
        };
        function assembleStyles() {
            const codes = new Map;
            const styles = {
                modifier: {
                    reset: [ 0, 0 ],
                    bold: [ 1, 22 ],
                    dim: [ 2, 22 ],
                    italic: [ 3, 23 ],
                    underline: [ 4, 24 ],
                    inverse: [ 7, 27 ],
                    hidden: [ 8, 28 ],
                    strikethrough: [ 9, 29 ]
                },
                color: {
                    black: [ 30, 39 ],
                    red: [ 31, 39 ],
                    green: [ 32, 39 ],
                    yellow: [ 33, 39 ],
                    blue: [ 34, 39 ],
                    magenta: [ 35, 39 ],
                    cyan: [ 36, 39 ],
                    white: [ 37, 39 ],
                    blackBright: [ 90, 39 ],
                    redBright: [ 91, 39 ],
                    greenBright: [ 92, 39 ],
                    yellowBright: [ 93, 39 ],
                    blueBright: [ 94, 39 ],
                    magentaBright: [ 95, 39 ],
                    cyanBright: [ 96, 39 ],
                    whiteBright: [ 97, 39 ]
                },
                bgColor: {
                    bgBlack: [ 40, 49 ],
                    bgRed: [ 41, 49 ],
                    bgGreen: [ 42, 49 ],
                    bgYellow: [ 43, 49 ],
                    bgBlue: [ 44, 49 ],
                    bgMagenta: [ 45, 49 ],
                    bgCyan: [ 46, 49 ],
                    bgWhite: [ 47, 49 ],
                    bgBlackBright: [ 100, 49 ],
                    bgRedBright: [ 101, 49 ],
                    bgGreenBright: [ 102, 49 ],
                    bgYellowBright: [ 103, 49 ],
                    bgBlueBright: [ 104, 49 ],
                    bgMagentaBright: [ 105, 49 ],
                    bgCyanBright: [ 106, 49 ],
                    bgWhiteBright: [ 107, 49 ]
                }
            };
            styles.color.gray = styles.color.blackBright;
            styles.bgColor.bgGray = styles.bgColor.bgBlackBright;
            styles.color.grey = styles.color.blackBright;
            styles.bgColor.bgGrey = styles.bgColor.bgBlackBright;
            for (const [groupName, group] of Object.entries(styles)) {
                for (const [styleName, style] of Object.entries(group)) {
                    styles[styleName] = {
                        open: `[${style[0]}m`,
                        close: `[${style[1]}m`
                    };
                    group[styleName] = styles[styleName];
                    codes.set(style[0], style[1]);
                }
                Object.defineProperty(styles, groupName, {
                    value: group,
                    enumerable: false
                });
            }
            Object.defineProperty(styles, "codes", {
                value: codes,
                enumerable: false
            });
            styles.color.close = "[39m";
            styles.bgColor.close = "[49m";
            setLazyProperty(styles.color, "ansi", (() => makeDynamicStyles(wrapAnsi16, "ansi16", ansi2ansi, false)));
            setLazyProperty(styles.color, "ansi256", (() => makeDynamicStyles(wrapAnsi256, "ansi256", ansi2ansi, false)));
            setLazyProperty(styles.color, "ansi16m", (() => makeDynamicStyles(wrapAnsi16m, "rgb", rgb2rgb, false)));
            setLazyProperty(styles.bgColor, "ansi", (() => makeDynamicStyles(wrapAnsi16, "ansi16", ansi2ansi, true)));
            setLazyProperty(styles.bgColor, "ansi256", (() => makeDynamicStyles(wrapAnsi256, "ansi256", ansi2ansi, true)));
            setLazyProperty(styles.bgColor, "ansi16m", (() => makeDynamicStyles(wrapAnsi16m, "rgb", rgb2rgb, true)));
            return styles;
        }
        Object.defineProperty(module, "exports", {
            enumerable: true,
            get: assembleStyles
        });
    },
    1201: (module, __unused_webpack_exports, __webpack_require__) => {
        "use strict";
        const ansiStyles = __webpack_require__(5756);
        const {stdout: stdoutColor, stderr: stderrColor} = __webpack_require__(9797);
        const {stringReplaceAll, stringEncaseCRLFWithFirstIndex} = __webpack_require__(8564);
        const {isArray} = Array;
        const levelMapping = [ "ansi", "ansi", "ansi256", "ansi16m" ];
        const styles = Object.create(null);
        const applyOptions = (object, options = {}) => {
            if (options.level && !(Number.isInteger(options.level) && options.level >= 0 && options.level <= 3)) {
                throw new Error("The `level` option should be an integer from 0 to 3");
            }
            const colorLevel = stdoutColor ? stdoutColor.level : 0;
            object.level = options.level === undefined ? colorLevel : options.level;
        };
        class ChalkClass {
            constructor(options) {
                return chalkFactory(options);
            }
        }
        const chalkFactory = options => {
            const chalk = {};
            applyOptions(chalk, options);
            chalk.template = (...arguments_) => chalkTag(chalk.template, ...arguments_);
            Object.setPrototypeOf(chalk, Chalk.prototype);
            Object.setPrototypeOf(chalk.template, chalk);
            chalk.template.constructor = () => {
                throw new Error("`chalk.constructor()` is deprecated. Use `new chalk.Instance()` instead.");
            };
            chalk.template.Instance = ChalkClass;
            return chalk.template;
        };
        function Chalk(options) {
            return chalkFactory(options);
        }
        for (const [styleName, style] of Object.entries(ansiStyles)) {
            styles[styleName] = {
                get() {
                    const builder = createBuilder(this, createStyler(style.open, style.close, this._styler), this._isEmpty);
                    Object.defineProperty(this, styleName, {
                        value: builder
                    });
                    return builder;
                }
            };
        }
        styles.visible = {
            get() {
                const builder = createBuilder(this, this._styler, true);
                Object.defineProperty(this, "visible", {
                    value: builder
                });
                return builder;
            }
        };
        const usedModels = [ "rgb", "hex", "keyword", "hsl", "hsv", "hwb", "ansi", "ansi256" ];
        for (const model of usedModels) {
            styles[model] = {
                get() {
                    const {level} = this;
                    return function(...arguments_) {
                        const styler = createStyler(ansiStyles.color[levelMapping[level]][model](...arguments_), ansiStyles.color.close, this._styler);
                        return createBuilder(this, styler, this._isEmpty);
                    };
                }
            };
        }
        for (const model of usedModels) {
            const bgModel = "bg" + model[0].toUpperCase() + model.slice(1);
            styles[bgModel] = {
                get() {
                    const {level} = this;
                    return function(...arguments_) {
                        const styler = createStyler(ansiStyles.bgColor[levelMapping[level]][model](...arguments_), ansiStyles.bgColor.close, this._styler);
                        return createBuilder(this, styler, this._isEmpty);
                    };
                }
            };
        }
        const proto = Object.defineProperties((() => {}), {
            ...styles,
            level: {
                enumerable: true,
                get() {
                    return this._generator.level;
                },
                set(level) {
                    this._generator.level = level;
                }
            }
        });
        const createStyler = (open, close, parent) => {
            let openAll;
            let closeAll;
            if (parent === undefined) {
                openAll = open;
                closeAll = close;
            } else {
                openAll = parent.openAll + open;
                closeAll = close + parent.closeAll;
            }
            return {
                open,
                close,
                openAll,
                closeAll,
                parent
            };
        };
        const createBuilder = (self, _styler, _isEmpty) => {
            const builder = (...arguments_) => {
                if (isArray(arguments_[0]) && isArray(arguments_[0].raw)) {
                    return applyStyle(builder, chalkTag(builder, ...arguments_));
                }
                return applyStyle(builder, arguments_.length === 1 ? "" + arguments_[0] : arguments_.join(" "));
            };
            Object.setPrototypeOf(builder, proto);
            builder._generator = self;
            builder._styler = _styler;
            builder._isEmpty = _isEmpty;
            return builder;
        };
        const applyStyle = (self, string) => {
            if (self.level <= 0 || !string) {
                return self._isEmpty ? "" : string;
            }
            let styler = self._styler;
            if (styler === undefined) {
                return string;
            }
            const {openAll, closeAll} = styler;
            if (string.indexOf("") !== -1) {
                while (styler !== undefined) {
                    string = stringReplaceAll(string, styler.close, styler.open);
                    styler = styler.parent;
                }
            }
            const lfIndex = string.indexOf("\n");
            if (lfIndex !== -1) {
                string = stringEncaseCRLFWithFirstIndex(string, closeAll, openAll, lfIndex);
            }
            return openAll + string + closeAll;
        };
        let template;
        const chalkTag = (chalk, ...strings) => {
            const [firstString] = strings;
            if (!isArray(firstString) || !isArray(firstString.raw)) {
                return strings.join(" ");
            }
            const arguments_ = strings.slice(1);
            const parts = [ firstString.raw[0] ];
            for (let i = 1; i < firstString.length; i++) {
                parts.push(String(arguments_[i - 1]).replace(/[{}\\]/g, "\\$&"), String(firstString.raw[i]));
            }
            if (template === undefined) {
                template = __webpack_require__(2154);
            }
            return template(chalk, parts.join(""));
        };
        Object.defineProperties(Chalk.prototype, styles);
        const chalk = Chalk();
        chalk.supportsColor = stdoutColor;
        chalk.stderr = Chalk({
            level: stderrColor ? stderrColor.level : 0
        });
        chalk.stderr.supportsColor = stderrColor;
        module.exports = chalk;
    },
    2154: module => {
        "use strict";
        const TEMPLATE_REGEX = /(?:\\(u(?:[a-f\d]{4}|\{[a-f\d]{1,6}\})|x[a-f\d]{2}|.))|(?:\{(~)?(\w+(?:\([^)]*\))?(?:\.\w+(?:\([^)]*\))?)*)(?:[ \t]|(?=\r?\n)))|(\})|((?:.|[\r\n\f])+?)/gi;
        const STYLE_REGEX = /(?:^|\.)(\w+)(?:\(([^)]*)\))?/g;
        const STRING_REGEX = /^(['"])((?:\\.|(?!\1)[^\\])*)\1$/;
        const ESCAPE_REGEX = /\\(u(?:[a-f\d]{4}|{[a-f\d]{1,6}})|x[a-f\d]{2}|.)|([^\\])/gi;
        const ESCAPES = new Map([ [ "n", "\n" ], [ "r", "\r" ], [ "t", "\t" ], [ "b", "\b" ], [ "f", "\f" ], [ "v", "\v" ], [ "0", "\0" ], [ "\\", "\\" ], [ "e", "" ], [ "a", "" ] ]);
        function unescape(c) {
            const u = c[0] === "u";
            const bracket = c[1] === "{";
            if (u && !bracket && c.length === 5 || c[0] === "x" && c.length === 3) {
                return String.fromCharCode(parseInt(c.slice(1), 16));
            }
            if (u && bracket) {
                return String.fromCodePoint(parseInt(c.slice(2, -1), 16));
            }
            return ESCAPES.get(c) || c;
        }
        function parseArguments(name, arguments_) {
            const results = [];
            const chunks = arguments_.trim().split(/\s*,\s*/g);
            let matches;
            for (const chunk of chunks) {
                const number = Number(chunk);
                if (!Number.isNaN(number)) {
                    results.push(number);
                } else if (matches = chunk.match(STRING_REGEX)) {
                    results.push(matches[2].replace(ESCAPE_REGEX, ((m, escape, character) => escape ? unescape(escape) : character)));
                } else {
                    throw new Error(`Invalid Chalk template style argument: ${chunk} (in style '${name}')`);
                }
            }
            return results;
        }
        function parseStyle(style) {
            STYLE_REGEX.lastIndex = 0;
            const results = [];
            let matches;
            while ((matches = STYLE_REGEX.exec(style)) !== null) {
                const name = matches[1];
                if (matches[2]) {
                    const args = parseArguments(name, matches[2]);
                    results.push([ name ].concat(args));
                } else {
                    results.push([ name ]);
                }
            }
            return results;
        }
        function buildStyle(chalk, styles) {
            const enabled = {};
            for (const layer of styles) {
                for (const style of layer.styles) {
                    enabled[style[0]] = layer.inverse ? null : style.slice(1);
                }
            }
            let current = chalk;
            for (const [styleName, styles] of Object.entries(enabled)) {
                if (!Array.isArray(styles)) {
                    continue;
                }
                if (!(styleName in current)) {
                    throw new Error(`Unknown Chalk style: ${styleName}`);
                }
                current = styles.length > 0 ? current[styleName](...styles) : current[styleName];
            }
            return current;
        }
        module.exports = (chalk, temporary) => {
            const styles = [];
            const chunks = [];
            let chunk = [];
            temporary.replace(TEMPLATE_REGEX, ((m, escapeCharacter, inverse, style, close, character) => {
                if (escapeCharacter) {
                    chunk.push(unescape(escapeCharacter));
                } else if (style) {
                    const string = chunk.join("");
                    chunk = [];
                    chunks.push(styles.length === 0 ? string : buildStyle(chalk, styles)(string));
                    styles.push({
                        inverse,
                        styles: parseStyle(style)
                    });
                } else if (close) {
                    if (styles.length === 0) {
                        throw new Error("Found extraneous } in Chalk template literal");
                    }
                    chunks.push(buildStyle(chalk, styles)(chunk.join("")));
                    chunk = [];
                    styles.pop();
                } else {
                    chunk.push(character);
                }
            }));
            chunks.push(chunk.join(""));
            if (styles.length > 0) {
                const errMessage = `Chalk template literal is missing ${styles.length} closing bracket${styles.length === 1 ? "" : "s"} (\`}\`)`;
                throw new Error(errMessage);
            }
            return chunks.join("");
        };
    },
    8564: module => {
        "use strict";
        const stringReplaceAll = (string, substring, replacer) => {
            let index = string.indexOf(substring);
            if (index === -1) {
                return string;
            }
            const substringLength = substring.length;
            let endIndex = 0;
            let returnValue = "";
            do {
                returnValue += string.substr(endIndex, index - endIndex) + substring + replacer;
                endIndex = index + substringLength;
                index = string.indexOf(substring, endIndex);
            } while (index !== -1);
            returnValue += string.substr(endIndex);
            return returnValue;
        };
        const stringEncaseCRLFWithFirstIndex = (string, prefix, postfix, index) => {
            let endIndex = 0;
            let returnValue = "";
            do {
                const gotCR = string[index - 1] === "\r";
                returnValue += string.substr(endIndex, (gotCR ? index - 1 : index) - endIndex) + prefix + (gotCR ? "\r\n" : "\n") + postfix;
                endIndex = index + 1;
                index = string.indexOf("\n", endIndex);
            } while (index !== -1);
            returnValue += string.substr(endIndex);
            return returnValue;
        };
        module.exports = {
            stringReplaceAll,
            stringEncaseCRLFWithFirstIndex
        };
    },
    2538: (module, __unused_webpack_exports, __webpack_require__) => {
        const cssKeywords = __webpack_require__(6150);
        const reverseKeywords = {};
        for (const key of Object.keys(cssKeywords)) {
            reverseKeywords[cssKeywords[key]] = key;
        }
        const convert = {
            rgb: {
                channels: 3,
                labels: "rgb"
            },
            hsl: {
                channels: 3,
                labels: "hsl"
            },
            hsv: {
                channels: 3,
                labels: "hsv"
            },
            hwb: {
                channels: 3,
                labels: "hwb"
            },
            cmyk: {
                channels: 4,
                labels: "cmyk"
            },
            xyz: {
                channels: 3,
                labels: "xyz"
            },
            lab: {
                channels: 3,
                labels: "lab"
            },
            lch: {
                channels: 3,
                labels: "lch"
            },
            hex: {
                channels: 1,
                labels: [ "hex" ]
            },
            keyword: {
                channels: 1,
                labels: [ "keyword" ]
            },
            ansi16: {
                channels: 1,
                labels: [ "ansi16" ]
            },
            ansi256: {
                channels: 1,
                labels: [ "ansi256" ]
            },
            hcg: {
                channels: 3,
                labels: [ "h", "c", "g" ]
            },
            apple: {
                channels: 3,
                labels: [ "r16", "g16", "b16" ]
            },
            gray: {
                channels: 1,
                labels: [ "gray" ]
            }
        };
        module.exports = convert;
        for (const model of Object.keys(convert)) {
            if (!("channels" in convert[model])) {
                throw new Error("missing channels property: " + model);
            }
            if (!("labels" in convert[model])) {
                throw new Error("missing channel labels property: " + model);
            }
            if (convert[model].labels.length !== convert[model].channels) {
                throw new Error("channel and label counts mismatch: " + model);
            }
            const {channels, labels} = convert[model];
            delete convert[model].channels;
            delete convert[model].labels;
            Object.defineProperty(convert[model], "channels", {
                value: channels
            });
            Object.defineProperty(convert[model], "labels", {
                value: labels
            });
        }
        convert.rgb.hsl = function(rgb) {
            const r = rgb[0] / 255;
            const g = rgb[1] / 255;
            const b = rgb[2] / 255;
            const min = Math.min(r, g, b);
            const max = Math.max(r, g, b);
            const delta = max - min;
            let h;
            let s;
            if (max === min) {
                h = 0;
            } else if (r === max) {
                h = (g - b) / delta;
            } else if (g === max) {
                h = 2 + (b - r) / delta;
            } else if (b === max) {
                h = 4 + (r - g) / delta;
            }
            h = Math.min(h * 60, 360);
            if (h < 0) {
                h += 360;
            }
            const l = (min + max) / 2;
            if (max === min) {
                s = 0;
            } else if (l <= .5) {
                s = delta / (max + min);
            } else {
                s = delta / (2 - max - min);
            }
            return [ h, s * 100, l * 100 ];
        };
        convert.rgb.hsv = function(rgb) {
            let rdif;
            let gdif;
            let bdif;
            let h;
            let s;
            const r = rgb[0] / 255;
            const g = rgb[1] / 255;
            const b = rgb[2] / 255;
            const v = Math.max(r, g, b);
            const diff = v - Math.min(r, g, b);
            const diffc = function(c) {
                return (v - c) / 6 / diff + 1 / 2;
            };
            if (diff === 0) {
                h = 0;
                s = 0;
            } else {
                s = diff / v;
                rdif = diffc(r);
                gdif = diffc(g);
                bdif = diffc(b);
                if (r === v) {
                    h = bdif - gdif;
                } else if (g === v) {
                    h = 1 / 3 + rdif - bdif;
                } else if (b === v) {
                    h = 2 / 3 + gdif - rdif;
                }
                if (h < 0) {
                    h += 1;
                } else if (h > 1) {
                    h -= 1;
                }
            }
            return [ h * 360, s * 100, v * 100 ];
        };
        convert.rgb.hwb = function(rgb) {
            const r = rgb[0];
            const g = rgb[1];
            let b = rgb[2];
            const h = convert.rgb.hsl(rgb)[0];
            const w = 1 / 255 * Math.min(r, Math.min(g, b));
            b = 1 - 1 / 255 * Math.max(r, Math.max(g, b));
            return [ h, w * 100, b * 100 ];
        };
        convert.rgb.cmyk = function(rgb) {
            const r = rgb[0] / 255;
            const g = rgb[1] / 255;
            const b = rgb[2] / 255;
            const k = Math.min(1 - r, 1 - g, 1 - b);
            const c = (1 - r - k) / (1 - k) || 0;
            const m = (1 - g - k) / (1 - k) || 0;
            const y = (1 - b - k) / (1 - k) || 0;
            return [ c * 100, m * 100, y * 100, k * 100 ];
        };
        function comparativeDistance(x, y) {
            return (x[0] - y[0]) ** 2 + (x[1] - y[1]) ** 2 + (x[2] - y[2]) ** 2;
        }
        convert.rgb.keyword = function(rgb) {
            const reversed = reverseKeywords[rgb];
            if (reversed) {
                return reversed;
            }
            let currentClosestDistance = Infinity;
            let currentClosestKeyword;
            for (const keyword of Object.keys(cssKeywords)) {
                const value = cssKeywords[keyword];
                const distance = comparativeDistance(rgb, value);
                if (distance < currentClosestDistance) {
                    currentClosestDistance = distance;
                    currentClosestKeyword = keyword;
                }
            }
            return currentClosestKeyword;
        };
        convert.keyword.rgb = function(keyword) {
            return cssKeywords[keyword];
        };
        convert.rgb.xyz = function(rgb) {
            let r = rgb[0] / 255;
            let g = rgb[1] / 255;
            let b = rgb[2] / 255;
            r = r > .04045 ? ((r + .055) / 1.055) ** 2.4 : r / 12.92;
            g = g > .04045 ? ((g + .055) / 1.055) ** 2.4 : g / 12.92;
            b = b > .04045 ? ((b + .055) / 1.055) ** 2.4 : b / 12.92;
            const x = r * .4124 + g * .3576 + b * .1805;
            const y = r * .2126 + g * .7152 + b * .0722;
            const z = r * .0193 + g * .1192 + b * .9505;
            return [ x * 100, y * 100, z * 100 ];
        };
        convert.rgb.lab = function(rgb) {
            const xyz = convert.rgb.xyz(rgb);
            let x = xyz[0];
            let y = xyz[1];
            let z = xyz[2];
            x /= 95.047;
            y /= 100;
            z /= 108.883;
            x = x > .008856 ? x ** (1 / 3) : 7.787 * x + 16 / 116;
            y = y > .008856 ? y ** (1 / 3) : 7.787 * y + 16 / 116;
            z = z > .008856 ? z ** (1 / 3) : 7.787 * z + 16 / 116;
            const l = 116 * y - 16;
            const a = 500 * (x - y);
            const b = 200 * (y - z);
            return [ l, a, b ];
        };
        convert.hsl.rgb = function(hsl) {
            const h = hsl[0] / 360;
            const s = hsl[1] / 100;
            const l = hsl[2] / 100;
            let t2;
            let t3;
            let val;
            if (s === 0) {
                val = l * 255;
                return [ val, val, val ];
            }
            if (l < .5) {
                t2 = l * (1 + s);
            } else {
                t2 = l + s - l * s;
            }
            const t1 = 2 * l - t2;
            const rgb = [ 0, 0, 0 ];
            for (let i = 0; i < 3; i++) {
                t3 = h + 1 / 3 * -(i - 1);
                if (t3 < 0) {
                    t3++;
                }
                if (t3 > 1) {
                    t3--;
                }
                if (6 * t3 < 1) {
                    val = t1 + (t2 - t1) * 6 * t3;
                } else if (2 * t3 < 1) {
                    val = t2;
                } else if (3 * t3 < 2) {
                    val = t1 + (t2 - t1) * (2 / 3 - t3) * 6;
                } else {
                    val = t1;
                }
                rgb[i] = val * 255;
            }
            return rgb;
        };
        convert.hsl.hsv = function(hsl) {
            const h = hsl[0];
            let s = hsl[1] / 100;
            let l = hsl[2] / 100;
            let smin = s;
            const lmin = Math.max(l, .01);
            l *= 2;
            s *= l <= 1 ? l : 2 - l;
            smin *= lmin <= 1 ? lmin : 2 - lmin;
            const v = (l + s) / 2;
            const sv = l === 0 ? 2 * smin / (lmin + smin) : 2 * s / (l + s);
            return [ h, sv * 100, v * 100 ];
        };
        convert.hsv.rgb = function(hsv) {
            const h = hsv[0] / 60;
            const s = hsv[1] / 100;
            let v = hsv[2] / 100;
            const hi = Math.floor(h) % 6;
            const f = h - Math.floor(h);
            const p = 255 * v * (1 - s);
            const q = 255 * v * (1 - s * f);
            const t = 255 * v * (1 - s * (1 - f));
            v *= 255;
            switch (hi) {
              case 0:
                return [ v, t, p ];

              case 1:
                return [ q, v, p ];

              case 2:
                return [ p, v, t ];

              case 3:
                return [ p, q, v ];

              case 4:
                return [ t, p, v ];

              case 5:
                return [ v, p, q ];
            }
        };
        convert.hsv.hsl = function(hsv) {
            const h = hsv[0];
            const s = hsv[1] / 100;
            const v = hsv[2] / 100;
            const vmin = Math.max(v, .01);
            let sl;
            let l;
            l = (2 - s) * v;
            const lmin = (2 - s) * vmin;
            sl = s * vmin;
            sl /= lmin <= 1 ? lmin : 2 - lmin;
            sl = sl || 0;
            l /= 2;
            return [ h, sl * 100, l * 100 ];
        };
        convert.hwb.rgb = function(hwb) {
            const h = hwb[0] / 360;
            let wh = hwb[1] / 100;
            let bl = hwb[2] / 100;
            const ratio = wh + bl;
            let f;
            if (ratio > 1) {
                wh /= ratio;
                bl /= ratio;
            }
            const i = Math.floor(6 * h);
            const v = 1 - bl;
            f = 6 * h - i;
            if ((i & 1) !== 0) {
                f = 1 - f;
            }
            const n = wh + f * (v - wh);
            let r;
            let g;
            let b;
            switch (i) {
              default:
              case 6:
              case 0:
                r = v;
                g = n;
                b = wh;
                break;

              case 1:
                r = n;
                g = v;
                b = wh;
                break;

              case 2:
                r = wh;
                g = v;
                b = n;
                break;

              case 3:
                r = wh;
                g = n;
                b = v;
                break;

              case 4:
                r = n;
                g = wh;
                b = v;
                break;

              case 5:
                r = v;
                g = wh;
                b = n;
                break;
            }
            return [ r * 255, g * 255, b * 255 ];
        };
        convert.cmyk.rgb = function(cmyk) {
            const c = cmyk[0] / 100;
            const m = cmyk[1] / 100;
            const y = cmyk[2] / 100;
            const k = cmyk[3] / 100;
            const r = 1 - Math.min(1, c * (1 - k) + k);
            const g = 1 - Math.min(1, m * (1 - k) + k);
            const b = 1 - Math.min(1, y * (1 - k) + k);
            return [ r * 255, g * 255, b * 255 ];
        };
        convert.xyz.rgb = function(xyz) {
            const x = xyz[0] / 100;
            const y = xyz[1] / 100;
            const z = xyz[2] / 100;
            let r;
            let g;
            let b;
            r = x * 3.2406 + y * -1.5372 + z * -.4986;
            g = x * -.9689 + y * 1.8758 + z * .0415;
            b = x * .0557 + y * -.204 + z * 1.057;
            r = r > .0031308 ? 1.055 * r ** (1 / 2.4) - .055 : r * 12.92;
            g = g > .0031308 ? 1.055 * g ** (1 / 2.4) - .055 : g * 12.92;
            b = b > .0031308 ? 1.055 * b ** (1 / 2.4) - .055 : b * 12.92;
            r = Math.min(Math.max(0, r), 1);
            g = Math.min(Math.max(0, g), 1);
            b = Math.min(Math.max(0, b), 1);
            return [ r * 255, g * 255, b * 255 ];
        };
        convert.xyz.lab = function(xyz) {
            let x = xyz[0];
            let y = xyz[1];
            let z = xyz[2];
            x /= 95.047;
            y /= 100;
            z /= 108.883;
            x = x > .008856 ? x ** (1 / 3) : 7.787 * x + 16 / 116;
            y = y > .008856 ? y ** (1 / 3) : 7.787 * y + 16 / 116;
            z = z > .008856 ? z ** (1 / 3) : 7.787 * z + 16 / 116;
            const l = 116 * y - 16;
            const a = 500 * (x - y);
            const b = 200 * (y - z);
            return [ l, a, b ];
        };
        convert.lab.xyz = function(lab) {
            const l = lab[0];
            const a = lab[1];
            const b = lab[2];
            let x;
            let y;
            let z;
            y = (l + 16) / 116;
            x = a / 500 + y;
            z = y - b / 200;
            const y2 = y ** 3;
            const x2 = x ** 3;
            const z2 = z ** 3;
            y = y2 > .008856 ? y2 : (y - 16 / 116) / 7.787;
            x = x2 > .008856 ? x2 : (x - 16 / 116) / 7.787;
            z = z2 > .008856 ? z2 : (z - 16 / 116) / 7.787;
            x *= 95.047;
            y *= 100;
            z *= 108.883;
            return [ x, y, z ];
        };
        convert.lab.lch = function(lab) {
            const l = lab[0];
            const a = lab[1];
            const b = lab[2];
            let h;
            const hr = Math.atan2(b, a);
            h = hr * 360 / 2 / Math.PI;
            if (h < 0) {
                h += 360;
            }
            const c = Math.sqrt(a * a + b * b);
            return [ l, c, h ];
        };
        convert.lch.lab = function(lch) {
            const l = lch[0];
            const c = lch[1];
            const h = lch[2];
            const hr = h / 360 * 2 * Math.PI;
            const a = c * Math.cos(hr);
            const b = c * Math.sin(hr);
            return [ l, a, b ];
        };
        convert.rgb.ansi16 = function(args, saturation = null) {
            const [r, g, b] = args;
            let value = saturation === null ? convert.rgb.hsv(args)[2] : saturation;
            value = Math.round(value / 50);
            if (value === 0) {
                return 30;
            }
            let ansi = 30 + (Math.round(b / 255) << 2 | Math.round(g / 255) << 1 | Math.round(r / 255));
            if (value === 2) {
                ansi += 60;
            }
            return ansi;
        };
        convert.hsv.ansi16 = function(args) {
            return convert.rgb.ansi16(convert.hsv.rgb(args), args[2]);
        };
        convert.rgb.ansi256 = function(args) {
            const r = args[0];
            const g = args[1];
            const b = args[2];
            if (r === g && g === b) {
                if (r < 8) {
                    return 16;
                }
                if (r > 248) {
                    return 231;
                }
                return Math.round((r - 8) / 247 * 24) + 232;
            }
            const ansi = 16 + 36 * Math.round(r / 255 * 5) + 6 * Math.round(g / 255 * 5) + Math.round(b / 255 * 5);
            return ansi;
        };
        convert.ansi16.rgb = function(args) {
            let color = args % 10;
            if (color === 0 || color === 7) {
                if (args > 50) {
                    color += 3.5;
                }
                color = color / 10.5 * 255;
                return [ color, color, color ];
            }
            const mult = (~~(args > 50) + 1) * .5;
            const r = (color & 1) * mult * 255;
            const g = (color >> 1 & 1) * mult * 255;
            const b = (color >> 2 & 1) * mult * 255;
            return [ r, g, b ];
        };
        convert.ansi256.rgb = function(args) {
            if (args >= 232) {
                const c = (args - 232) * 10 + 8;
                return [ c, c, c ];
            }
            args -= 16;
            let rem;
            const r = Math.floor(args / 36) / 5 * 255;
            const g = Math.floor((rem = args % 36) / 6) / 5 * 255;
            const b = rem % 6 / 5 * 255;
            return [ r, g, b ];
        };
        convert.rgb.hex = function(args) {
            const integer = ((Math.round(args[0]) & 255) << 16) + ((Math.round(args[1]) & 255) << 8) + (Math.round(args[2]) & 255);
            const string = integer.toString(16).toUpperCase();
            return "000000".substring(string.length) + string;
        };
        convert.hex.rgb = function(args) {
            const match = args.toString(16).match(/[a-f0-9]{6}|[a-f0-9]{3}/i);
            if (!match) {
                return [ 0, 0, 0 ];
            }
            let colorString = match[0];
            if (match[0].length === 3) {
                colorString = colorString.split("").map((char => char + char)).join("");
            }
            const integer = parseInt(colorString, 16);
            const r = integer >> 16 & 255;
            const g = integer >> 8 & 255;
            const b = integer & 255;
            return [ r, g, b ];
        };
        convert.rgb.hcg = function(rgb) {
            const r = rgb[0] / 255;
            const g = rgb[1] / 255;
            const b = rgb[2] / 255;
            const max = Math.max(Math.max(r, g), b);
            const min = Math.min(Math.min(r, g), b);
            const chroma = max - min;
            let grayscale;
            let hue;
            if (chroma < 1) {
                grayscale = min / (1 - chroma);
            } else {
                grayscale = 0;
            }
            if (chroma <= 0) {
                hue = 0;
            } else if (max === r) {
                hue = (g - b) / chroma % 6;
            } else if (max === g) {
                hue = 2 + (b - r) / chroma;
            } else {
                hue = 4 + (r - g) / chroma;
            }
            hue /= 6;
            hue %= 1;
            return [ hue * 360, chroma * 100, grayscale * 100 ];
        };
        convert.hsl.hcg = function(hsl) {
            const s = hsl[1] / 100;
            const l = hsl[2] / 100;
            const c = l < .5 ? 2 * s * l : 2 * s * (1 - l);
            let f = 0;
            if (c < 1) {
                f = (l - .5 * c) / (1 - c);
            }
            return [ hsl[0], c * 100, f * 100 ];
        };
        convert.hsv.hcg = function(hsv) {
            const s = hsv[1] / 100;
            const v = hsv[2] / 100;
            const c = s * v;
            let f = 0;
            if (c < 1) {
                f = (v - c) / (1 - c);
            }
            return [ hsv[0], c * 100, f * 100 ];
        };
        convert.hcg.rgb = function(hcg) {
            const h = hcg[0] / 360;
            const c = hcg[1] / 100;
            const g = hcg[2] / 100;
            if (c === 0) {
                return [ g * 255, g * 255, g * 255 ];
            }
            const pure = [ 0, 0, 0 ];
            const hi = h % 1 * 6;
            const v = hi % 1;
            const w = 1 - v;
            let mg = 0;
            switch (Math.floor(hi)) {
              case 0:
                pure[0] = 1;
                pure[1] = v;
                pure[2] = 0;
                break;

              case 1:
                pure[0] = w;
                pure[1] = 1;
                pure[2] = 0;
                break;

              case 2:
                pure[0] = 0;
                pure[1] = 1;
                pure[2] = v;
                break;

              case 3:
                pure[0] = 0;
                pure[1] = w;
                pure[2] = 1;
                break;

              case 4:
                pure[0] = v;
                pure[1] = 0;
                pure[2] = 1;
                break;

              default:
                pure[0] = 1;
                pure[1] = 0;
                pure[2] = w;
            }
            mg = (1 - c) * g;
            return [ (c * pure[0] + mg) * 255, (c * pure[1] + mg) * 255, (c * pure[2] + mg) * 255 ];
        };
        convert.hcg.hsv = function(hcg) {
            const c = hcg[1] / 100;
            const g = hcg[2] / 100;
            const v = c + g * (1 - c);
            let f = 0;
            if (v > 0) {
                f = c / v;
            }
            return [ hcg[0], f * 100, v * 100 ];
        };
        convert.hcg.hsl = function(hcg) {
            const c = hcg[1] / 100;
            const g = hcg[2] / 100;
            const l = g * (1 - c) + .5 * c;
            let s = 0;
            if (l > 0 && l < .5) {
                s = c / (2 * l);
            } else if (l >= .5 && l < 1) {
                s = c / (2 * (1 - l));
            }
            return [ hcg[0], s * 100, l * 100 ];
        };
        convert.hcg.hwb = function(hcg) {
            const c = hcg[1] / 100;
            const g = hcg[2] / 100;
            const v = c + g * (1 - c);
            return [ hcg[0], (v - c) * 100, (1 - v) * 100 ];
        };
        convert.hwb.hcg = function(hwb) {
            const w = hwb[1] / 100;
            const b = hwb[2] / 100;
            const v = 1 - b;
            const c = v - w;
            let g = 0;
            if (c < 1) {
                g = (v - c) / (1 - c);
            }
            return [ hwb[0], c * 100, g * 100 ];
        };
        convert.apple.rgb = function(apple) {
            return [ apple[0] / 65535 * 255, apple[1] / 65535 * 255, apple[2] / 65535 * 255 ];
        };
        convert.rgb.apple = function(rgb) {
            return [ rgb[0] / 255 * 65535, rgb[1] / 255 * 65535, rgb[2] / 255 * 65535 ];
        };
        convert.gray.rgb = function(args) {
            return [ args[0] / 100 * 255, args[0] / 100 * 255, args[0] / 100 * 255 ];
        };
        convert.gray.hsl = function(args) {
            return [ 0, 0, args[0] ];
        };
        convert.gray.hsv = convert.gray.hsl;
        convert.gray.hwb = function(gray) {
            return [ 0, 100, gray[0] ];
        };
        convert.gray.cmyk = function(gray) {
            return [ 0, 0, 0, gray[0] ];
        };
        convert.gray.lab = function(gray) {
            return [ gray[0], 0, 0 ];
        };
        convert.gray.hex = function(gray) {
            const val = Math.round(gray[0] / 100 * 255) & 255;
            const integer = (val << 16) + (val << 8) + val;
            const string = integer.toString(16).toUpperCase();
            return "000000".substring(string.length) + string;
        };
        convert.rgb.gray = function(rgb) {
            const val = (rgb[0] + rgb[1] + rgb[2]) / 3;
            return [ val / 255 * 100 ];
        };
    },
    9208: (module, __unused_webpack_exports, __webpack_require__) => {
        const conversions = __webpack_require__(2538);
        const route = __webpack_require__(2051);
        const convert = {};
        const models = Object.keys(conversions);
        function wrapRaw(fn) {
            const wrappedFn = function(...args) {
                const arg0 = args[0];
                if (arg0 === undefined || arg0 === null) {
                    return arg0;
                }
                if (arg0.length > 1) {
                    args = arg0;
                }
                return fn(args);
            };
            if ("conversion" in fn) {
                wrappedFn.conversion = fn.conversion;
            }
            return wrappedFn;
        }
        function wrapRounded(fn) {
            const wrappedFn = function(...args) {
                const arg0 = args[0];
                if (arg0 === undefined || arg0 === null) {
                    return arg0;
                }
                if (arg0.length > 1) {
                    args = arg0;
                }
                const result = fn(args);
                if (typeof result === "object") {
                    for (let len = result.length, i = 0; i < len; i++) {
                        result[i] = Math.round(result[i]);
                    }
                }
                return result;
            };
            if ("conversion" in fn) {
                wrappedFn.conversion = fn.conversion;
            }
            return wrappedFn;
        }
        models.forEach((fromModel => {
            convert[fromModel] = {};
            Object.defineProperty(convert[fromModel], "channels", {
                value: conversions[fromModel].channels
            });
            Object.defineProperty(convert[fromModel], "labels", {
                value: conversions[fromModel].labels
            });
            const routes = route(fromModel);
            const routeModels = Object.keys(routes);
            routeModels.forEach((toModel => {
                const fn = routes[toModel];
                convert[fromModel][toModel] = wrapRounded(fn);
                convert[fromModel][toModel].raw = wrapRaw(fn);
            }));
        }));
        module.exports = convert;
    },
    2051: (module, __unused_webpack_exports, __webpack_require__) => {
        const conversions = __webpack_require__(2538);
        function buildGraph() {
            const graph = {};
            const models = Object.keys(conversions);
            for (let len = models.length, i = 0; i < len; i++) {
                graph[models[i]] = {
                    distance: -1,
                    parent: null
                };
            }
            return graph;
        }
        function deriveBFS(fromModel) {
            const graph = buildGraph();
            const queue = [ fromModel ];
            graph[fromModel].distance = 0;
            while (queue.length) {
                const current = queue.pop();
                const adjacents = Object.keys(conversions[current]);
                for (let len = adjacents.length, i = 0; i < len; i++) {
                    const adjacent = adjacents[i];
                    const node = graph[adjacent];
                    if (node.distance === -1) {
                        node.distance = graph[current].distance + 1;
                        node.parent = current;
                        queue.unshift(adjacent);
                    }
                }
            }
            return graph;
        }
        function link(from, to) {
            return function(args) {
                return to(from(args));
            };
        }
        function wrapConversion(toModel, graph) {
            const path = [ graph[toModel].parent, toModel ];
            let fn = conversions[graph[toModel].parent][toModel];
            let cur = graph[toModel].parent;
            while (graph[cur].parent) {
                path.unshift(graph[cur].parent);
                fn = link(conversions[graph[cur].parent][cur], fn);
                cur = graph[cur].parent;
            }
            fn.conversion = path;
            return fn;
        }
        module.exports = function(fromModel) {
            const graph = deriveBFS(fromModel);
            const conversion = {};
            const models = Object.keys(graph);
            for (let len = models.length, i = 0; i < len; i++) {
                const toModel = models[i];
                const node = graph[toModel];
                if (node.parent === null) {
                    continue;
                }
                conversion[toModel] = wrapConversion(toModel, graph);
            }
            return conversion;
        };
    },
    6150: module => {
        "use strict";
        module.exports = {
            aliceblue: [ 240, 248, 255 ],
            antiquewhite: [ 250, 235, 215 ],
            aqua: [ 0, 255, 255 ],
            aquamarine: [ 127, 255, 212 ],
            azure: [ 240, 255, 255 ],
            beige: [ 245, 245, 220 ],
            bisque: [ 255, 228, 196 ],
            black: [ 0, 0, 0 ],
            blanchedalmond: [ 255, 235, 205 ],
            blue: [ 0, 0, 255 ],
            blueviolet: [ 138, 43, 226 ],
            brown: [ 165, 42, 42 ],
            burlywood: [ 222, 184, 135 ],
            cadetblue: [ 95, 158, 160 ],
            chartreuse: [ 127, 255, 0 ],
            chocolate: [ 210, 105, 30 ],
            coral: [ 255, 127, 80 ],
            cornflowerblue: [ 100, 149, 237 ],
            cornsilk: [ 255, 248, 220 ],
            crimson: [ 220, 20, 60 ],
            cyan: [ 0, 255, 255 ],
            darkblue: [ 0, 0, 139 ],
            darkcyan: [ 0, 139, 139 ],
            darkgoldenrod: [ 184, 134, 11 ],
            darkgray: [ 169, 169, 169 ],
            darkgreen: [ 0, 100, 0 ],
            darkgrey: [ 169, 169, 169 ],
            darkkhaki: [ 189, 183, 107 ],
            darkmagenta: [ 139, 0, 139 ],
            darkolivegreen: [ 85, 107, 47 ],
            darkorange: [ 255, 140, 0 ],
            darkorchid: [ 153, 50, 204 ],
            darkred: [ 139, 0, 0 ],
            darksalmon: [ 233, 150, 122 ],
            darkseagreen: [ 143, 188, 143 ],
            darkslateblue: [ 72, 61, 139 ],
            darkslategray: [ 47, 79, 79 ],
            darkslategrey: [ 47, 79, 79 ],
            darkturquoise: [ 0, 206, 209 ],
            darkviolet: [ 148, 0, 211 ],
            deeppink: [ 255, 20, 147 ],
            deepskyblue: [ 0, 191, 255 ],
            dimgray: [ 105, 105, 105 ],
            dimgrey: [ 105, 105, 105 ],
            dodgerblue: [ 30, 144, 255 ],
            firebrick: [ 178, 34, 34 ],
            floralwhite: [ 255, 250, 240 ],
            forestgreen: [ 34, 139, 34 ],
            fuchsia: [ 255, 0, 255 ],
            gainsboro: [ 220, 220, 220 ],
            ghostwhite: [ 248, 248, 255 ],
            gold: [ 255, 215, 0 ],
            goldenrod: [ 218, 165, 32 ],
            gray: [ 128, 128, 128 ],
            green: [ 0, 128, 0 ],
            greenyellow: [ 173, 255, 47 ],
            grey: [ 128, 128, 128 ],
            honeydew: [ 240, 255, 240 ],
            hotpink: [ 255, 105, 180 ],
            indianred: [ 205, 92, 92 ],
            indigo: [ 75, 0, 130 ],
            ivory: [ 255, 255, 240 ],
            khaki: [ 240, 230, 140 ],
            lavender: [ 230, 230, 250 ],
            lavenderblush: [ 255, 240, 245 ],
            lawngreen: [ 124, 252, 0 ],
            lemonchiffon: [ 255, 250, 205 ],
            lightblue: [ 173, 216, 230 ],
            lightcoral: [ 240, 128, 128 ],
            lightcyan: [ 224, 255, 255 ],
            lightgoldenrodyellow: [ 250, 250, 210 ],
            lightgray: [ 211, 211, 211 ],
            lightgreen: [ 144, 238, 144 ],
            lightgrey: [ 211, 211, 211 ],
            lightpink: [ 255, 182, 193 ],
            lightsalmon: [ 255, 160, 122 ],
            lightseagreen: [ 32, 178, 170 ],
            lightskyblue: [ 135, 206, 250 ],
            lightslategray: [ 119, 136, 153 ],
            lightslategrey: [ 119, 136, 153 ],
            lightsteelblue: [ 176, 196, 222 ],
            lightyellow: [ 255, 255, 224 ],
            lime: [ 0, 255, 0 ],
            limegreen: [ 50, 205, 50 ],
            linen: [ 250, 240, 230 ],
            magenta: [ 255, 0, 255 ],
            maroon: [ 128, 0, 0 ],
            mediumaquamarine: [ 102, 205, 170 ],
            mediumblue: [ 0, 0, 205 ],
            mediumorchid: [ 186, 85, 211 ],
            mediumpurple: [ 147, 112, 219 ],
            mediumseagreen: [ 60, 179, 113 ],
            mediumslateblue: [ 123, 104, 238 ],
            mediumspringgreen: [ 0, 250, 154 ],
            mediumturquoise: [ 72, 209, 204 ],
            mediumvioletred: [ 199, 21, 133 ],
            midnightblue: [ 25, 25, 112 ],
            mintcream: [ 245, 255, 250 ],
            mistyrose: [ 255, 228, 225 ],
            moccasin: [ 255, 228, 181 ],
            navajowhite: [ 255, 222, 173 ],
            navy: [ 0, 0, 128 ],
            oldlace: [ 253, 245, 230 ],
            olive: [ 128, 128, 0 ],
            olivedrab: [ 107, 142, 35 ],
            orange: [ 255, 165, 0 ],
            orangered: [ 255, 69, 0 ],
            orchid: [ 218, 112, 214 ],
            palegoldenrod: [ 238, 232, 170 ],
            palegreen: [ 152, 251, 152 ],
            paleturquoise: [ 175, 238, 238 ],
            palevioletred: [ 219, 112, 147 ],
            papayawhip: [ 255, 239, 213 ],
            peachpuff: [ 255, 218, 185 ],
            peru: [ 205, 133, 63 ],
            pink: [ 255, 192, 203 ],
            plum: [ 221, 160, 221 ],
            powderblue: [ 176, 224, 230 ],
            purple: [ 128, 0, 128 ],
            rebeccapurple: [ 102, 51, 153 ],
            red: [ 255, 0, 0 ],
            rosybrown: [ 188, 143, 143 ],
            royalblue: [ 65, 105, 225 ],
            saddlebrown: [ 139, 69, 19 ],
            salmon: [ 250, 128, 114 ],
            sandybrown: [ 244, 164, 96 ],
            seagreen: [ 46, 139, 87 ],
            seashell: [ 255, 245, 238 ],
            sienna: [ 160, 82, 45 ],
            silver: [ 192, 192, 192 ],
            skyblue: [ 135, 206, 235 ],
            slateblue: [ 106, 90, 205 ],
            slategray: [ 112, 128, 144 ],
            slategrey: [ 112, 128, 144 ],
            snow: [ 255, 250, 250 ],
            springgreen: [ 0, 255, 127 ],
            steelblue: [ 70, 130, 180 ],
            tan: [ 210, 180, 140 ],
            teal: [ 0, 128, 128 ],
            thistle: [ 216, 191, 216 ],
            tomato: [ 255, 99, 71 ],
            turquoise: [ 64, 224, 208 ],
            violet: [ 238, 130, 238 ],
            wheat: [ 245, 222, 179 ],
            white: [ 255, 255, 255 ],
            whitesmoke: [ 245, 245, 245 ],
            yellow: [ 255, 255, 0 ],
            yellowgreen: [ 154, 205, 50 ]
        };
    },
    4288: module => {
        "use strict";
        module.exports = (flag, argv = process.argv) => {
            const prefix = flag.startsWith("-") ? "" : flag.length === 1 ? "-" : "--";
            const position = argv.indexOf(prefix + flag);
            const terminatorPosition = argv.indexOf("--");
            return position !== -1 && (terminatorPosition === -1 || position < terminatorPosition);
        };
    },
    7706: (module, __unused_webpack_exports, __webpack_require__) => {
        const ANY = Symbol("SemVer ANY");
        class Comparator {
            static get ANY() {
                return ANY;
            }
            constructor(comp, options) {
                options = parseOptions(options);
                if (comp instanceof Comparator) {
                    if (comp.loose === !!options.loose) {
                        return comp;
                    } else {
                        comp = comp.value;
                    }
                }
                comp = comp.trim().split(/\s+/).join(" ");
                debug("comparator", comp, options);
                this.options = options;
                this.loose = !!options.loose;
                this.parse(comp);
                if (this.semver === ANY) {
                    this.value = "";
                } else {
                    this.value = this.operator + this.semver.version;
                }
                debug("comp", this);
            }
            parse(comp) {
                const r = this.options.loose ? re[t.COMPARATORLOOSE] : re[t.COMPARATOR];
                const m = comp.match(r);
                if (!m) {
                    throw new TypeError(`Invalid comparator: ${comp}`);
                }
                this.operator = m[1] !== undefined ? m[1] : "";
                if (this.operator === "=") {
                    this.operator = "";
                }
                if (!m[2]) {
                    this.semver = ANY;
                } else {
                    this.semver = new SemVer(m[2], this.options.loose);
                }
            }
            toString() {
                return this.value;
            }
            test(version) {
                debug("Comparator.test", version, this.options.loose);
                if (this.semver === ANY || version === ANY) {
                    return true;
                }
                if (typeof version === "string") {
                    try {
                        version = new SemVer(version, this.options);
                    } catch (er) {
                        return false;
                    }
                }
                return cmp(version, this.operator, this.semver, this.options);
            }
            intersects(comp, options) {
                if (!(comp instanceof Comparator)) {
                    throw new TypeError("a Comparator is required");
                }
                if (this.operator === "") {
                    if (this.value === "") {
                        return true;
                    }
                    return new Range(comp.value, options).test(this.value);
                } else if (comp.operator === "") {
                    if (comp.value === "") {
                        return true;
                    }
                    return new Range(this.value, options).test(comp.semver);
                }
                options = parseOptions(options);
                if (options.includePrerelease && (this.value === "<0.0.0-0" || comp.value === "<0.0.0-0")) {
                    return false;
                }
                if (!options.includePrerelease && (this.value.startsWith("<0.0.0") || comp.value.startsWith("<0.0.0"))) {
                    return false;
                }
                if (this.operator.startsWith(">") && comp.operator.startsWith(">")) {
                    return true;
                }
                if (this.operator.startsWith("<") && comp.operator.startsWith("<")) {
                    return true;
                }
                if (this.semver.version === comp.semver.version && this.operator.includes("=") && comp.operator.includes("=")) {
                    return true;
                }
                if (cmp(this.semver, "<", comp.semver, options) && this.operator.startsWith(">") && comp.operator.startsWith("<")) {
                    return true;
                }
                if (cmp(this.semver, ">", comp.semver, options) && this.operator.startsWith("<") && comp.operator.startsWith(">")) {
                    return true;
                }
                return false;
            }
        }
        module.exports = Comparator;
        const parseOptions = __webpack_require__(3867);
        const {safeRe: re, t} = __webpack_require__(9541);
        const cmp = __webpack_require__(1918);
        const debug = __webpack_require__(5432);
        const SemVer = __webpack_require__(3013);
        const Range = __webpack_require__(6833);
    },
    6833: (module, __unused_webpack_exports, __webpack_require__) => {
        class Range {
            constructor(range, options) {
                options = parseOptions(options);
                if (range instanceof Range) {
                    if (range.loose === !!options.loose && range.includePrerelease === !!options.includePrerelease) {
                        return range;
                    } else {
                        return new Range(range.raw, options);
                    }
                }
                if (range instanceof Comparator) {
                    this.raw = range.value;
                    this.set = [ [ range ] ];
                    this.format();
                    return this;
                }
                this.options = options;
                this.loose = !!options.loose;
                this.includePrerelease = !!options.includePrerelease;
                this.raw = range.trim().split(/\s+/).join(" ");
                this.set = this.raw.split("||").map((r => this.parseRange(r.trim()))).filter((c => c.length));
                if (!this.set.length) {
                    throw new TypeError(`Invalid SemVer Range: ${this.raw}`);
                }
                if (this.set.length > 1) {
                    const first = this.set[0];
                    this.set = this.set.filter((c => !isNullSet(c[0])));
                    if (this.set.length === 0) {
                        this.set = [ first ];
                    } else if (this.set.length > 1) {
                        for (const c of this.set) {
                            if (c.length === 1 && isAny(c[0])) {
                                this.set = [ c ];
                                break;
                            }
                        }
                    }
                }
                this.format();
            }
            format() {
                this.range = this.set.map((comps => comps.join(" ").trim())).join("||").trim();
                return this.range;
            }
            toString() {
                return this.range;
            }
            parseRange(range) {
                const memoOpts = (this.options.includePrerelease && FLAG_INCLUDE_PRERELEASE) | (this.options.loose && FLAG_LOOSE);
                const memoKey = memoOpts + ":" + range;
                const cached = cache.get(memoKey);
                if (cached) {
                    return cached;
                }
                const loose = this.options.loose;
                const hr = loose ? re[t.HYPHENRANGELOOSE] : re[t.HYPHENRANGE];
                range = range.replace(hr, hyphenReplace(this.options.includePrerelease));
                debug("hyphen replace", range);
                range = range.replace(re[t.COMPARATORTRIM], comparatorTrimReplace);
                debug("comparator trim", range);
                range = range.replace(re[t.TILDETRIM], tildeTrimReplace);
                debug("tilde trim", range);
                range = range.replace(re[t.CARETTRIM], caretTrimReplace);
                debug("caret trim", range);
                let rangeList = range.split(" ").map((comp => parseComparator(comp, this.options))).join(" ").split(/\s+/).map((comp => replaceGTE0(comp, this.options)));
                if (loose) {
                    rangeList = rangeList.filter((comp => {
                        debug("loose invalid filter", comp, this.options);
                        return !!comp.match(re[t.COMPARATORLOOSE]);
                    }));
                }
                debug("range list", rangeList);
                const rangeMap = new Map;
                const comparators = rangeList.map((comp => new Comparator(comp, this.options)));
                for (const comp of comparators) {
                    if (isNullSet(comp)) {
                        return [ comp ];
                    }
                    rangeMap.set(comp.value, comp);
                }
                if (rangeMap.size > 1 && rangeMap.has("")) {
                    rangeMap.delete("");
                }
                const result = [ ...rangeMap.values() ];
                cache.set(memoKey, result);
                return result;
            }
            intersects(range, options) {
                if (!(range instanceof Range)) {
                    throw new TypeError("a Range is required");
                }
                return this.set.some((thisComparators => isSatisfiable(thisComparators, options) && range.set.some((rangeComparators => isSatisfiable(rangeComparators, options) && thisComparators.every((thisComparator => rangeComparators.every((rangeComparator => thisComparator.intersects(rangeComparator, options)))))))));
            }
            test(version) {
                if (!version) {
                    return false;
                }
                if (typeof version === "string") {
                    try {
                        version = new SemVer(version, this.options);
                    } catch (er) {
                        return false;
                    }
                }
                for (let i = 0; i < this.set.length; i++) {
                    if (testSet(this.set[i], version, this.options)) {
                        return true;
                    }
                }
                return false;
            }
        }
        module.exports = Range;
        const LRU = __webpack_require__(6923);
        const cache = new LRU({
            max: 1e3
        });
        const parseOptions = __webpack_require__(3867);
        const Comparator = __webpack_require__(7706);
        const debug = __webpack_require__(5432);
        const SemVer = __webpack_require__(3013);
        const {safeRe: re, t, comparatorTrimReplace, tildeTrimReplace, caretTrimReplace} = __webpack_require__(9541);
        const {FLAG_INCLUDE_PRERELEASE, FLAG_LOOSE} = __webpack_require__(9041);
        const isNullSet = c => c.value === "<0.0.0-0";
        const isAny = c => c.value === "";
        const isSatisfiable = (comparators, options) => {
            let result = true;
            const remainingComparators = comparators.slice();
            let testComparator = remainingComparators.pop();
            while (result && remainingComparators.length) {
                result = remainingComparators.every((otherComparator => testComparator.intersects(otherComparator, options)));
                testComparator = remainingComparators.pop();
            }
            return result;
        };
        const parseComparator = (comp, options) => {
            debug("comp", comp, options);
            comp = replaceCarets(comp, options);
            debug("caret", comp);
            comp = replaceTildes(comp, options);
            debug("tildes", comp);
            comp = replaceXRanges(comp, options);
            debug("xrange", comp);
            comp = replaceStars(comp, options);
            debug("stars", comp);
            return comp;
        };
        const isX = id => !id || id.toLowerCase() === "x" || id === "*";
        const replaceTildes = (comp, options) => comp.trim().split(/\s+/).map((c => replaceTilde(c, options))).join(" ");
        const replaceTilde = (comp, options) => {
            const r = options.loose ? re[t.TILDELOOSE] : re[t.TILDE];
            return comp.replace(r, ((_, M, m, p, pr) => {
                debug("tilde", comp, _, M, m, p, pr);
                let ret;
                if (isX(M)) {
                    ret = "";
                } else if (isX(m)) {
                    ret = `>=${M}.0.0 <${+M + 1}.0.0-0`;
                } else if (isX(p)) {
                    ret = `>=${M}.${m}.0 <${M}.${+m + 1}.0-0`;
                } else if (pr) {
                    debug("replaceTilde pr", pr);
                    ret = `>=${M}.${m}.${p}-${pr} <${M}.${+m + 1}.0-0`;
                } else {
                    ret = `>=${M}.${m}.${p} <${M}.${+m + 1}.0-0`;
                }
                debug("tilde return", ret);
                return ret;
            }));
        };
        const replaceCarets = (comp, options) => comp.trim().split(/\s+/).map((c => replaceCaret(c, options))).join(" ");
        const replaceCaret = (comp, options) => {
            debug("caret", comp, options);
            const r = options.loose ? re[t.CARETLOOSE] : re[t.CARET];
            const z = options.includePrerelease ? "-0" : "";
            return comp.replace(r, ((_, M, m, p, pr) => {
                debug("caret", comp, _, M, m, p, pr);
                let ret;
                if (isX(M)) {
                    ret = "";
                } else if (isX(m)) {
                    ret = `>=${M}.0.0${z} <${+M + 1}.0.0-0`;
                } else if (isX(p)) {
                    if (M === "0") {
                        ret = `>=${M}.${m}.0${z} <${M}.${+m + 1}.0-0`;
                    } else {
                        ret = `>=${M}.${m}.0${z} <${+M + 1}.0.0-0`;
                    }
                } else if (pr) {
                    debug("replaceCaret pr", pr);
                    if (M === "0") {
                        if (m === "0") {
                            ret = `>=${M}.${m}.${p}-${pr} <${M}.${m}.${+p + 1}-0`;
                        } else {
                            ret = `>=${M}.${m}.${p}-${pr} <${M}.${+m + 1}.0-0`;
                        }
                    } else {
                        ret = `>=${M}.${m}.${p}-${pr} <${+M + 1}.0.0-0`;
                    }
                } else {
                    debug("no pr");
                    if (M === "0") {
                        if (m === "0") {
                            ret = `>=${M}.${m}.${p}${z} <${M}.${m}.${+p + 1}-0`;
                        } else {
                            ret = `>=${M}.${m}.${p}${z} <${M}.${+m + 1}.0-0`;
                        }
                    } else {
                        ret = `>=${M}.${m}.${p} <${+M + 1}.0.0-0`;
                    }
                }
                debug("caret return", ret);
                return ret;
            }));
        };
        const replaceXRanges = (comp, options) => {
            debug("replaceXRanges", comp, options);
            return comp.split(/\s+/).map((c => replaceXRange(c, options))).join(" ");
        };
        const replaceXRange = (comp, options) => {
            comp = comp.trim();
            const r = options.loose ? re[t.XRANGELOOSE] : re[t.XRANGE];
            return comp.replace(r, ((ret, gtlt, M, m, p, pr) => {
                debug("xRange", comp, ret, gtlt, M, m, p, pr);
                const xM = isX(M);
                const xm = xM || isX(m);
                const xp = xm || isX(p);
                const anyX = xp;
                if (gtlt === "=" && anyX) {
                    gtlt = "";
                }
                pr = options.includePrerelease ? "-0" : "";
                if (xM) {
                    if (gtlt === ">" || gtlt === "<") {
                        ret = "<0.0.0-0";
                    } else {
                        ret = "*";
                    }
                } else if (gtlt && anyX) {
                    if (xm) {
                        m = 0;
                    }
                    p = 0;
                    if (gtlt === ">") {
                        gtlt = ">=";
                        if (xm) {
                            M = +M + 1;
                            m = 0;
                            p = 0;
                        } else {
                            m = +m + 1;
                            p = 0;
                        }
                    } else if (gtlt === "<=") {
                        gtlt = "<";
                        if (xm) {
                            M = +M + 1;
                        } else {
                            m = +m + 1;
                        }
                    }
                    if (gtlt === "<") {
                        pr = "-0";
                    }
                    ret = `${gtlt + M}.${m}.${p}${pr}`;
                } else if (xm) {
                    ret = `>=${M}.0.0${pr} <${+M + 1}.0.0-0`;
                } else if (xp) {
                    ret = `>=${M}.${m}.0${pr} <${M}.${+m + 1}.0-0`;
                }
                debug("xRange return", ret);
                return ret;
            }));
        };
        const replaceStars = (comp, options) => {
            debug("replaceStars", comp, options);
            return comp.trim().replace(re[t.STAR], "");
        };
        const replaceGTE0 = (comp, options) => {
            debug("replaceGTE0", comp, options);
            return comp.trim().replace(re[options.includePrerelease ? t.GTE0PRE : t.GTE0], "");
        };
        const hyphenReplace = incPr => ($0, from, fM, fm, fp, fpr, fb, to, tM, tm, tp, tpr, tb) => {
            if (isX(fM)) {
                from = "";
            } else if (isX(fm)) {
                from = `>=${fM}.0.0${incPr ? "-0" : ""}`;
            } else if (isX(fp)) {
                from = `>=${fM}.${fm}.0${incPr ? "-0" : ""}`;
            } else if (fpr) {
                from = `>=${from}`;
            } else {
                from = `>=${from}${incPr ? "-0" : ""}`;
            }
            if (isX(tM)) {
                to = "";
            } else if (isX(tm)) {
                to = `<${+tM + 1}.0.0-0`;
            } else if (isX(tp)) {
                to = `<${tM}.${+tm + 1}.0-0`;
            } else if (tpr) {
                to = `<=${tM}.${tm}.${tp}-${tpr}`;
            } else if (incPr) {
                to = `<${tM}.${tm}.${+tp + 1}-0`;
            } else {
                to = `<=${to}`;
            }
            return `${from} ${to}`.trim();
        };
        const testSet = (set, version, options) => {
            for (let i = 0; i < set.length; i++) {
                if (!set[i].test(version)) {
                    return false;
                }
            }
            if (version.prerelease.length && !options.includePrerelease) {
                for (let i = 0; i < set.length; i++) {
                    debug(set[i].semver);
                    if (set[i].semver === Comparator.ANY) {
                        continue;
                    }
                    if (set[i].semver.prerelease.length > 0) {
                        const allowed = set[i].semver;
                        if (allowed.major === version.major && allowed.minor === version.minor && allowed.patch === version.patch) {
                            return true;
                        }
                    }
                }
                return false;
            }
            return true;
        };
    },
    3013: (module, __unused_webpack_exports, __webpack_require__) => {
        const debug = __webpack_require__(5432);
        const {MAX_LENGTH, MAX_SAFE_INTEGER} = __webpack_require__(9041);
        const {safeRe: re, t} = __webpack_require__(9541);
        const parseOptions = __webpack_require__(3867);
        const {compareIdentifiers} = __webpack_require__(3650);
        class SemVer {
            constructor(version, options) {
                options = parseOptions(options);
                if (version instanceof SemVer) {
                    if (version.loose === !!options.loose && version.includePrerelease === !!options.includePrerelease) {
                        return version;
                    } else {
                        version = version.version;
                    }
                } else if (typeof version !== "string") {
                    throw new TypeError(`Invalid version. Must be a string. Got type "${typeof version}".`);
                }
                if (version.length > MAX_LENGTH) {
                    throw new TypeError(`version is longer than ${MAX_LENGTH} characters`);
                }
                debug("SemVer", version, options);
                this.options = options;
                this.loose = !!options.loose;
                this.includePrerelease = !!options.includePrerelease;
                const m = version.trim().match(options.loose ? re[t.LOOSE] : re[t.FULL]);
                if (!m) {
                    throw new TypeError(`Invalid Version: ${version}`);
                }
                this.raw = version;
                this.major = +m[1];
                this.minor = +m[2];
                this.patch = +m[3];
                if (this.major > MAX_SAFE_INTEGER || this.major < 0) {
                    throw new TypeError("Invalid major version");
                }
                if (this.minor > MAX_SAFE_INTEGER || this.minor < 0) {
                    throw new TypeError("Invalid minor version");
                }
                if (this.patch > MAX_SAFE_INTEGER || this.patch < 0) {
                    throw new TypeError("Invalid patch version");
                }
                if (!m[4]) {
                    this.prerelease = [];
                } else {
                    this.prerelease = m[4].split(".").map((id => {
                        if (/^[0-9]+$/.test(id)) {
                            const num = +id;
                            if (num >= 0 && num < MAX_SAFE_INTEGER) {
                                return num;
                            }
                        }
                        return id;
                    }));
                }
                this.build = m[5] ? m[5].split(".") : [];
                this.format();
            }
            format() {
                this.version = `${this.major}.${this.minor}.${this.patch}`;
                if (this.prerelease.length) {
                    this.version += `-${this.prerelease.join(".")}`;
                }
                return this.version;
            }
            toString() {
                return this.version;
            }
            compare(other) {
                debug("SemVer.compare", this.version, this.options, other);
                if (!(other instanceof SemVer)) {
                    if (typeof other === "string" && other === this.version) {
                        return 0;
                    }
                    other = new SemVer(other, this.options);
                }
                if (other.version === this.version) {
                    return 0;
                }
                return this.compareMain(other) || this.comparePre(other);
            }
            compareMain(other) {
                if (!(other instanceof SemVer)) {
                    other = new SemVer(other, this.options);
                }
                return compareIdentifiers(this.major, other.major) || compareIdentifiers(this.minor, other.minor) || compareIdentifiers(this.patch, other.patch);
            }
            comparePre(other) {
                if (!(other instanceof SemVer)) {
                    other = new SemVer(other, this.options);
                }
                if (this.prerelease.length && !other.prerelease.length) {
                    return -1;
                } else if (!this.prerelease.length && other.prerelease.length) {
                    return 1;
                } else if (!this.prerelease.length && !other.prerelease.length) {
                    return 0;
                }
                let i = 0;
                do {
                    const a = this.prerelease[i];
                    const b = other.prerelease[i];
                    debug("prerelease compare", i, a, b);
                    if (a === undefined && b === undefined) {
                        return 0;
                    } else if (b === undefined) {
                        return 1;
                    } else if (a === undefined) {
                        return -1;
                    } else if (a === b) {
                        continue;
                    } else {
                        return compareIdentifiers(a, b);
                    }
                } while (++i);
            }
            compareBuild(other) {
                if (!(other instanceof SemVer)) {
                    other = new SemVer(other, this.options);
                }
                let i = 0;
                do {
                    const a = this.build[i];
                    const b = other.build[i];
                    debug("prerelease compare", i, a, b);
                    if (a === undefined && b === undefined) {
                        return 0;
                    } else if (b === undefined) {
                        return 1;
                    } else if (a === undefined) {
                        return -1;
                    } else if (a === b) {
                        continue;
                    } else {
                        return compareIdentifiers(a, b);
                    }
                } while (++i);
            }
            inc(release, identifier, identifierBase) {
                switch (release) {
                  case "premajor":
                    this.prerelease.length = 0;
                    this.patch = 0;
                    this.minor = 0;
                    this.major++;
                    this.inc("pre", identifier, identifierBase);
                    break;

                  case "preminor":
                    this.prerelease.length = 0;
                    this.patch = 0;
                    this.minor++;
                    this.inc("pre", identifier, identifierBase);
                    break;

                  case "prepatch":
                    this.prerelease.length = 0;
                    this.inc("patch", identifier, identifierBase);
                    this.inc("pre", identifier, identifierBase);
                    break;

                  case "prerelease":
                    if (this.prerelease.length === 0) {
                        this.inc("patch", identifier, identifierBase);
                    }
                    this.inc("pre", identifier, identifierBase);
                    break;

                  case "major":
                    if (this.minor !== 0 || this.patch !== 0 || this.prerelease.length === 0) {
                        this.major++;
                    }
                    this.minor = 0;
                    this.patch = 0;
                    this.prerelease = [];
                    break;

                  case "minor":
                    if (this.patch !== 0 || this.prerelease.length === 0) {
                        this.minor++;
                    }
                    this.patch = 0;
                    this.prerelease = [];
                    break;

                  case "patch":
                    if (this.prerelease.length === 0) {
                        this.patch++;
                    }
                    this.prerelease = [];
                    break;

                  case "pre":
                    {
                        const base = Number(identifierBase) ? 1 : 0;
                        if (!identifier && identifierBase === false) {
                            throw new Error("invalid increment argument: identifier is empty");
                        }
                        if (this.prerelease.length === 0) {
                            this.prerelease = [ base ];
                        } else {
                            let i = this.prerelease.length;
                            while (--i >= 0) {
                                if (typeof this.prerelease[i] === "number") {
                                    this.prerelease[i]++;
                                    i = -2;
                                }
                            }
                            if (i === -1) {
                                if (identifier === this.prerelease.join(".") && identifierBase === false) {
                                    throw new Error("invalid increment argument: identifier already exists");
                                }
                                this.prerelease.push(base);
                            }
                        }
                        if (identifier) {
                            let prerelease = [ identifier, base ];
                            if (identifierBase === false) {
                                prerelease = [ identifier ];
                            }
                            if (compareIdentifiers(this.prerelease[0], identifier) === 0) {
                                if (isNaN(this.prerelease[1])) {
                                    this.prerelease = prerelease;
                                }
                            } else {
                                this.prerelease = prerelease;
                            }
                        }
                        break;
                    }

                  default:
                    throw new Error(`invalid increment argument: ${release}`);
                }
                this.raw = this.format();
                if (this.build.length) {
                    this.raw += `+${this.build.join(".")}`;
                }
                return this;
            }
        }
        module.exports = SemVer;
    },
    3470: (module, __unused_webpack_exports, __webpack_require__) => {
        const parse = __webpack_require__(7507);
        const clean = (version, options) => {
            const s = parse(version.trim().replace(/^[=v]+/, ""), options);
            return s ? s.version : null;
        };
        module.exports = clean;
    },
    1918: (module, __unused_webpack_exports, __webpack_require__) => {
        const eq = __webpack_require__(8443);
        const neq = __webpack_require__(1017);
        const gt = __webpack_require__(6077);
        const gte = __webpack_require__(4578);
        const lt = __webpack_require__(866);
        const lte = __webpack_require__(698);
        const cmp = (a, op, b, loose) => {
            switch (op) {
              case "===":
                if (typeof a === "object") {
                    a = a.version;
                }
                if (typeof b === "object") {
                    b = b.version;
                }
                return a === b;

              case "!==":
                if (typeof a === "object") {
                    a = a.version;
                }
                if (typeof b === "object") {
                    b = b.version;
                }
                return a !== b;

              case "":
              case "=":
              case "==":
                return eq(a, b, loose);

              case "!=":
                return neq(a, b, loose);

              case ">":
                return gt(a, b, loose);

              case ">=":
                return gte(a, b, loose);

              case "<":
                return lt(a, b, loose);

              case "<=":
                return lte(a, b, loose);

              default:
                throw new TypeError(`Invalid operator: ${op}`);
            }
        };
        module.exports = cmp;
    },
    4115: (module, __unused_webpack_exports, __webpack_require__) => {
        const SemVer = __webpack_require__(3013);
        const parse = __webpack_require__(7507);
        const {safeRe: re, t} = __webpack_require__(9541);
        const coerce = (version, options) => {
            if (version instanceof SemVer) {
                return version;
            }
            if (typeof version === "number") {
                version = String(version);
            }
            if (typeof version !== "string") {
                return null;
            }
            options = options || {};
            let match = null;
            if (!options.rtl) {
                match = version.match(re[t.COERCE]);
            } else {
                let next;
                while ((next = re[t.COERCERTL].exec(version)) && (!match || match.index + match[0].length !== version.length)) {
                    if (!match || next.index + next[0].length !== match.index + match[0].length) {
                        match = next;
                    }
                    re[t.COERCERTL].lastIndex = next.index + next[1].length + next[2].length;
                }
                re[t.COERCERTL].lastIndex = -1;
            }
            if (match === null) {
                return null;
            }
            return parse(`${match[2]}.${match[3] || "0"}.${match[4] || "0"}`, options);
        };
        module.exports = coerce;
    },
    6845: (module, __unused_webpack_exports, __webpack_require__) => {
        const SemVer = __webpack_require__(3013);
        const compareBuild = (a, b, loose) => {
            const versionA = new SemVer(a, loose);
            const versionB = new SemVer(b, loose);
            return versionA.compare(versionB) || versionA.compareBuild(versionB);
        };
        module.exports = compareBuild;
    },
    2310: (module, __unused_webpack_exports, __webpack_require__) => {
        const compare = __webpack_require__(2247);
        const compareLoose = (a, b) => compare(a, b, true);
        module.exports = compareLoose;
    },
    2247: (module, __unused_webpack_exports, __webpack_require__) => {
        const SemVer = __webpack_require__(3013);
        const compare = (a, b, loose) => new SemVer(a, loose).compare(new SemVer(b, loose));
        module.exports = compare;
    },
    5209: (module, __unused_webpack_exports, __webpack_require__) => {
        const parse = __webpack_require__(7507);
        const diff = (version1, version2) => {
            const v1 = parse(version1, null, true);
            const v2 = parse(version2, null, true);
            const comparison = v1.compare(v2);
            if (comparison === 0) {
                return null;
            }
            const v1Higher = comparison > 0;
            const highVersion = v1Higher ? v1 : v2;
            const lowVersion = v1Higher ? v2 : v1;
            const highHasPre = !!highVersion.prerelease.length;
            const lowHasPre = !!lowVersion.prerelease.length;
            if (lowHasPre && !highHasPre) {
                if (!lowVersion.patch && !lowVersion.minor) {
                    return "major";
                }
                if (highVersion.patch) {
                    return "patch";
                }
                if (highVersion.minor) {
                    return "minor";
                }
                return "major";
            }
            const prefix = highHasPre ? "pre" : "";
            if (v1.major !== v2.major) {
                return prefix + "major";
            }
            if (v1.minor !== v2.minor) {
                return prefix + "minor";
            }
            if (v1.patch !== v2.patch) {
                return prefix + "patch";
            }
            return "prerelease";
        };
        module.exports = diff;
    },
    8443: (module, __unused_webpack_exports, __webpack_require__) => {
        const compare = __webpack_require__(2247);
        const eq = (a, b, loose) => compare(a, b, loose) === 0;
        module.exports = eq;
    },
    6077: (module, __unused_webpack_exports, __webpack_require__) => {
        const compare = __webpack_require__(2247);
        const gt = (a, b, loose) => compare(a, b, loose) > 0;
        module.exports = gt;
    },
    4578: (module, __unused_webpack_exports, __webpack_require__) => {
        const compare = __webpack_require__(2247);
        const gte = (a, b, loose) => compare(a, b, loose) >= 0;
        module.exports = gte;
    },
    5210: (module, __unused_webpack_exports, __webpack_require__) => {
        const SemVer = __webpack_require__(3013);
        const inc = (version, release, options, identifier, identifierBase) => {
            if (typeof options === "string") {
                identifierBase = identifier;
                identifier = options;
                options = undefined;
            }
            try {
                return new SemVer(version instanceof SemVer ? version.version : version, options).inc(release, identifier, identifierBase).version;
            } catch (er) {
                return null;
            }
        };
        module.exports = inc;
    },
    866: (module, __unused_webpack_exports, __webpack_require__) => {
        const compare = __webpack_require__(2247);
        const lt = (a, b, loose) => compare(a, b, loose) < 0;
        module.exports = lt;
    },
    698: (module, __unused_webpack_exports, __webpack_require__) => {
        const compare = __webpack_require__(2247);
        const lte = (a, b, loose) => compare(a, b, loose) <= 0;
        module.exports = lte;
    },
    5847: (module, __unused_webpack_exports, __webpack_require__) => {
        const SemVer = __webpack_require__(3013);
        const major = (a, loose) => new SemVer(a, loose).major;
        module.exports = major;
    },
    1757: (module, __unused_webpack_exports, __webpack_require__) => {
        const SemVer = __webpack_require__(3013);
        const minor = (a, loose) => new SemVer(a, loose).minor;
        module.exports = minor;
    },
    1017: (module, __unused_webpack_exports, __webpack_require__) => {
        const compare = __webpack_require__(2247);
        const neq = (a, b, loose) => compare(a, b, loose) !== 0;
        module.exports = neq;
    },
    7507: (module, __unused_webpack_exports, __webpack_require__) => {
        const SemVer = __webpack_require__(3013);
        const parse = (version, options, throwErrors = false) => {
            if (version instanceof SemVer) {
                return version;
            }
            try {
                return new SemVer(version, options);
            } catch (er) {
                if (!throwErrors) {
                    return null;
                }
                throw er;
            }
        };
        module.exports = parse;
    },
    8150: (module, __unused_webpack_exports, __webpack_require__) => {
        const SemVer = __webpack_require__(3013);
        const patch = (a, loose) => new SemVer(a, loose).patch;
        module.exports = patch;
    },
    8011: (module, __unused_webpack_exports, __webpack_require__) => {
        const parse = __webpack_require__(7507);
        const prerelease = (version, options) => {
            const parsed = parse(version, options);
            return parsed && parsed.prerelease.length ? parsed.prerelease : null;
        };
        module.exports = prerelease;
    },
    9201: (module, __unused_webpack_exports, __webpack_require__) => {
        const compare = __webpack_require__(2247);
        const rcompare = (a, b, loose) => compare(b, a, loose);
        module.exports = rcompare;
    },
    7391: (module, __unused_webpack_exports, __webpack_require__) => {
        const compareBuild = __webpack_require__(6845);
        const rsort = (list, loose) => list.sort(((a, b) => compareBuild(b, a, loose)));
        module.exports = rsort;
    },
    8915: (module, __unused_webpack_exports, __webpack_require__) => {
        const Range = __webpack_require__(6833);
        const satisfies = (version, range, options) => {
            try {
                range = new Range(range, options);
            } catch (er) {
                return false;
            }
            return range.test(version);
        };
        module.exports = satisfies;
    },
    1934: (module, __unused_webpack_exports, __webpack_require__) => {
        const compareBuild = __webpack_require__(6845);
        const sort = (list, loose) => list.sort(((a, b) => compareBuild(a, b, loose)));
        module.exports = sort;
    },
    2555: (module, __unused_webpack_exports, __webpack_require__) => {
        const parse = __webpack_require__(7507);
        const valid = (version, options) => {
            const v = parse(version, options);
            return v ? v.version : null;
        };
        module.exports = valid;
    },
    6027: (module, __unused_webpack_exports, __webpack_require__) => {
        const internalRe = __webpack_require__(9541);
        const constants = __webpack_require__(9041);
        const SemVer = __webpack_require__(3013);
        const identifiers = __webpack_require__(3650);
        const parse = __webpack_require__(7507);
        const valid = __webpack_require__(2555);
        const clean = __webpack_require__(3470);
        const inc = __webpack_require__(5210);
        const diff = __webpack_require__(5209);
        const major = __webpack_require__(5847);
        const minor = __webpack_require__(1757);
        const patch = __webpack_require__(8150);
        const prerelease = __webpack_require__(8011);
        const compare = __webpack_require__(2247);
        const rcompare = __webpack_require__(9201);
        const compareLoose = __webpack_require__(2310);
        const compareBuild = __webpack_require__(6845);
        const sort = __webpack_require__(1934);
        const rsort = __webpack_require__(7391);
        const gt = __webpack_require__(6077);
        const lt = __webpack_require__(866);
        const eq = __webpack_require__(8443);
        const neq = __webpack_require__(1017);
        const gte = __webpack_require__(4578);
        const lte = __webpack_require__(698);
        const cmp = __webpack_require__(1918);
        const coerce = __webpack_require__(4115);
        const Comparator = __webpack_require__(7706);
        const Range = __webpack_require__(6833);
        const satisfies = __webpack_require__(8915);
        const toComparators = __webpack_require__(8378);
        const maxSatisfying = __webpack_require__(1678);
        const minSatisfying = __webpack_require__(1553);
        const minVersion = __webpack_require__(2262);
        const validRange = __webpack_require__(7396);
        const outside = __webpack_require__(939);
        const gtr = __webpack_require__(4933);
        const ltr = __webpack_require__(7233);
        const intersects = __webpack_require__(8842);
        const simplifyRange = __webpack_require__(3018);
        const subset = __webpack_require__(8563);
        module.exports = {
            parse,
            valid,
            clean,
            inc,
            diff,
            major,
            minor,
            patch,
            prerelease,
            compare,
            rcompare,
            compareLoose,
            compareBuild,
            sort,
            rsort,
            gt,
            lt,
            eq,
            neq,
            gte,
            lte,
            cmp,
            coerce,
            Comparator,
            Range,
            satisfies,
            toComparators,
            maxSatisfying,
            minSatisfying,
            minVersion,
            validRange,
            outside,
            gtr,
            ltr,
            intersects,
            simplifyRange,
            subset,
            SemVer,
            re: internalRe.re,
            src: internalRe.src,
            tokens: internalRe.t,
            SEMVER_SPEC_VERSION: constants.SEMVER_SPEC_VERSION,
            RELEASE_TYPES: constants.RELEASE_TYPES,
            compareIdentifiers: identifiers.compareIdentifiers,
            rcompareIdentifiers: identifiers.rcompareIdentifiers
        };
    },
    9041: module => {
        const SEMVER_SPEC_VERSION = "2.0.0";
        const MAX_LENGTH = 256;
        const MAX_SAFE_INTEGER = Number.MAX_SAFE_INTEGER || 9007199254740991;
        const MAX_SAFE_COMPONENT_LENGTH = 16;
        const MAX_SAFE_BUILD_LENGTH = MAX_LENGTH - 6;
        const RELEASE_TYPES = [ "major", "premajor", "minor", "preminor", "patch", "prepatch", "prerelease" ];
        module.exports = {
            MAX_LENGTH,
            MAX_SAFE_COMPONENT_LENGTH,
            MAX_SAFE_BUILD_LENGTH,
            MAX_SAFE_INTEGER,
            RELEASE_TYPES,
            SEMVER_SPEC_VERSION,
            FLAG_INCLUDE_PRERELEASE: 1,
            FLAG_LOOSE: 2
        };
    },
    5432: module => {
        const debug = typeof process === "object" && process.env && process.env.NODE_DEBUG && /\bsemver\b/i.test(process.env.NODE_DEBUG) ? (...args) => console.error("SEMVER", ...args) : () => {};
        module.exports = debug;
    },
    3650: module => {
        const numeric = /^[0-9]+$/;
        const compareIdentifiers = (a, b) => {
            const anum = numeric.test(a);
            const bnum = numeric.test(b);
            if (anum && bnum) {
                a = +a;
                b = +b;
            }
            return a === b ? 0 : anum && !bnum ? -1 : bnum && !anum ? 1 : a < b ? -1 : 1;
        };
        const rcompareIdentifiers = (a, b) => compareIdentifiers(b, a);
        module.exports = {
            compareIdentifiers,
            rcompareIdentifiers
        };
    },
    3867: module => {
        const looseOption = Object.freeze({
            loose: true
        });
        const emptyOpts = Object.freeze({});
        const parseOptions = options => {
            if (!options) {
                return emptyOpts;
            }
            if (typeof options !== "object") {
                return looseOption;
            }
            return options;
        };
        module.exports = parseOptions;
    },
    9541: (module, exports, __webpack_require__) => {
        const {MAX_SAFE_COMPONENT_LENGTH, MAX_SAFE_BUILD_LENGTH, MAX_LENGTH} = __webpack_require__(9041);
        const debug = __webpack_require__(5432);
        exports = module.exports = {};
        const re = exports.re = [];
        const safeRe = exports.safeRe = [];
        const src = exports.src = [];
        const t = exports.t = {};
        let R = 0;
        const LETTERDASHNUMBER = "[a-zA-Z0-9-]";
        const safeRegexReplacements = [ [ "\\s", 1 ], [ "\\d", MAX_LENGTH ], [ LETTERDASHNUMBER, MAX_SAFE_BUILD_LENGTH ] ];
        const makeSafeRegex = value => {
            for (const [token, max] of safeRegexReplacements) {
                value = value.split(`${token}*`).join(`${token}{0,${max}}`).split(`${token}+`).join(`${token}{1,${max}}`);
            }
            return value;
        };
        const createToken = (name, value, isGlobal) => {
            const safe = makeSafeRegex(value);
            const index = R++;
            debug(name, index, value);
            t[name] = index;
            src[index] = value;
            re[index] = new RegExp(value, isGlobal ? "g" : undefined);
            safeRe[index] = new RegExp(safe, isGlobal ? "g" : undefined);
        };
        createToken("NUMERICIDENTIFIER", "0|[1-9]\\d*");
        createToken("NUMERICIDENTIFIERLOOSE", "\\d+");
        createToken("NONNUMERICIDENTIFIER", `\\d*[a-zA-Z-]${LETTERDASHNUMBER}*`);
        createToken("MAINVERSION", `(${src[t.NUMERICIDENTIFIER]})\\.` + `(${src[t.NUMERICIDENTIFIER]})\\.` + `(${src[t.NUMERICIDENTIFIER]})`);
        createToken("MAINVERSIONLOOSE", `(${src[t.NUMERICIDENTIFIERLOOSE]})\\.` + `(${src[t.NUMERICIDENTIFIERLOOSE]})\\.` + `(${src[t.NUMERICIDENTIFIERLOOSE]})`);
        createToken("PRERELEASEIDENTIFIER", `(?:${src[t.NUMERICIDENTIFIER]}|${src[t.NONNUMERICIDENTIFIER]})`);
        createToken("PRERELEASEIDENTIFIERLOOSE", `(?:${src[t.NUMERICIDENTIFIERLOOSE]}|${src[t.NONNUMERICIDENTIFIER]})`);
        createToken("PRERELEASE", `(?:-(${src[t.PRERELEASEIDENTIFIER]}(?:\\.${src[t.PRERELEASEIDENTIFIER]})*))`);
        createToken("PRERELEASELOOSE", `(?:-?(${src[t.PRERELEASEIDENTIFIERLOOSE]}(?:\\.${src[t.PRERELEASEIDENTIFIERLOOSE]})*))`);
        createToken("BUILDIDENTIFIER", `${LETTERDASHNUMBER}+`);
        createToken("BUILD", `(?:\\+(${src[t.BUILDIDENTIFIER]}(?:\\.${src[t.BUILDIDENTIFIER]})*))`);
        createToken("FULLPLAIN", `v?${src[t.MAINVERSION]}${src[t.PRERELEASE]}?${src[t.BUILD]}?`);
        createToken("FULL", `^${src[t.FULLPLAIN]}$`);
        createToken("LOOSEPLAIN", `[v=\\s]*${src[t.MAINVERSIONLOOSE]}${src[t.PRERELEASELOOSE]}?${src[t.BUILD]}?`);
        createToken("LOOSE", `^${src[t.LOOSEPLAIN]}$`);
        createToken("GTLT", "((?:<|>)?=?)");
        createToken("XRANGEIDENTIFIERLOOSE", `${src[t.NUMERICIDENTIFIERLOOSE]}|x|X|\\*`);
        createToken("XRANGEIDENTIFIER", `${src[t.NUMERICIDENTIFIER]}|x|X|\\*`);
        createToken("XRANGEPLAIN", `[v=\\s]*(${src[t.XRANGEIDENTIFIER]})` + `(?:\\.(${src[t.XRANGEIDENTIFIER]})` + `(?:\\.(${src[t.XRANGEIDENTIFIER]})` + `(?:${src[t.PRERELEASE]})?${src[t.BUILD]}?` + `)?)?`);
        createToken("XRANGEPLAINLOOSE", `[v=\\s]*(${src[t.XRANGEIDENTIFIERLOOSE]})` + `(?:\\.(${src[t.XRANGEIDENTIFIERLOOSE]})` + `(?:\\.(${src[t.XRANGEIDENTIFIERLOOSE]})` + `(?:${src[t.PRERELEASELOOSE]})?${src[t.BUILD]}?` + `)?)?`);
        createToken("XRANGE", `^${src[t.GTLT]}\\s*${src[t.XRANGEPLAIN]}$`);
        createToken("XRANGELOOSE", `^${src[t.GTLT]}\\s*${src[t.XRANGEPLAINLOOSE]}$`);
        createToken("COERCE", `${"(^|[^\\d])" + "(\\d{1,"}${MAX_SAFE_COMPONENT_LENGTH}})` + `(?:\\.(\\d{1,${MAX_SAFE_COMPONENT_LENGTH}}))?` + `(?:\\.(\\d{1,${MAX_SAFE_COMPONENT_LENGTH}}))?` + `(?:$|[^\\d])`);
        createToken("COERCERTL", src[t.COERCE], true);
        createToken("LONETILDE", "(?:~>?)");
        createToken("TILDETRIM", `(\\s*)${src[t.LONETILDE]}\\s+`, true);
        exports.tildeTrimReplace = "$1~";
        createToken("TILDE", `^${src[t.LONETILDE]}${src[t.XRANGEPLAIN]}$`);
        createToken("TILDELOOSE", `^${src[t.LONETILDE]}${src[t.XRANGEPLAINLOOSE]}$`);
        createToken("LONECARET", "(?:\\^)");
        createToken("CARETTRIM", `(\\s*)${src[t.LONECARET]}\\s+`, true);
        exports.caretTrimReplace = "$1^";
        createToken("CARET", `^${src[t.LONECARET]}${src[t.XRANGEPLAIN]}$`);
        createToken("CARETLOOSE", `^${src[t.LONECARET]}${src[t.XRANGEPLAINLOOSE]}$`);
        createToken("COMPARATORLOOSE", `^${src[t.GTLT]}\\s*(${src[t.LOOSEPLAIN]})$|^$`);
        createToken("COMPARATOR", `^${src[t.GTLT]}\\s*(${src[t.FULLPLAIN]})$|^$`);
        createToken("COMPARATORTRIM", `(\\s*)${src[t.GTLT]}\\s*(${src[t.LOOSEPLAIN]}|${src[t.XRANGEPLAIN]})`, true);
        exports.comparatorTrimReplace = "$1$2$3";
        createToken("HYPHENRANGE", `^\\s*(${src[t.XRANGEPLAIN]})` + `\\s+-\\s+` + `(${src[t.XRANGEPLAIN]})` + `\\s*$`);
        createToken("HYPHENRANGELOOSE", `^\\s*(${src[t.XRANGEPLAINLOOSE]})` + `\\s+-\\s+` + `(${src[t.XRANGEPLAINLOOSE]})` + `\\s*$`);
        createToken("STAR", "(<|>)?=?\\s*\\*");
        createToken("GTE0", "^\\s*>=\\s*0\\.0\\.0\\s*$");
        createToken("GTE0PRE", "^\\s*>=\\s*0\\.0\\.0-0\\s*$");
    },
    6923: (module, __unused_webpack_exports, __webpack_require__) => {
        "use strict";
        const Yallist = __webpack_require__(1455);
        const MAX = Symbol("max");
        const LENGTH = Symbol("length");
        const LENGTH_CALCULATOR = Symbol("lengthCalculator");
        const ALLOW_STALE = Symbol("allowStale");
        const MAX_AGE = Symbol("maxAge");
        const DISPOSE = Symbol("dispose");
        const NO_DISPOSE_ON_SET = Symbol("noDisposeOnSet");
        const LRU_LIST = Symbol("lruList");
        const CACHE = Symbol("cache");
        const UPDATE_AGE_ON_GET = Symbol("updateAgeOnGet");
        const naiveLength = () => 1;
        class LRUCache {
            constructor(options) {
                if (typeof options === "number") options = {
                    max: options
                };
                if (!options) options = {};
                if (options.max && (typeof options.max !== "number" || options.max < 0)) throw new TypeError("max must be a non-negative number");
                const max = this[MAX] = options.max || Infinity;
                const lc = options.length || naiveLength;
                this[LENGTH_CALCULATOR] = typeof lc !== "function" ? naiveLength : lc;
                this[ALLOW_STALE] = options.stale || false;
                if (options.maxAge && typeof options.maxAge !== "number") throw new TypeError("maxAge must be a number");
                this[MAX_AGE] = options.maxAge || 0;
                this[DISPOSE] = options.dispose;
                this[NO_DISPOSE_ON_SET] = options.noDisposeOnSet || false;
                this[UPDATE_AGE_ON_GET] = options.updateAgeOnGet || false;
                this.reset();
            }
            set max(mL) {
                if (typeof mL !== "number" || mL < 0) throw new TypeError("max must be a non-negative number");
                this[MAX] = mL || Infinity;
                trim(this);
            }
            get max() {
                return this[MAX];
            }
            set allowStale(allowStale) {
                this[ALLOW_STALE] = !!allowStale;
            }
            get allowStale() {
                return this[ALLOW_STALE];
            }
            set maxAge(mA) {
                if (typeof mA !== "number") throw new TypeError("maxAge must be a non-negative number");
                this[MAX_AGE] = mA;
                trim(this);
            }
            get maxAge() {
                return this[MAX_AGE];
            }
            set lengthCalculator(lC) {
                if (typeof lC !== "function") lC = naiveLength;
                if (lC !== this[LENGTH_CALCULATOR]) {
                    this[LENGTH_CALCULATOR] = lC;
                    this[LENGTH] = 0;
                    this[LRU_LIST].forEach((hit => {
                        hit.length = this[LENGTH_CALCULATOR](hit.value, hit.key);
                        this[LENGTH] += hit.length;
                    }));
                }
                trim(this);
            }
            get lengthCalculator() {
                return this[LENGTH_CALCULATOR];
            }
            get length() {
                return this[LENGTH];
            }
            get itemCount() {
                return this[LRU_LIST].length;
            }
            rforEach(fn, thisp) {
                thisp = thisp || this;
                for (let walker = this[LRU_LIST].tail; walker !== null; ) {
                    const prev = walker.prev;
                    forEachStep(this, fn, walker, thisp);
                    walker = prev;
                }
            }
            forEach(fn, thisp) {
                thisp = thisp || this;
                for (let walker = this[LRU_LIST].head; walker !== null; ) {
                    const next = walker.next;
                    forEachStep(this, fn, walker, thisp);
                    walker = next;
                }
            }
            keys() {
                return this[LRU_LIST].toArray().map((k => k.key));
            }
            values() {
                return this[LRU_LIST].toArray().map((k => k.value));
            }
            reset() {
                if (this[DISPOSE] && this[LRU_LIST] && this[LRU_LIST].length) {
                    this[LRU_LIST].forEach((hit => this[DISPOSE](hit.key, hit.value)));
                }
                this[CACHE] = new Map;
                this[LRU_LIST] = new Yallist;
                this[LENGTH] = 0;
            }
            dump() {
                return this[LRU_LIST].map((hit => isStale(this, hit) ? false : {
                    k: hit.key,
                    v: hit.value,
                    e: hit.now + (hit.maxAge || 0)
                })).toArray().filter((h => h));
            }
            dumpLru() {
                return this[LRU_LIST];
            }
            set(key, value, maxAge) {
                maxAge = maxAge || this[MAX_AGE];
                if (maxAge && typeof maxAge !== "number") throw new TypeError("maxAge must be a number");
                const now = maxAge ? Date.now() : 0;
                const len = this[LENGTH_CALCULATOR](value, key);
                if (this[CACHE].has(key)) {
                    if (len > this[MAX]) {
                        del(this, this[CACHE].get(key));
                        return false;
                    }
                    const node = this[CACHE].get(key);
                    const item = node.value;
                    if (this[DISPOSE]) {
                        if (!this[NO_DISPOSE_ON_SET]) this[DISPOSE](key, item.value);
                    }
                    item.now = now;
                    item.maxAge = maxAge;
                    item.value = value;
                    this[LENGTH] += len - item.length;
                    item.length = len;
                    this.get(key);
                    trim(this);
                    return true;
                }
                const hit = new Entry(key, value, len, now, maxAge);
                if (hit.length > this[MAX]) {
                    if (this[DISPOSE]) this[DISPOSE](key, value);
                    return false;
                }
                this[LENGTH] += hit.length;
                this[LRU_LIST].unshift(hit);
                this[CACHE].set(key, this[LRU_LIST].head);
                trim(this);
                return true;
            }
            has(key) {
                if (!this[CACHE].has(key)) return false;
                const hit = this[CACHE].get(key).value;
                return !isStale(this, hit);
            }
            get(key) {
                return get(this, key, true);
            }
            peek(key) {
                return get(this, key, false);
            }
            pop() {
                const node = this[LRU_LIST].tail;
                if (!node) return null;
                del(this, node);
                return node.value;
            }
            del(key) {
                del(this, this[CACHE].get(key));
            }
            load(arr) {
                this.reset();
                const now = Date.now();
                for (let l = arr.length - 1; l >= 0; l--) {
                    const hit = arr[l];
                    const expiresAt = hit.e || 0;
                    if (expiresAt === 0) this.set(hit.k, hit.v); else {
                        const maxAge = expiresAt - now;
                        if (maxAge > 0) {
                            this.set(hit.k, hit.v, maxAge);
                        }
                    }
                }
            }
            prune() {
                this[CACHE].forEach(((value, key) => get(this, key, false)));
            }
        }
        const get = (self, key, doUse) => {
            const node = self[CACHE].get(key);
            if (node) {
                const hit = node.value;
                if (isStale(self, hit)) {
                    del(self, node);
                    if (!self[ALLOW_STALE]) return undefined;
                } else {
                    if (doUse) {
                        if (self[UPDATE_AGE_ON_GET]) node.value.now = Date.now();
                        self[LRU_LIST].unshiftNode(node);
                    }
                }
                return hit.value;
            }
        };
        const isStale = (self, hit) => {
            if (!hit || !hit.maxAge && !self[MAX_AGE]) return false;
            const diff = Date.now() - hit.now;
            return hit.maxAge ? diff > hit.maxAge : self[MAX_AGE] && diff > self[MAX_AGE];
        };
        const trim = self => {
            if (self[LENGTH] > self[MAX]) {
                for (let walker = self[LRU_LIST].tail; self[LENGTH] > self[MAX] && walker !== null; ) {
                    const prev = walker.prev;
                    del(self, walker);
                    walker = prev;
                }
            }
        };
        const del = (self, node) => {
            if (node) {
                const hit = node.value;
                if (self[DISPOSE]) self[DISPOSE](hit.key, hit.value);
                self[LENGTH] -= hit.length;
                self[CACHE].delete(hit.key);
                self[LRU_LIST].removeNode(node);
            }
        };
        class Entry {
            constructor(key, value, length, now, maxAge) {
                this.key = key;
                this.value = value;
                this.length = length;
                this.now = now;
                this.maxAge = maxAge || 0;
            }
        }
        const forEachStep = (self, fn, node, thisp) => {
            let hit = node.value;
            if (isStale(self, hit)) {
                del(self, node);
                if (!self[ALLOW_STALE]) hit = undefined;
            }
            if (hit) fn.call(thisp, hit.value, hit.key, self);
        };
        module.exports = LRUCache;
    },
    4933: (module, __unused_webpack_exports, __webpack_require__) => {
        const outside = __webpack_require__(939);
        const gtr = (version, range, options) => outside(version, range, ">", options);
        module.exports = gtr;
    },
    8842: (module, __unused_webpack_exports, __webpack_require__) => {
        const Range = __webpack_require__(6833);
        const intersects = (r1, r2, options) => {
            r1 = new Range(r1, options);
            r2 = new Range(r2, options);
            return r1.intersects(r2, options);
        };
        module.exports = intersects;
    },
    7233: (module, __unused_webpack_exports, __webpack_require__) => {
        const outside = __webpack_require__(939);
        const ltr = (version, range, options) => outside(version, range, "<", options);
        module.exports = ltr;
    },
    1678: (module, __unused_webpack_exports, __webpack_require__) => {
        const SemVer = __webpack_require__(3013);
        const Range = __webpack_require__(6833);
        const maxSatisfying = (versions, range, options) => {
            let max = null;
            let maxSV = null;
            let rangeObj = null;
            try {
                rangeObj = new Range(range, options);
            } catch (er) {
                return null;
            }
            versions.forEach((v => {
                if (rangeObj.test(v)) {
                    if (!max || maxSV.compare(v) === -1) {
                        max = v;
                        maxSV = new SemVer(max, options);
                    }
                }
            }));
            return max;
        };
        module.exports = maxSatisfying;
    },
    1553: (module, __unused_webpack_exports, __webpack_require__) => {
        const SemVer = __webpack_require__(3013);
        const Range = __webpack_require__(6833);
        const minSatisfying = (versions, range, options) => {
            let min = null;
            let minSV = null;
            let rangeObj = null;
            try {
                rangeObj = new Range(range, options);
            } catch (er) {
                return null;
            }
            versions.forEach((v => {
                if (rangeObj.test(v)) {
                    if (!min || minSV.compare(v) === 1) {
                        min = v;
                        minSV = new SemVer(min, options);
                    }
                }
            }));
            return min;
        };
        module.exports = minSatisfying;
    },
    2262: (module, __unused_webpack_exports, __webpack_require__) => {
        const SemVer = __webpack_require__(3013);
        const Range = __webpack_require__(6833);
        const gt = __webpack_require__(6077);
        const minVersion = (range, loose) => {
            range = new Range(range, loose);
            let minver = new SemVer("0.0.0");
            if (range.test(minver)) {
                return minver;
            }
            minver = new SemVer("0.0.0-0");
            if (range.test(minver)) {
                return minver;
            }
            minver = null;
            for (let i = 0; i < range.set.length; ++i) {
                const comparators = range.set[i];
                let setMin = null;
                comparators.forEach((comparator => {
                    const compver = new SemVer(comparator.semver.version);
                    switch (comparator.operator) {
                      case ">":
                        if (compver.prerelease.length === 0) {
                            compver.patch++;
                        } else {
                            compver.prerelease.push(0);
                        }
                        compver.raw = compver.format();

                      case "":
                      case ">=":
                        if (!setMin || gt(compver, setMin)) {
                            setMin = compver;
                        }
                        break;

                      case "<":
                      case "<=":
                        break;

                      default:
                        throw new Error(`Unexpected operation: ${comparator.operator}`);
                    }
                }));
                if (setMin && (!minver || gt(minver, setMin))) {
                    minver = setMin;
                }
            }
            if (minver && range.test(minver)) {
                return minver;
            }
            return null;
        };
        module.exports = minVersion;
    },
    939: (module, __unused_webpack_exports, __webpack_require__) => {
        const SemVer = __webpack_require__(3013);
        const Comparator = __webpack_require__(7706);
        const {ANY} = Comparator;
        const Range = __webpack_require__(6833);
        const satisfies = __webpack_require__(8915);
        const gt = __webpack_require__(6077);
        const lt = __webpack_require__(866);
        const lte = __webpack_require__(698);
        const gte = __webpack_require__(4578);
        const outside = (version, range, hilo, options) => {
            version = new SemVer(version, options);
            range = new Range(range, options);
            let gtfn, ltefn, ltfn, comp, ecomp;
            switch (hilo) {
              case ">":
                gtfn = gt;
                ltefn = lte;
                ltfn = lt;
                comp = ">";
                ecomp = ">=";
                break;

              case "<":
                gtfn = lt;
                ltefn = gte;
                ltfn = gt;
                comp = "<";
                ecomp = "<=";
                break;

              default:
                throw new TypeError('Must provide a hilo val of "<" or ">"');
            }
            if (satisfies(version, range, options)) {
                return false;
            }
            for (let i = 0; i < range.set.length; ++i) {
                const comparators = range.set[i];
                let high = null;
                let low = null;
                comparators.forEach((comparator => {
                    if (comparator.semver === ANY) {
                        comparator = new Comparator(">=0.0.0");
                    }
                    high = high || comparator;
                    low = low || comparator;
                    if (gtfn(comparator.semver, high.semver, options)) {
                        high = comparator;
                    } else if (ltfn(comparator.semver, low.semver, options)) {
                        low = comparator;
                    }
                }));
                if (high.operator === comp || high.operator === ecomp) {
                    return false;
                }
                if ((!low.operator || low.operator === comp) && ltefn(version, low.semver)) {
                    return false;
                } else if (low.operator === ecomp && ltfn(version, low.semver)) {
                    return false;
                }
            }
            return true;
        };
        module.exports = outside;
    },
    3018: (module, __unused_webpack_exports, __webpack_require__) => {
        const satisfies = __webpack_require__(8915);
        const compare = __webpack_require__(2247);
        module.exports = (versions, range, options) => {
            const set = [];
            let first = null;
            let prev = null;
            const v = versions.sort(((a, b) => compare(a, b, options)));
            for (const version of v) {
                const included = satisfies(version, range, options);
                if (included) {
                    prev = version;
                    if (!first) {
                        first = version;
                    }
                } else {
                    if (prev) {
                        set.push([ first, prev ]);
                    }
                    prev = null;
                    first = null;
                }
            }
            if (first) {
                set.push([ first, null ]);
            }
            const ranges = [];
            for (const [min, max] of set) {
                if (min === max) {
                    ranges.push(min);
                } else if (!max && min === v[0]) {
                    ranges.push("*");
                } else if (!max) {
                    ranges.push(`>=${min}`);
                } else if (min === v[0]) {
                    ranges.push(`<=${max}`);
                } else {
                    ranges.push(`${min} - ${max}`);
                }
            }
            const simplified = ranges.join(" || ");
            const original = typeof range.raw === "string" ? range.raw : String(range);
            return simplified.length < original.length ? simplified : range;
        };
    },
    8563: (module, __unused_webpack_exports, __webpack_require__) => {
        const Range = __webpack_require__(6833);
        const Comparator = __webpack_require__(7706);
        const {ANY} = Comparator;
        const satisfies = __webpack_require__(8915);
        const compare = __webpack_require__(2247);
        const subset = (sub, dom, options = {}) => {
            if (sub === dom) {
                return true;
            }
            sub = new Range(sub, options);
            dom = new Range(dom, options);
            let sawNonNull = false;
            OUTER: for (const simpleSub of sub.set) {
                for (const simpleDom of dom.set) {
                    const isSub = simpleSubset(simpleSub, simpleDom, options);
                    sawNonNull = sawNonNull || isSub !== null;
                    if (isSub) {
                        continue OUTER;
                    }
                }
                if (sawNonNull) {
                    return false;
                }
            }
            return true;
        };
        const minimumVersionWithPreRelease = [ new Comparator(">=0.0.0-0") ];
        const minimumVersion = [ new Comparator(">=0.0.0") ];
        const simpleSubset = (sub, dom, options) => {
            if (sub === dom) {
                return true;
            }
            if (sub.length === 1 && sub[0].semver === ANY) {
                if (dom.length === 1 && dom[0].semver === ANY) {
                    return true;
                } else if (options.includePrerelease) {
                    sub = minimumVersionWithPreRelease;
                } else {
                    sub = minimumVersion;
                }
            }
            if (dom.length === 1 && dom[0].semver === ANY) {
                if (options.includePrerelease) {
                    return true;
                } else {
                    dom = minimumVersion;
                }
            }
            const eqSet = new Set;
            let gt, lt;
            for (const c of sub) {
                if (c.operator === ">" || c.operator === ">=") {
                    gt = higherGT(gt, c, options);
                } else if (c.operator === "<" || c.operator === "<=") {
                    lt = lowerLT(lt, c, options);
                } else {
                    eqSet.add(c.semver);
                }
            }
            if (eqSet.size > 1) {
                return null;
            }
            let gtltComp;
            if (gt && lt) {
                gtltComp = compare(gt.semver, lt.semver, options);
                if (gtltComp > 0) {
                    return null;
                } else if (gtltComp === 0 && (gt.operator !== ">=" || lt.operator !== "<=")) {
                    return null;
                }
            }
            for (const eq of eqSet) {
                if (gt && !satisfies(eq, String(gt), options)) {
                    return null;
                }
                if (lt && !satisfies(eq, String(lt), options)) {
                    return null;
                }
                for (const c of dom) {
                    if (!satisfies(eq, String(c), options)) {
                        return false;
                    }
                }
                return true;
            }
            let higher, lower;
            let hasDomLT, hasDomGT;
            let needDomLTPre = lt && !options.includePrerelease && lt.semver.prerelease.length ? lt.semver : false;
            let needDomGTPre = gt && !options.includePrerelease && gt.semver.prerelease.length ? gt.semver : false;
            if (needDomLTPre && needDomLTPre.prerelease.length === 1 && lt.operator === "<" && needDomLTPre.prerelease[0] === 0) {
                needDomLTPre = false;
            }
            for (const c of dom) {
                hasDomGT = hasDomGT || c.operator === ">" || c.operator === ">=";
                hasDomLT = hasDomLT || c.operator === "<" || c.operator === "<=";
                if (gt) {
                    if (needDomGTPre) {
                        if (c.semver.prerelease && c.semver.prerelease.length && c.semver.major === needDomGTPre.major && c.semver.minor === needDomGTPre.minor && c.semver.patch === needDomGTPre.patch) {
                            needDomGTPre = false;
                        }
                    }
                    if (c.operator === ">" || c.operator === ">=") {
                        higher = higherGT(gt, c, options);
                        if (higher === c && higher !== gt) {
                            return false;
                        }
                    } else if (gt.operator === ">=" && !satisfies(gt.semver, String(c), options)) {
                        return false;
                    }
                }
                if (lt) {
                    if (needDomLTPre) {
                        if (c.semver.prerelease && c.semver.prerelease.length && c.semver.major === needDomLTPre.major && c.semver.minor === needDomLTPre.minor && c.semver.patch === needDomLTPre.patch) {
                            needDomLTPre = false;
                        }
                    }
                    if (c.operator === "<" || c.operator === "<=") {
                        lower = lowerLT(lt, c, options);
                        if (lower === c && lower !== lt) {
                            return false;
                        }
                    } else if (lt.operator === "<=" && !satisfies(lt.semver, String(c), options)) {
                        return false;
                    }
                }
                if (!c.operator && (lt || gt) && gtltComp !== 0) {
                    return false;
                }
            }
            if (gt && hasDomLT && !lt && gtltComp !== 0) {
                return false;
            }
            if (lt && hasDomGT && !gt && gtltComp !== 0) {
                return false;
            }
            if (needDomGTPre || needDomLTPre) {
                return false;
            }
            return true;
        };
        const higherGT = (a, b, options) => {
            if (!a) {
                return b;
            }
            const comp = compare(a.semver, b.semver, options);
            return comp > 0 ? a : comp < 0 ? b : b.operator === ">" && a.operator === ">=" ? b : a;
        };
        const lowerLT = (a, b, options) => {
            if (!a) {
                return b;
            }
            const comp = compare(a.semver, b.semver, options);
            return comp < 0 ? a : comp > 0 ? b : b.operator === "<" && a.operator === "<=" ? b : a;
        };
        module.exports = subset;
    },
    8378: (module, __unused_webpack_exports, __webpack_require__) => {
        const Range = __webpack_require__(6833);
        const toComparators = (range, options) => new Range(range, options).set.map((comp => comp.map((c => c.value)).join(" ").trim().split(" ")));
        module.exports = toComparators;
    },
    7396: (module, __unused_webpack_exports, __webpack_require__) => {
        const Range = __webpack_require__(6833);
        const validRange = (range, options) => {
            try {
                return new Range(range, options).range || "*";
            } catch (er) {
                return null;
            }
        };
        module.exports = validRange;
    },
    9797: (module, __unused_webpack_exports, __webpack_require__) => {
        "use strict";
        const os = __webpack_require__(2037);
        const tty = __webpack_require__(6224);
        const hasFlag = __webpack_require__(4288);
        const {env} = process;
        let forceColor;
        if (hasFlag("no-color") || hasFlag("no-colors") || hasFlag("color=false") || hasFlag("color=never")) {
            forceColor = 0;
        } else if (hasFlag("color") || hasFlag("colors") || hasFlag("color=true") || hasFlag("color=always")) {
            forceColor = 1;
        }
        if ("FORCE_COLOR" in env) {
            if (env.FORCE_COLOR === "true") {
                forceColor = 1;
            } else if (env.FORCE_COLOR === "false") {
                forceColor = 0;
            } else {
                forceColor = env.FORCE_COLOR.length === 0 ? 1 : Math.min(parseInt(env.FORCE_COLOR, 10), 3);
            }
        }
        function translateLevel(level) {
            if (level === 0) {
                return false;
            }
            return {
                level,
                hasBasic: true,
                has256: level >= 2,
                has16m: level >= 3
            };
        }
        function supportsColor(haveStream, streamIsTTY) {
            if (forceColor === 0) {
                return 0;
            }
            if (hasFlag("color=16m") || hasFlag("color=full") || hasFlag("color=truecolor")) {
                return 3;
            }
            if (hasFlag("color=256")) {
                return 2;
            }
            if (haveStream && !streamIsTTY && forceColor === undefined) {
                return 0;
            }
            const min = forceColor || 0;
            if (env.TERM === "dumb") {
                return min;
            }
            if (process.platform === "win32") {
                const osRelease = os.release().split(".");
                if (Number(osRelease[0]) >= 10 && Number(osRelease[2]) >= 10586) {
                    return Number(osRelease[2]) >= 14931 ? 3 : 2;
                }
                return 1;
            }
            if ("CI" in env) {
                if ([ "TRAVIS", "CIRCLECI", "APPVEYOR", "GITLAB_CI", "GITHUB_ACTIONS", "BUILDKITE" ].some((sign => sign in env)) || env.CI_NAME === "codeship") {
                    return 1;
                }
                return min;
            }
            if ("TEAMCITY_VERSION" in env) {
                return /^(9\.(0*[1-9]\d*)\.|\d{2,}\.)/.test(env.TEAMCITY_VERSION) ? 1 : 0;
            }
            if (env.COLORTERM === "truecolor") {
                return 3;
            }
            if ("TERM_PROGRAM" in env) {
                const version = parseInt((env.TERM_PROGRAM_VERSION || "").split(".")[0], 10);
                switch (env.TERM_PROGRAM) {
                  case "iTerm.app":
                    return version >= 3 ? 3 : 2;

                  case "Apple_Terminal":
                    return 2;
                }
            }
            if (/-256(color)?$/i.test(env.TERM)) {
                return 2;
            }
            if (/^screen|^xterm|^vt100|^vt220|^rxvt|color|ansi|cygwin|linux/i.test(env.TERM)) {
                return 1;
            }
            if ("COLORTERM" in env) {
                return 1;
            }
            return min;
        }
        function getSupportLevel(stream) {
            const level = supportsColor(stream, stream && stream.isTTY);
            return translateLevel(level);
        }
        module.exports = {
            supportsColor: getSupportLevel,
            stdout: translateLevel(supportsColor(true, tty.isatty(1))),
            stderr: translateLevel(supportsColor(true, tty.isatty(2)))
        };
    },
    3278: module => {
        "use strict";
        module.exports = function(Yallist) {
            Yallist.prototype[Symbol.iterator] = function*() {
                for (let walker = this.head; walker; walker = walker.next) {
                    yield walker.value;
                }
            };
        };
    },
    1455: (module, __unused_webpack_exports, __webpack_require__) => {
        "use strict";
        module.exports = Yallist;
        Yallist.Node = Node;
        Yallist.create = Yallist;
        function Yallist(list) {
            var self = this;
            if (!(self instanceof Yallist)) {
                self = new Yallist;
            }
            self.tail = null;
            self.head = null;
            self.length = 0;
            if (list && typeof list.forEach === "function") {
                list.forEach((function(item) {
                    self.push(item);
                }));
            } else if (arguments.length > 0) {
                for (var i = 0, l = arguments.length; i < l; i++) {
                    self.push(arguments[i]);
                }
            }
            return self;
        }
        Yallist.prototype.removeNode = function(node) {
            if (node.list !== this) {
                throw new Error("removing node which does not belong to this list");
            }
            var next = node.next;
            var prev = node.prev;
            if (next) {
                next.prev = prev;
            }
            if (prev) {
                prev.next = next;
            }
            if (node === this.head) {
                this.head = next;
            }
            if (node === this.tail) {
                this.tail = prev;
            }
            node.list.length--;
            node.next = null;
            node.prev = null;
            node.list = null;
            return next;
        };
        Yallist.prototype.unshiftNode = function(node) {
            if (node === this.head) {
                return;
            }
            if (node.list) {
                node.list.removeNode(node);
            }
            var head = this.head;
            node.list = this;
            node.next = head;
            if (head) {
                head.prev = node;
            }
            this.head = node;
            if (!this.tail) {
                this.tail = node;
            }
            this.length++;
        };
        Yallist.prototype.pushNode = function(node) {
            if (node === this.tail) {
                return;
            }
            if (node.list) {
                node.list.removeNode(node);
            }
            var tail = this.tail;
            node.list = this;
            node.prev = tail;
            if (tail) {
                tail.next = node;
            }
            this.tail = node;
            if (!this.head) {
                this.head = node;
            }
            this.length++;
        };
        Yallist.prototype.push = function() {
            for (var i = 0, l = arguments.length; i < l; i++) {
                push(this, arguments[i]);
            }
            return this.length;
        };
        Yallist.prototype.unshift = function() {
            for (var i = 0, l = arguments.length; i < l; i++) {
                unshift(this, arguments[i]);
            }
            return this.length;
        };
        Yallist.prototype.pop = function() {
            if (!this.tail) {
                return undefined;
            }
            var res = this.tail.value;
            this.tail = this.tail.prev;
            if (this.tail) {
                this.tail.next = null;
            } else {
                this.head = null;
            }
            this.length--;
            return res;
        };
        Yallist.prototype.shift = function() {
            if (!this.head) {
                return undefined;
            }
            var res = this.head.value;
            this.head = this.head.next;
            if (this.head) {
                this.head.prev = null;
            } else {
                this.tail = null;
            }
            this.length--;
            return res;
        };
        Yallist.prototype.forEach = function(fn, thisp) {
            thisp = thisp || this;
            for (var walker = this.head, i = 0; walker !== null; i++) {
                fn.call(thisp, walker.value, i, this);
                walker = walker.next;
            }
        };
        Yallist.prototype.forEachReverse = function(fn, thisp) {
            thisp = thisp || this;
            for (var walker = this.tail, i = this.length - 1; walker !== null; i--) {
                fn.call(thisp, walker.value, i, this);
                walker = walker.prev;
            }
        };
        Yallist.prototype.get = function(n) {
            for (var i = 0, walker = this.head; walker !== null && i < n; i++) {
                walker = walker.next;
            }
            if (i === n && walker !== null) {
                return walker.value;
            }
        };
        Yallist.prototype.getReverse = function(n) {
            for (var i = 0, walker = this.tail; walker !== null && i < n; i++) {
                walker = walker.prev;
            }
            if (i === n && walker !== null) {
                return walker.value;
            }
        };
        Yallist.prototype.map = function(fn, thisp) {
            thisp = thisp || this;
            var res = new Yallist;
            for (var walker = this.head; walker !== null; ) {
                res.push(fn.call(thisp, walker.value, this));
                walker = walker.next;
            }
            return res;
        };
        Yallist.prototype.mapReverse = function(fn, thisp) {
            thisp = thisp || this;
            var res = new Yallist;
            for (var walker = this.tail; walker !== null; ) {
                res.push(fn.call(thisp, walker.value, this));
                walker = walker.prev;
            }
            return res;
        };
        Yallist.prototype.reduce = function(fn, initial) {
            var acc;
            var walker = this.head;
            if (arguments.length > 1) {
                acc = initial;
            } else if (this.head) {
                walker = this.head.next;
                acc = this.head.value;
            } else {
                throw new TypeError("Reduce of empty list with no initial value");
            }
            for (var i = 0; walker !== null; i++) {
                acc = fn(acc, walker.value, i);
                walker = walker.next;
            }
            return acc;
        };
        Yallist.prototype.reduceReverse = function(fn, initial) {
            var acc;
            var walker = this.tail;
            if (arguments.length > 1) {
                acc = initial;
            } else if (this.tail) {
                walker = this.tail.prev;
                acc = this.tail.value;
            } else {
                throw new TypeError("Reduce of empty list with no initial value");
            }
            for (var i = this.length - 1; walker !== null; i--) {
                acc = fn(acc, walker.value, i);
                walker = walker.prev;
            }
            return acc;
        };
        Yallist.prototype.toArray = function() {
            var arr = new Array(this.length);
            for (var i = 0, walker = this.head; walker !== null; i++) {
                arr[i] = walker.value;
                walker = walker.next;
            }
            return arr;
        };
        Yallist.prototype.toArrayReverse = function() {
            var arr = new Array(this.length);
            for (var i = 0, walker = this.tail; walker !== null; i++) {
                arr[i] = walker.value;
                walker = walker.prev;
            }
            return arr;
        };
        Yallist.prototype.slice = function(from, to) {
            to = to || this.length;
            if (to < 0) {
                to += this.length;
            }
            from = from || 0;
            if (from < 0) {
                from += this.length;
            }
            var ret = new Yallist;
            if (to < from || to < 0) {
                return ret;
            }
            if (from < 0) {
                from = 0;
            }
            if (to > this.length) {
                to = this.length;
            }
            for (var i = 0, walker = this.head; walker !== null && i < from; i++) {
                walker = walker.next;
            }
            for (;walker !== null && i < to; i++, walker = walker.next) {
                ret.push(walker.value);
            }
            return ret;
        };
        Yallist.prototype.sliceReverse = function(from, to) {
            to = to || this.length;
            if (to < 0) {
                to += this.length;
            }
            from = from || 0;
            if (from < 0) {
                from += this.length;
            }
            var ret = new Yallist;
            if (to < from || to < 0) {
                return ret;
            }
            if (from < 0) {
                from = 0;
            }
            if (to > this.length) {
                to = this.length;
            }
            for (var i = this.length, walker = this.tail; walker !== null && i > to; i--) {
                walker = walker.prev;
            }
            for (;walker !== null && i > from; i--, walker = walker.prev) {
                ret.push(walker.value);
            }
            return ret;
        };
        Yallist.prototype.splice = function(start, deleteCount, ...nodes) {
            if (start > this.length) {
                start = this.length - 1;
            }
            if (start < 0) {
                start = this.length + start;
            }
            for (var i = 0, walker = this.head; walker !== null && i < start; i++) {
                walker = walker.next;
            }
            var ret = [];
            for (var i = 0; walker && i < deleteCount; i++) {
                ret.push(walker.value);
                walker = this.removeNode(walker);
            }
            if (walker === null) {
                walker = this.tail;
            }
            if (walker !== this.head && walker !== this.tail) {
                walker = walker.prev;
            }
            for (var i = 0; i < nodes.length; i++) {
                walker = insert(this, walker, nodes[i]);
            }
            return ret;
        };
        Yallist.prototype.reverse = function() {
            var head = this.head;
            var tail = this.tail;
            for (var walker = head; walker !== null; walker = walker.prev) {
                var p = walker.prev;
                walker.prev = walker.next;
                walker.next = p;
            }
            this.head = tail;
            this.tail = head;
            return this;
        };
        function insert(self, node, value) {
            var inserted = node === self.head ? new Node(value, null, node, self) : new Node(value, node, node.next, self);
            if (inserted.next === null) {
                self.tail = inserted;
            }
            if (inserted.prev === null) {
                self.head = inserted;
            }
            self.length++;
            return inserted;
        }
        function push(self, item) {
            self.tail = new Node(item, self.tail, null, self);
            if (!self.head) {
                self.head = self.tail;
            }
            self.length++;
        }
        function unshift(self, item) {
            self.head = new Node(item, null, self.head, self);
            if (!self.tail) {
                self.tail = self.head;
            }
            self.length++;
        }
        function Node(value, prev, next, list) {
            if (!(this instanceof Node)) {
                return new Node(value, prev, next, list);
            }
            this.list = list;
            this.value = value;
            if (prev) {
                prev.next = this;
                this.prev = prev;
            } else {
                this.prev = null;
            }
            if (next) {
                next.prev = this;
                this.next = next;
            } else {
                this.next = null;
            }
        }
        try {
            __webpack_require__(3278)(Yallist);
        } catch (er) {}
    },
    6829: (__unused_webpack_module, exports, __webpack_require__) => {
        "use strict";
        Object.defineProperty(exports, "__esModule", {
            value: true
        });
        exports.NodeRelease = void 0;
        const process = __webpack_require__(7282);
        const semver_1 = __webpack_require__(6027);
        const ONE_DAY_IN_MILLISECONDS = 864e5;
        class NodeRelease {
            constructor(majorVersion, opts) {
                var _a, _b;
                this.majorVersion = majorVersion;
                this.endOfLifeDate = opts.endOfLife === true ? undefined : opts.endOfLife;
                this.untested = (_a = opts.untested) !== null && _a !== void 0 ? _a : false;
                this.supportedRange = new semver_1.Range((_b = opts.supportedRange) !== null && _b !== void 0 ? _b : `^${majorVersion}.0.0`);
                this.endOfLife = opts.endOfLife === true || opts.endOfLife.getTime() <= Date.now();
                this.deprecated = !this.endOfLife && opts.endOfLife !== true && opts.endOfLife.getTime() - NodeRelease.DEPRECATION_WINDOW_MS <= Date.now();
                this.supported = !this.untested && !this.endOfLife;
            }
            static forThisRuntime() {
                const semver = new semver_1.SemVer(process.version);
                const majorVersion = semver.major;
                for (const nodeRelease of this.ALL_RELEASES) {
                    if (nodeRelease.majorVersion === majorVersion) {
                        return {
                            nodeRelease,
                            knownBroken: !nodeRelease.supportedRange.test(semver)
                        };
                    }
                }
                return {
                    nodeRelease: undefined,
                    knownBroken: false
                };
            }
            toString() {
                const eolInfo = this.endOfLifeDate ? ` (Planned end-of-life: ${this.endOfLifeDate.toISOString().slice(0, 10)})` : "";
                return `${this.supportedRange.raw}${eolInfo}`;
            }
        }
        exports.NodeRelease = NodeRelease;
        NodeRelease.DEPRECATION_WINDOW_MS = 30 * ONE_DAY_IN_MILLISECONDS;
        NodeRelease.ALL_RELEASES = [ ...[ 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11 ].map((majorVersion => new NodeRelease(majorVersion, {
            endOfLife: true
        }))), new NodeRelease(12, {
            endOfLife: new Date("2022-04-30"),
            supportedRange: "^12.7.0"
        }), new NodeRelease(13, {
            endOfLife: new Date("2020-06-01"),
            untested: true
        }), new NodeRelease(14, {
            endOfLife: new Date("2023-04-30"),
            supportedRange: "^14.17.0"
        }), new NodeRelease(15, {
            endOfLife: new Date("2021-06-01"),
            untested: true
        }), new NodeRelease(16, {
            endOfLife: new Date("2023-09-11"),
            supportedRange: "^16.3.0"
        }), new NodeRelease(17, {
            endOfLife: new Date("2022-06-01"),
            supportedRange: "^17.3.0",
            untested: true
        }), new NodeRelease(19, {
            endOfLife: new Date("2023-06-01"),
            untested: true
        }), new NodeRelease(18, {
            endOfLife: new Date("2025-04-30")
        }), new NodeRelease(20, {
            endOfLife: new Date("2026-04-30")
        }), new NodeRelease(21, {
            endOfLife: new Date("2024-06-01"),
            untested: true
        }), new NodeRelease(22, {
            endOfLife: new Date("2027-04-30")
        }) ];
    },
    7962: (__unused_webpack_module, exports, __webpack_require__) => {
        "use strict";
        Object.defineProperty(exports, "__esModule", {
            value: true
        });
        exports.checkNode = exports.NodeRelease = void 0;
        const chalk_1 = __webpack_require__(1201);
        const console_1 = __webpack_require__(6206);
        const process_1 = __webpack_require__(7282);
        const constants_1 = __webpack_require__(6829);
        var constants_2 = __webpack_require__(6829);
        Object.defineProperty(exports, "NodeRelease", {
            enumerable: true,
            get: function() {
                return constants_2.NodeRelease;
            }
        });
        function checkNode(envPrefix = "JSII") {
            var _a;
            const {nodeRelease, knownBroken} = constants_1.NodeRelease.forThisRuntime();
            const defaultCallToAction = "Should you encounter odd runtime issues, please try using one of the supported release before filing a bug report.";
            if (nodeRelease === null || nodeRelease === void 0 ? void 0 : nodeRelease.endOfLife) {
                const silenceVariable = `${envPrefix}_SILENCE_WARNING_END_OF_LIFE_NODE_VERSION`;
                const silencedVersions = ((_a = process.env[silenceVariable]) !== null && _a !== void 0 ? _a : "").split(",").map((v => v.trim()));
                const qualifier = nodeRelease.endOfLifeDate ? ` on ${nodeRelease.endOfLifeDate.toISOString().slice(0, 10)}` : "";
                if (!silencedVersions.includes(nodeRelease.majorVersion.toString())) veryVisibleMessage(chalk_1.bgRed.white.bold, `Node ${nodeRelease.majorVersion} has reached end-of-life${qualifier} and is not supported.`, `Please upgrade to a supported node version as soon as possible.`);
            } else if (knownBroken) {
                const silenceVariable = `${envPrefix}_SILENCE_WARNING_KNOWN_BROKEN_NODE_VERSION`;
                if (!process.env[silenceVariable]) veryVisibleMessage(chalk_1.bgRed.white.bold, `Node ${process_1.version} is unsupported and has known compatibility issues with this software.`, defaultCallToAction, silenceVariable);
            } else if (!nodeRelease || nodeRelease.untested) {
                const silenceVariable = `${envPrefix}_SILENCE_WARNING_UNTESTED_NODE_VERSION`;
                if (!process.env[silenceVariable]) {
                    veryVisibleMessage(chalk_1.bgYellow.black, `This software has not been tested with node ${process_1.version}.`, defaultCallToAction, silenceVariable);
                }
            } else if (nodeRelease === null || nodeRelease === void 0 ? void 0 : nodeRelease.deprecated) {
                const silenceVariable = `${envPrefix}_SILENCE_WARNING_DEPRECATED_NODE_VERSION`;
                if (!process.env[silenceVariable]) {
                    const deadline = nodeRelease.endOfLifeDate.toISOString().slice(0, 10);
                    veryVisibleMessage(chalk_1.bgYellowBright.black, `Node ${nodeRelease.majorVersion} is approaching end-of-life and will no longer be supported in new releases after ${deadline}.`, `Please upgrade to a supported node version as soon as possible.`, silenceVariable);
                }
            }
            function veryVisibleMessage(chalk, message, callToAction, silenceVariable) {
                const lines = [ message, callToAction, "", `This software is currently running on node ${process_1.version}.`, "As of the current release of this software, supported node releases are:", ...constants_1.NodeRelease.ALL_RELEASES.filter((release => release.supported)).sort(((l, r) => {
                    var _a, _b, _c, _d;
                    return ((_b = (_a = r.endOfLifeDate) === null || _a === void 0 ? void 0 : _a.getTime()) !== null && _b !== void 0 ? _b : 0) - ((_d = (_c = l.endOfLifeDate) === null || _c === void 0 ? void 0 : _c.getTime()) !== null && _d !== void 0 ? _d : 0);
                })).map((release => `- ${release.toString()}${release.deprecated ? " [DEPRECATED]" : ""}`)), ...silenceVariable ? [ "", `This warning can be silenced by setting the ${silenceVariable} environment variable.` ] : [] ];
                const len = Math.max(...lines.map((l => l.length)));
                const border = chalk("!".repeat(len + 8));
                const spacer = chalk(`!!  ${" ".repeat(len)}  !!`);
                (0, console_1.error)(border);
                (0, console_1.error)(spacer);
                for (const line of lines) {
                    (0, console_1.error)(chalk(`!!  ${line.padEnd(len, " ")}  !!`));
                }
                (0, console_1.error)(spacer);
                (0, console_1.error)(border);
            }
        }
        exports.checkNode = checkNode;
    },
    9317: (module, __unused_webpack_exports, __webpack_require__) => {
        "use strict";
        const index_1 = __webpack_require__(7962);
        (0, index_1.checkNode)();
        module.exports = {};
    },
    2081: module => {
        "use strict";
        module.exports = require("child_process");
    },
    6206: module => {
        "use strict";
        module.exports = require("console");
    },
    2037: module => {
        "use strict";
        module.exports = require("os");
    },
    4822: module => {
        "use strict";
        module.exports = require("path");
    },
    7282: module => {
        "use strict";
        module.exports = require("process");
    },
    6224: module => {
        "use strict";
        module.exports = require("tty");
    }
};

var __webpack_module_cache__ = {};

function __webpack_require__(moduleId) {
    var cachedModule = __webpack_module_cache__[moduleId];
    if (cachedModule !== undefined) {
        return cachedModule.exports;
    }
    var module = __webpack_module_cache__[moduleId] = {
        id: moduleId,
        loaded: false,
        exports: {}
    };
    __webpack_modules__[moduleId](module, module.exports, __webpack_require__);
    module.loaded = true;
    return module.exports;
}

(() => {
    __webpack_require__.nmd = module => {
        module.paths = [];
        if (!module.children) module.children = [];
        return module;
    };
})();

var __webpack_exports__ = {};

(() => {
    "use strict";
    var exports = __webpack_exports__;
    var __webpack_unused_export__;
    __webpack_unused_export__ = {
        value: true
    };
    __webpack_require__(9317);
    const child_process_1 = __webpack_require__(2081);
    const console_1 = __webpack_require__(6206);
    const os_1 = __webpack_require__(2037);
    const path_1 = __webpack_require__(4822);
    const child = (0, child_process_1.spawn)(process.execPath, [ ...process.execArgv, "--preserve-symlinks", (0, 
    path_1.resolve)(__dirname, "..", "lib", "program.js") ], {
        stdio: [ "ignore", "pipe", "pipe", "pipe" ]
    });
    child.once("end", ((code, signal) => {
        var _a;
        if (signal != null) {
            process.exit(128 + ((_a = os_1.constants.signals[signal]) !== null && _a !== void 0 ? _a : 0));
        }
        process.exit(code);
    }));
    child.once("error", (err => {
        console.error("Failed to spawn child process:", err.stack);
        process.exit(-1);
    }));
    for (const signal of Object.keys(os_1.constants.signals)) {
        if (signal === "SIGKILL" || signal === "SIGSTOP") {
            continue;
        }
        process.on(signal, (sig => child.kill(sig)));
    }
    function makeHandler(tag) {
        return chunk => {
            const buffer = Buffer.from(chunk);
            (0, console_1.error)(JSON.stringify({
                [tag]: buffer.toString("base64")
            }));
        };
    }
    child.stdout.on("data", makeHandler("stdout"));
    child.stderr.on("data", makeHandler("stderr"));
    const commands = child.stdio[3];
    process.stdin.pipe(commands);
    commands.pipe(process.stdout);
})();
//# sourceMappingURL=jsii-runtime.js.map