function TRBL(name, imp) {
    this.name = TRBL.extractMain(name);
    this.sides = {
        'top': null,
        'right': null,
        'bottom': null,
        'left': null
    };
    this.imp = imp ? 4 : 0;
}

TRBL.props = {
    'margin': 1,
    'margin-top': 1,
    'margin-right': 1,
    'margin-bottom': 1,
    'margin-left': 1,
    'padding': 1,
    'padding-top': 1,
    'padding-right': 1,
    'padding-bottom': 1,
    'padding-left': 1
};

TRBL.extractMain = function(name) {
    var i = name.indexOf('-');
    return i === -1 ? name : name.substr(0, i);
};

TRBL.prototype.impSum = function() {
    var imp = 0, n = 0;
    for (var k in this.sides) {
        if (this.sides[k]) {
            n++;
            if (this.sides[k].imp) imp++;
        }
    }
    return imp === n ? imp : 0;
};

TRBL.prototype.add = function(name, sValue, tValue, imp) {
    var s = this.sides,
        currentSide,
        i, x, side, a = [], last,
        imp = imp ? 1 : 0,
        wasUnary = false;
    if ((i = name.lastIndexOf('-')) !== -1) {
        side = name.substr(i + 1);
        if (side in s) {
            if (!(currentSide = s[side]) || (imp && !currentSide.imp)) {
                s[side] = { s: imp ? sValue.substring(0, sValue.length - 10) : sValue, t: [tValue[0]], imp: imp };
                if (tValue[0][1] === 'unary') s[side].t.push(tValue[1]);
            }
            return true;
        }
    } else if (name === this.name) {
        for (i = 0; i < tValue.length; i++) {
            x = tValue[i];
            last = a[a.length - 1];
            switch(x[1]) {
                case 'unary':
                    a.push({ s: x[2], t: [x], imp: imp });
                    wasUnary = true;
                    break;
                case 'number':
                case 'ident':
                    if (wasUnary) {
                        last.t.push(x);
                        last.s += x[2];
                    } else {
                        a.push({ s: x[2], t: [x], imp: imp });
                    }
                    wasUnary = false;
                    break;
                case 'percentage':
                    if (wasUnary) {
                        last.t.push(x);
                        last.s += x[2][2] + '%';
                    } else {
                        a.push({ s: x[2][2] + '%', t: [x], imp: imp });
                    }
                    wasUnary = false;
                    break;
                case 'dimension':
                    if (wasUnary) {
                        last.t.push(x);
                        last.s += x[2][2] + x[3][2];
                    } else {
                        a.push({ s: x[2][2] + x[3][2], t: [x], imp: imp });
                    }
                    wasUnary = false;
                    break;
                case 's':
                case 'comment':
                case 'important':
                    break;
                default:
                    return false;
            }
        }

        if (a.length > 4) return false;

        if (!a[1]) a[1] = a[0];
        if (!a[2]) a[2] = a[0];
        if (!a[3]) a[3] = a[1];

        if (!s.top) s.top = a[0];
        if (!s.right) s.right = a[1];
        if (!s.bottom) s.bottom = a[2];
        if (!s.left) s.left = a[3];

        return true;
    }
};

TRBL.prototype.isOkToMinimize = function() {
    var s = this.sides,
        imp,
        ieReg = /\\9$/;

    if (!!(s.top && s.right && s.bottom && s.left)) {
        imp = s.top.imp + s.right.imp + s.bottom.imp + s.left.imp;

        if (ieReg.test(s.top.s) || ieReg.test(s.right.s) || ieReg.test(s.bottom.s) || ieReg.test(s.left.s)) {
            return false;
        }

        return (imp === 0 || imp === 4 || imp === this.imp);
    }
    return false;
};

TRBL.prototype.getValue = function() {
    var s = this.sides,
        a = [s.top, s.right, s.bottom, s.left],
        r = [{}, 'value'];

    if (s.left.s === s.right.s) {
        a.length--;
        if (s.bottom.s === s.top.s) {
            a.length--;
            if (s.right.s === s.top.s) {
                a.length--;
            }
        }
    }

    for (var i = 0; i < a.length - 1; i++) {
        r = r.concat(a[i].t);
        r.push([{ s: ' ' }, 's', ' ']);
    }
    r = r.concat(a[i].t);

    if (this.impSum()) r.push([{ s: '!important'}, 'important']);

    return r;
};

TRBL.prototype.getProperty = function() {
    return [{ s: this.name }, 'property', [{ s: this.name }, 'ident', this.name]];
};

TRBL.prototype.getString = function() {
    var p = this.getProperty(),
        v = this.getValue().slice(2),
        r = p[0].s + ':';

    for (var i = 0; i < v.length; i++) r += v[i][0].s;

    return r;
};
var NON_LENGTH_UNIT = ['deg', 'grad', 'rad', 'turn', 's', 'ms', 'Hz', 'kHz', 'dpi', 'dpcm', 'dppx'];

function CSSOCompressor() {}

CSSOCompressor.prototype.init = function() {
    this.props = {};
    this.shorts = {};
    this.shorts2 = {};

    this.ccrules = {}; // clean comment rules — special case to resolve ambiguity
    this.crules = {}; // compress rules
    this.prules = {}; // prepare rules
    this.frrules = {}; // freeze ruleset rules
    this.msrules = {}; // mark shorthands rules
    this.csrules = {}; // clean shorthands rules
    this.rbrules = {}; // restructure block rules
    this.rjrules = {}; // rejoin ruleset rules
    this.rrrules = {}; // restructure ruleset rules
    this.frules = {}; // finalize rules

    this.initRules(this.crules, this.defCCfg);
    this.initRules(this.ccrules, this.cleanCfg);
    this.initRules(this.frrules, this.frCfg);
    this.initRules(this.prules, this.preCfg);
    this.initRules(this.msrules, this.msCfg);
    this.initRules(this.csrules, this.csCfg);
    this.initRules(this.rbrules, this.defRBCfg);
    this.initRules(this.rjrules, this.defRJCfg);
    this.initRules(this.rrrules, this.defRRCfg);
    this.initRules(this.frules, this.defFCfg);

    this.shortGroupID = 0;
    this.lastShortGroupID = 0;
    this.lastShortSelector = 0;
};

CSSOCompressor.prototype.initRules = function(r, cfg) {
    var o = this.order,
        p = this.profile,
        x, i, k,
        t = [];

    for (i = 0; i < o.length; i++) if (o[i] in cfg) t.push(o[i]);

    if (!t.length) t = o;
    for (i = 0; i < t.length; i++) {
        x = p[t[i]];
        for (k in x) r[k] ? r[k].push(t[i]) : r[k] = [t[i]];
    }
};

CSSOCompressor.prototype.cleanCfg = {
    'cleanComment': 1
};

CSSOCompressor.prototype.defCCfg = {
    'cleanCharset': 1,
    'cleanImport': 1,
    'cleanWhitespace': 1,
    'cleanDecldelim': 1,
    'compressNumber': 1,
    'cleanUnary': 1,
    'compressColor': 1,
    'compressDimension': 1,
    'compressString': 1,
    'compressFontWeight': 1,
    'compressFont': 1,
    'compressBackground': 1,
    'cleanEmpty': 1
};

CSSOCompressor.prototype.defRBCfg = {
    'restructureBlock': 1
};

CSSOCompressor.prototype.defRJCfg = {
    'rejoinRuleset': 1,
    'cleanEmpty': 1
};

CSSOCompressor.prototype.defRRCfg = {
    'restructureRuleset': 1,
    'cleanEmpty': 1
};

CSSOCompressor.prototype.defFCfg = {
    'cleanEmpty': 1,
    'delimSelectors': 1,
    'delimBlocks': 1
};

CSSOCompressor.prototype.preCfg = {
    'destroyDelims': 1,
    'preTranslate': 1
};

CSSOCompressor.prototype.msCfg = {
    'markShorthands': 1
};

CSSOCompressor.prototype.frCfg = {
    'freezeRulesets': 1
};

CSSOCompressor.prototype.csCfg = {
    'cleanShorthands': 1,
    'cleanEmpty': 1
};

CSSOCompressor.prototype.order = [
    'cleanCharset',
    'cleanImport',
    'cleanComment',
    'cleanWhitespace',
    'compressNumber',
    'cleanUnary',
    'compressColor',
    'compressDimension',
    'compressString',
    'compressFontWeight',
    'compressFont',
    'compressBackground',
    'freezeRulesets',
    'destroyDelims',
    'preTranslate',
    'markShorthands',
    'cleanShorthands',
    'restructureBlock',
    'rejoinRuleset',
    'restructureRuleset',
    'cleanEmpty',
    'delimSelectors',
    'delimBlocks'
];

CSSOCompressor.prototype.profile = {
    'cleanCharset': {
        'atrules': 1
    },
    'cleanImport': {
        'atrules': 1
    },
    'cleanWhitespace': {
        's': 1
    },
    'compressNumber': {
        'number': 1
    },
    'cleanUnary': {
        'unary': 1
    },
    'compressColor': {
        'vhash': 1,
        'funktion': 1,
        'ident': 1
    },
    'compressDimension': {
        'dimension': 1
    },
    'compressString': {
        'string': 1
    },
    'compressFontWeight': {
        'declaration': 1
    },
    'compressFont': {
        'declaration': 1
    },
    'compressBackground': {
        'declaration': 1
    },
    'cleanComment': {
        'comment': 1
    },
    'cleanDecldelim': {
        'block': 1
    },
    'cleanEmpty': {
        'ruleset': 1,
        'atruleb': 1,
        'atruler': 1
    },
    'destroyDelims': {
        'decldelim': 1,
        'delim': 1
    },
    'preTranslate': {
        'declaration': 1,
        'property': 1,
        'simpleselector': 1,
        'filter': 1,
        'value': 1,
        'number': 1,
        'percentage': 1,
        'dimension': 1,
        'ident': 1
    },
    'restructureBlock': {
        'block': 1
    },
    'rejoinRuleset': {
        'ruleset': 1
    },
    'restructureRuleset': {
        'ruleset': 1
    },
    'delimSelectors': {
        'selector': 1
    },
    'delimBlocks': {
        'block': 1
    },
    'markShorthands': {
        'block': 1
    },
    'cleanShorthands': {
        'declaration': 1
    },
    'freezeRulesets': {
        'ruleset': 1
    }
};

CSSOCompressor.prototype.isContainer = function(o) {
    if (Array.isArray(o)) {
        for (var i = 0; i < o.length; i++) if (Array.isArray(o[i])) return true;
    }
};

CSSOCompressor.prototype.process = function(rules, token, container, i, path) {
    var rule = token[1];
    if (rule && rules[rule]) {
        var r = rules[rule],
            x1 = token, x2,
            o = this.order, k;
        for (var k = 0; k < r.length; k++) {
            x2 = this[r[k]](x1, rule, container, i, path);
            if (x2 === null) return null;
            else if (x2 !== undefined) x1 = x2;
        }
    }
    return x1;
};

CSSOCompressor.prototype.compress = function(tree, ro) {
    tree = tree || ['stylesheet'];
    this.init();
    this.info = true;

    var x = (typeof tree[0] !== 'string') ? tree : this.injectInfo([tree])[0],
        l0, l1 = 100000000000, ls,
        x0, x1, xs,
        protectedComment = this.findProtectedComment(tree);

    // compression without restructure
    x = this.walk(this.ccrules, x, '/0');
    x = this.walk(this.crules, x, '/0');
    x = this.walk(this.prules, x, '/0');
    x = this.walk(this.frrules, x, '/0');

    ls = translator.translate(cleanInfo(x)).length;

    if (!ro) { // restructure ON
        xs = this.copyArray(x);
        x = this.walk(this.rjrules, x, '/0');
        this.disjoin(x);
        x = this.walk(this.msrules, x, '/0');
        x = this.walk(this.csrules, x, '/0');
        x = this.walk(this.rbrules, x, '/0');
        do {
            l0 = l1;
            x0 = this.copyArray(x);
            x = this.walk(this.rjrules, x, '/0');
            x = this.walk(this.rrrules, x, '/0');
            l1 = translator.translate(cleanInfo(x)).length;
            x1 = this.copyArray(x);
        } while (l0 > l1);
        if (ls < l0 && ls < l1) x = xs;
        else if (l0 < l1) x = x0;
    }

    x = this.walk(this.frules, x, '/0');

    if (protectedComment) x.splice(2, 0, protectedComment);

    return x;
};

CSSOCompressor.prototype.findProtectedComment = function(tree) {
    var token;
    for (var i = 2; i < tree.length; i++) {
        token = tree[i];
        if (token[1] === 'comment' && token[2].length > 0 && token[2].charAt(0) === '!') return token;
        if (token[1] !== 's') return;
    }
};

CSSOCompressor.prototype.injectInfo = function(token) {
    var t;
    for (var i = token.length - 1; i > -1; i--) {
        t = token[i];
        if (t && Array.isArray(t)) {
            if (this.isContainer(t)) t = this.injectInfo(t);
            t.splice(0, 0, {});
        }
    }
    return token;
};

CSSOCompressor.prototype.disjoin = function(container) {
    var t, s, r, sr;

    for (var i = container.length - 1; i > -1; i--) {
        t = container[i];
        if (t && Array.isArray(t)) {
            if (t[1] === 'ruleset') {
                t[0].shortGroupID = this.shortGroupID++;
                s = t[2];
                if (s.length > 3) {
                    sr = s.slice(0, 2);
                    for (var k = s.length - 1; k > 1; k--) {
                        r = this.copyArray(t);
                        r[2] = sr.concat([s[k]]);
                        r[2][0].s = s[k][0].s;
                        container.splice(i + 1, 0, r);
                    }
                    container.splice(i, 1);
                }
            }
        }
        if (this.isContainer(t)) this.disjoin(t);
    }
};

CSSOCompressor.prototype.walk = function(rules, container, path) {
    var t, x;
    for (var i = container.length - 1; i > -1; i--) {
        t = container[i];
        if (t && Array.isArray(t)) {
            t[0].parent = container;
            if (this.isContainer(t)) t = this.walk(rules, t, path + '/' + i); // go inside
            if (t === null) container.splice(i, 1);
            else {
                if (x = this.process(rules, t, container, i, path)) container[i] = x; // compressed not null
                else if (x === null) container.splice(i, 1); // null is the mark to delete token
            }
        }
    }
    return container.length ? container : null;
};

CSSOCompressor.prototype.freezeRulesets = function(token, rule, container, i) {
    var info = token[0],
        selector = token[2];

    info.freeze = this.freezeNeeded(selector);
    info.freezeID = this.selectorSignature(selector);
    info.pseudoID = this.composePseudoID(selector);
    info.pseudoSignature = this.pseudoSelectorSignature(selector, this.allowedPClasses, true);
    this.markSimplePseudo(selector);

    return token;
};

CSSOCompressor.prototype.markSimplePseudo = function(selector) {
    var ss, sg = {};

    for (var i = 2; i < selector.length; i++) {
        ss = selector[i];
        ss[0].pseudo = this.containsPseudo(ss);
        ss[0].sg = sg;
        sg[ss[0].s] = 1;
    }
};

CSSOCompressor.prototype.composePseudoID = function(selector) {
    var a = [], ss;

    for (var i = 2; i < selector.length; i++) {
        ss = selector[i];
        if (this.containsPseudo(ss)) {
            a.push(ss[0].s);
        }
    }

    a.sort();

    return a.join(',');
};

CSSOCompressor.prototype.containsPseudo = function(sselector) {
    for (var j = 2; j < sselector.length; j++) {
        switch (sselector[j][1]) {
            case 'pseudoc':
            case 'pseudoe':
            case 'nthselector':
                if (!(sselector[j][2][2] in this.notFPClasses)) return true;
        }
    }
};

CSSOCompressor.prototype.selectorSignature = function(selector) {
    var a = [];

    for (var i = 2; i < selector.length; i++) {
        a.push(translator.translate(cleanInfo(selector[i])));
    }

    a.sort();

    return a.join(',');
};

CSSOCompressor.prototype.pseudoSelectorSignature = function(selector, exclude, dontAppendExcludeMark) {
    var a = [], b = {}, ss, wasExclude = false;
    exclude = exclude || {};

    for (var i = 2; i < selector.length; i++) {
        ss = selector[i];
        for (var j = 2; j < ss.length; j++) {
            switch (ss[j][1]) {
                case 'pseudoc':
                case 'pseudoe':
                case 'nthselector':
                    if (!(ss[j][2][2] in exclude)) b[ss[j][2][2]] = 1;
                    else wasExclude = true;
                    break;
            }
        }
    }

    for (var k in b) a.push(k);

    a.sort();

    return a.join(',') + (dontAppendExcludeMark? '' : wasExclude);
};

CSSOCompressor.prototype.notFPClasses = {
    'link': 1,
    'visited': 1,
    'hover': 1,
    'active': 1,
    'first-letter': 1,
    'first-line': 1
};

CSSOCompressor.prototype.notFPElements = {
    'first-letter': 1,
    'first-line': 1
};

CSSOCompressor.prototype.freezeNeeded = function(selector) {
    var ss;
    for (var i = 2; i < selector.length; i++) {
        ss = selector[i];
        for (var j = 2; j < ss.length; j++) {
            switch (ss[j][1]) {
                case 'pseudoc':
                    if (!(ss[j][2][2] in this.notFPClasses)) return true;
                    break;
                case 'pseudoe':
                    if (!(ss[j][2][2] in this.notFPElements)) return true;
                    break;
                case 'nthselector':
                    return true;
                    break;
            }
        }
    }
    return false;
};

CSSOCompressor.prototype.cleanCharset = function(token, rule, container, i) {
    if (token[2][2][2] === 'charset') {
        for (i = i - 1; i > 1; i--) {
            if (container[i][1] !== 's' && container[i][1] !== 'comment') return null;
        }
    }
};

CSSOCompressor.prototype.cleanImport = function(token, rule, container, i) {
    var x;
    for (i = i - 1; i > 1; i--) {
        x = container[i][1];
        if (x !== 's' && x !== 'comment') {
            if (x === 'atrules') {
                x = container[i][2][2][2];
                if (x !== 'import' && x !== 'charset') return null;
            } else return null;
        }
    }
};

CSSOCompressor.prototype.cleanComment = function(token, rule, container, i) {
    var pr = ((container[1] === 'braces' && i === 4) ||
              (container[1] !== 'braces' && i === 2)) ? null : container[i - 1][1],
        nr = i === container.length - 1 ? null : container[i + 1][1];

    if (nr !== null && pr !== null) {
        if (this._cleanComment(nr) || this._cleanComment(pr)) return null;
    } else return null;
};

CSSOCompressor.prototype._cleanComment = function(r) {
    switch(r) {
        case 's':
        case 'operator':
        case 'attrselector':
        case 'block':
        case 'decldelim':
        case 'ruleset':
        case 'declaration':
        case 'atruleb':
        case 'atrules':
        case 'atruler':
        case 'important':
        case 'nth':
        case 'combinator':
            return true;
    }
};

CSSOCompressor.prototype.nextToken = function(container, type, i, exactly) {
    var t, r;
    for (; i < container.length; i++) {
        t = container[i];
        if (Array.isArray(t)) {
            r = t[1];
            if (r === type) return t;
            else if (exactly && r !== 's') return;
        }
    }
};

CSSOCompressor.prototype.cleanWhitespace = function(token, rule, container, i) {
    var pr = ((container[1] === 'braces' && i === 4) ||
              (container[1] !== 'braces' && i === 2)) ? null : container[i - 1][1],
        nr = i === container.length - 1 ? null : container[i + 1][1];

    if (nr === 'unknown') token[2] = '\n';
    else {
        if (!(container[1] === 'atrulerq' && !pr) && !this.issue16(container, i) && !this.issue165(container, pr, nr) && !this.issue134(pr, nr)) {
            if (nr !== null && pr !== null) {
                if (this._cleanWhitespace(nr, false) || this._cleanWhitespace(pr, true)) return null;
            } else return null;
        }

        token[2] = ' ';
    }

    return token;
};

// See https://github.com/afelix/csso/issues/16
CSSOCompressor.prototype.issue16 = function(container, i) {
    return (i !== 2 && i !== container.length - 1 && container[i - 1][1] === 'uri');
};

//See https://github.com/css/csso/issues/165
CSSOCompressor.prototype.issue165 = function(container, pr, nr) {
    return container[1] === 'atrulerq' && pr === 'braces' && nr === 'ident';
};

//See https://github.com/css/csso/issues/134
CSSOCompressor.prototype.issue134 = function(pr, nr) {
    return pr === 'funktion' && (nr === 'funktion' || nr === 'vhash');
};

CSSOCompressor.prototype._cleanWhitespace = function(r, left) {
    switch(r) {
        case 's':
        case 'operator':
        case 'attrselector':
        case 'block':
        case 'decldelim':
        case 'ruleset':
        case 'declaration':
        case 'atruleb':
        case 'atrules':
        case 'atruler':
        case 'important':
        case 'nth':
        case 'combinator':
            return true;
    }
    if (left) {
        switch(r) {
            case 'funktion':
            case 'braces':
            case 'uri':
                return true;
        }
    }
};

CSSOCompressor.prototype.cleanDecldelim = function(token) {
    for (var i = token.length - 1; i > 1; i--) {
        if (token[i][1] === 'decldelim' &&
            token[i + 1][1] !== 'declaration') token.splice(i, 1);
    }
    if (token[2][1] === 'decldelim') token.splice(2, 1);
    return token;
};

CSSOCompressor.prototype.compressNumber = function(token, rule, container, i) {
    var x = token[2];

    if (/^0*/.test(x)) x = x.replace(/^0+/, '');
    if (/\.0*$/.test(x)) x = x.replace(/\.0*$/, '');
    if (/\..*[1-9]+0+$/.test(x)) x = x.replace(/0+$/, '');
    if (x === '.' || x === '') x = '0';

    token[2] = x;
    token[0].s = x;
    return token;
};

CSSOCompressor.prototype.findDeclaration = function(token) {
    var parent = token;
    while ((parent = parent[0].parent) && parent[1] !== 'declaration');
    return parent;
};

CSSOCompressor.prototype.cleanUnary = function(token, rule, container, i) {
    var next = container[i + 1];
    if (next && next[1] === 'number' && next[2] === '0') return null;
    return token;
};

CSSOCompressor.prototype.compressColor = function(token, rule, container, i) {
    switch(rule) {
        case 'vhash':
            return this.compressHashColor(token);
        case 'funktion':
            return this.compressFunctionColor(token);
        case 'ident':
            return this.compressIdentColor(token, rule, container, i);
    }
};

CSSOCompressor.prototype.compressIdentColor = function(token, rule, container) {
    var map = { 'yellow': 'ff0',
                'fuchsia': 'f0f',
                'white': 'fff',
                'black': '000',
                'blue': '00f',
                'aqua': '0ff' },
        allow = { 'value': 1, 'functionBody': 1 },
        _x = token[2].toLowerCase();

    if (container[1] in allow && _x in map) return [{}, 'vhash', map[_x]];
};

CSSOCompressor.prototype.compressHashColor = function(token) {
    return this._compressHashColor(token[2], token[0]);
};

CSSOCompressor.prototype._compressHashColor = function(x, info) {
    var map = { 'f00': 'red',
                'c0c0c0': 'silver',
                '808080': 'gray',
                '800000': 'maroon',
                '800080': 'purple',
                '008000': 'green',
                '808000': 'olive',
                '000080': 'navy',
                '008080': 'teal'},
        _x = x;
    x = x.toLowerCase();

    if (x.length === 6 &&
        x.charAt(0) === x.charAt(1) &&
        x.charAt(2) === x.charAt(3) &&
        x.charAt(4) === x.charAt(5)) x = x.charAt(0) + x.charAt(2) + x.charAt(4);

    return map[x] ? [info, 'string', map[x]] : [info, 'vhash', (x.length < _x.length ? x : _x)];
};

CSSOCompressor.prototype.compressFunctionColor = function(token) {
    var i, v = [], t, h = '', body;

    if (token[2][2] === 'rgb') {
        body = token[3];
        for (i = 2; i < body.length; i++) {
            t = body[i][1];
            if (t === 'number') v.push(body[i]);
            else if (t !== 'operator') { v = []; break }
        }

        // check if color is followed by number
        var parent = token[0].parent;
        var parentIx = parent.indexOf(token);
        var appendSpace = parent[parentIx + 1] && parent[parentIx + 1][1] != 's';

        if (v.length === 3) {
            h += (t = Number(v[0][2]).toString(16)).length === 1 ? '0' + t : t;
            h += (t = Number(v[1][2]).toString(16)).length === 1 ? '0' + t : t;
            h += (t = Number(v[2][2]).toString(16)).length === 1 ? '0' + t : t;
            if (h.length === 6) {
                var vhash = this._compressHashColor(h, {});
                if (appendSpace) {
                    // che: I guess this is not right: modify color token with
                    // indentation, but I can't find any better solution right now
                    vhash[2] += ' ';
                }

                return vhash;
            }
        }
    }
};

CSSOCompressor.prototype.compressDimension = function(token) {
    if (token[2][2] === '0') {
      if (NON_LENGTH_UNIT.indexOf(token[3][2]) >= 0) {
        return;
      }

      return token[2]
    }
};

CSSOCompressor.prototype.compressString = function(token, rule, container) {
    var s = token[2], r = '', c;
    for (var i = 0; i < s.length; i++) {
        c = s.charAt(i);
        if (c === '\\' && s.charAt(i + 1) === '\n') i++;
        else r += c;
    }
//    if (container[1] === 'attrib' && /^('|")[a-zA-Z0-9]*('|")$/.test(r)) {
//        r = r.substring(1, r.length - 1);
//    }
    if (s.length !== r.length) return [{}, 'string', r];
};

CSSOCompressor.prototype.compressFontWeight = function(token) {
    var p = token[2],
        v = token[3];
    if (p[2][2].indexOf('font-weight') !== -1 && v[2][1] === 'ident') {
        if (v[2][2] === 'normal') v[2] = [{}, 'number', '400'];
        else if (v[2][2] === 'bold') v[2] = [{}, 'number', '700'];
        return token;
    }
};

CSSOCompressor.prototype.compressFont = function(token) {
    var p = token[2],
        v = token[3],
        i, x, t;
    if (/font$/.test(p[2][2]) && v.length) {
        v.splice(2, 0, [{}, 's', '']);
        for (i = v.length - 1; i > 2; i--) {
            x = v[i];
            if (x[1] === 'ident') {
                x = x[2];
                if (x === 'bold') v[i] = [{}, 'number', '700'];
                else if (x === 'normal') {
                    t = v[i - 1];
                    if (t[1] === 'operator' && t[2] === '/') v.splice(--i, 2);
                    else v.splice(i, 1);
                    if (v[i - 1][1] === 's') v.splice(--i, 1);
                }
                else if (x === 'medium' && v[i + 1] && v[i + 1][2] !== '/') {
                    v.splice(i, 1);
                    if (v[i - 1][1] === 's') v.splice(--i, 1);
                }
            }
        }
        if (v.length > 2 && v[2][1] === 's') v.splice(2, 1);
        if (v.length === 2) v.push([{}, 'ident', 'normal']);
        return token;
    }
};

CSSOCompressor.prototype.compressBackground = function(token) {
    var p = token[2],
        v = token[3],
        i, x, t,
        n = v[v.length - 1][1] === 'important' ? 3 : 2;
    if (/background$/.test(p[2][2]) && v.length) {
        v.splice(2, 0, [{}, 's', '']);
        for (i = v.length - 1; i > n; i--) {
            x = v[i];
            if (x[1] === 'ident') {
                x = x[2];
                if (x === 'transparent' || x === 'none' || x === 'repeat' || x === 'scroll') {
                    v.splice(i, 1);
                    if (v[i - 1][1] === 's') v.splice(--i, 1);
                }
            }
        }
        if (v.length > 2 && v[2][1] === 's') v.splice(2, 1);
        if (v.length === 2) v.splice(2, 0, [{}, 'number', '0'], [{}, 's', ' '], [{}, 'number', '0']);
        return token;
    }
};

CSSOCompressor.prototype.cleanEmpty = function(token, rule) {
    switch(rule) {
        case 'ruleset':
            if (token[3].length === 2) return null;
            break;
        case 'atruleb':
            if (token[token.length - 1].length < 3) return null;
            break;
        case 'atruler':
            if (token[4].length < 3) return null;
            break;
    }
};

CSSOCompressor.prototype.destroyDelims = function() {
    return null;
};

CSSOCompressor.prototype.preTranslate = function(token) {
    token[0].s = translator.translate(cleanInfo(token));
    return token;
};

CSSOCompressor.prototype.markShorthands = function(token, rule, container, j, path) {
    if (container[1] === 'ruleset') {
        var selector = container[2][2][0].s,
            freeze = container[0].freeze,
            freezeID = container[0].freezeID;
    } else {
        var selector = '',
            freeze = false,
            freezeID = 'fake';
    }
    var x, p, v, imp, s, key, sh,
        pre = this.pathUp(path) + '/' + (freeze ? '&' + freezeID + '&' : '') + selector + '/',
        createNew, shortsI, shortGroupID = container[0].shortGroupID;

    for (var i = token.length - 1; i > -1; i--) {
        createNew = true;
        x = token[i];
        if (x[1] === 'declaration') {
            v = x[3];
            imp = v[v.length - 1][1] === 'important';
            p = x[2][0].s;
            x[0].id = path + '/' + i;
            if (p in TRBL.props) {
                key = pre + TRBL.extractMain(p);
                var shorts = this.shorts2[key] || [];
                shortsI = shorts.length === 0 ? 0 : shorts.length - 1;

                if (!this.lastShortSelector || selector === this.lastShortSelector || shortGroupID === this.lastShortGroupID) {
                    if (shorts.length) {
                        sh = shorts[shortsI];
                        //if (imp && !sh.imp) sh.invalid = true;
                        createNew = false;
                    }
                }

                if (createNew) {
                    x[0].replaceByShort = true;
                    x[0].shorthandKey = { key: key, i: shortsI };
                    sh = new TRBL(p, imp);
                    shorts.push(sh);
                }

                if (!sh.invalid) {
                    x[0].removeByShort = true;
                    x[0].shorthandKey = { key: key, i: shortsI };
                    sh.add(p, v[0].s, v.slice(2), imp);
                }

                this.shorts2[key] = shorts;

                this.lastShortSelector = selector;
                this.lastShortGroupID = shortGroupID;
            }
        }
    }


    return token;
};

CSSOCompressor.prototype.cleanShorthands = function(token) {
    if (token[0].removeByShort || token[0].replaceByShort) {
        var s, t, sKey = token[0].shorthandKey;

        s = this.shorts2[sKey.key][sKey.i];

        if (!s.invalid && s.isOkToMinimize()) {
            if (token[0].replaceByShort) {
                t = [{}, 'declaration', s.getProperty(), s.getValue()];
                t[0].s = translator.translate(cleanInfo(t));
                return t;
            } else return null;
        }
    }
};

CSSOCompressor.prototype.dontRestructure = {
    'src': 1, // https://github.com/afelix/csso/issues/50
    'clip': 1, // https://github.com/afelix/csso/issues/57
    'display': 1 // https://github.com/afelix/csso/issues/71
};

CSSOCompressor.prototype.restructureBlock = function(token, rule, container, j, path) {
    if (container[1] === 'ruleset') {
        var props = this.props,
            isPseudo = container[2][2][0].pseudo,
            selector = container[2][2][0].s,
            freeze = container[0].freeze,
            freezeID = container[0].freezeID,
            pseudoID = container[0].pseudoID,
            sg = container[2][2][0].sg;
    } else {
        var props = {},
            isPseudo = false,
            selector = '',
            freeze = false,
            freezeID = 'fake',
            pseudoID = 'fake',
            sg = {};
    }

    var x, p, v, imp, t,
        pre = this.pathUp(path) + '/' + selector + '/',
        ppre;
    for (var i = token.length - 1; i > -1; i--) {
        x = token[i];
        if (x[1] === 'declaration') {
            v = x[3];
            imp = v[v.length - 1][1] === 'important';
            p = x[2][0].s;
            ppre = this.buildPPre(pre, p, v, x, freeze);
            x[0].id = path + '/' + i;
            if (!this.dontRestructure[p] && (t = props[ppre])) {
                if ((isPseudo && freezeID === t.freezeID) || // pseudo from equal selectors group
                    (!isPseudo && pseudoID === t.pseudoID) || // not pseudo from equal pseudo signature group
                    (isPseudo && pseudoID === t.pseudoID && this.hashInHash(sg, t.sg))) { // pseudo from covered selectors group
                    if (imp && !t.imp) {
                        props[ppre] = { block: token, imp: imp, id: x[0].id, sg: sg,
                                        freeze: freeze, path: path, freezeID: freezeID, pseudoID: pseudoID };
                        this.deleteProperty(t.block, t.id);
                    } else {
                        token.splice(i, 1);
                    }
                }
            } else if (this.needless(p, props, pre, imp, v, x, freeze)) {
                token.splice(i, 1);
            } else {
                props[ppre] = { block: token, imp: imp, id: x[0].id, sg: sg,
                                freeze: freeze, path: path, freezeID: freezeID, pseudoID: pseudoID };
            }
        }
    }
    return token;
};

CSSOCompressor.prototype.buildPPre = function(pre, p, v, d, freeze) {
    var fp = freeze ? 'ft:' : 'ff:';
    if (p.indexOf('background') !== -1) return fp + pre + d[0].s;

    var _v = v.slice(2),
        colorMark = [
            0, // ident, vhash, rgb
            0, // hsl
            0, // hsla
            0  // rgba
        ],
        vID = '';

    for (var i = 0; i < _v.length; i++) {
        if (!vID) vID = this.getVendorIDFromToken(_v[i]);
        switch(_v[i][1]) {
            case 'vhash':
            case 'ident':
                colorMark[0] = 1; break;
            case 'funktion':
                switch(_v[i][2][2]) {
                    case 'rgb':
                        colorMark[0] = 1; break;
                    case 'hsl':
                        colorMark[1] = 1; break;
                    case 'hsla':
                        colorMark[2] = 1; break;
                    case 'rgba':
                        colorMark[3] = 1; break;
                }
                break;
        }
    }

    return fp + pre + p + colorMark.join('') + (vID ? vID : '');
};

CSSOCompressor.prototype.vendorID = {
    '-o-': 'o',
    '-moz-': 'm',
    '-webkit-': 'w',
    '-ms-': 'i',
    '-epub-': 'e',
    '-apple-': 'a',
    '-xv-': 'x',
    '-wap-': 'p'
};

CSSOCompressor.prototype.getVendorIDFromToken = function(token) {
    var vID;
    switch(token[1]) {
        case 'ident':
            if (vID = this.getVendorFromString(token[2])) return this.vendorID[vID];
            break;
        case 'funktion':
            if (vID = this.getVendorFromString(token[2][2])) return this.vendorID[vID];
            break;
    }
};

CSSOCompressor.prototype.getVendorFromString = function(string) {
    var vendor = string.charAt(0), i;
    if (vendor === '-') {
        if ((i = string.indexOf('-', 2)) !== -1) return string.substr(0, i + 1);
    }
    return '';
};

CSSOCompressor.prototype.deleteProperty = function(block, id) {
    var d;
    for (var i = block.length - 1; i > 1; i--) {
        d = block[i];
        if (Array.isArray(d) && d[1] === 'declaration' && d[0].id === id) {
            block.splice(i, 1);
            return;
        }
    }
};

CSSOCompressor.prototype.nlTable = {
    'border-width': ['border'],
    'border-style': ['border'],
    'border-color': ['border'],
    'border-top': ['border'],
    'border-right': ['border'],
    'border-bottom': ['border'],
    'border-left': ['border'],
    'border-top-width': ['border-top', 'border-width', 'border'],
    'border-right-width': ['border-right', 'border-width', 'border'],
    'border-bottom-width': ['border-bottom', 'border-width', 'border'],
    'border-left-width': ['border-left', 'border-width', 'border'],
    'border-top-style': ['border-top', 'border-style', 'border'],
    'border-right-style': ['border-right', 'border-style', 'border'],
    'border-bottom-style': ['border-bottom', 'border-style', 'border'],
    'border-left-style': ['border-left', 'border-style', 'border'],
    'border-top-color': ['border-top', 'border-color', 'border'],
    'border-right-color': ['border-right', 'border-color', 'border'],
    'border-bottom-color': ['border-bottom', 'border-color', 'border'],
    'border-left-color': ['border-left', 'border-color', 'border'],
    'margin-top': ['margin'],
    'margin-right': ['margin'],
    'margin-bottom': ['margin'],
    'margin-left': ['margin'],
    'padding-top': ['padding'],
    'padding-right': ['padding'],
    'padding-bottom': ['padding'],
    'padding-left': ['padding'],
    'font-style': ['font'],
    'font-variant': ['font'],
    'font-weight': ['font'],
    'font-size': ['font'],
    'font-family': ['font'],
    'list-style-type': ['list-style'],
    'list-style-position': ['list-style'],
    'list-style-image': ['list-style']
};

CSSOCompressor.prototype.needless = function(name, props, pre, imp, v, d, freeze) {
    var hack = name.charAt(0);
    if (hack === '*' || hack === '_' || hack === '$') name = name.substr(1);
    else if (hack === '/' && name.charAt(1) === '/') {
        hack = '//';
        name = name.substr(2);
    } else hack = '';

    var vendor = this.getVendorFromString(name),
        prop = name.substr(vendor.length),
        x, t, ppre;

    if (prop in this.nlTable) {
        x = this.nlTable[prop];
        for (var i = 0; i < x.length; i++) {
            ppre = this.buildPPre(pre, hack + vendor + x[i], v, d, freeze);
            if (t = props[ppre]) return (!imp || t.imp);
        }
    }
};

CSSOCompressor.prototype.rejoinRuleset = function(token, rule, container, i) {
    var p = (i === 2 || container[i - 1][1] === 'unknown') ? null : container[i - 1],
        ps = p ? p[2].slice(2) : [],
        pb = p ? p[3].slice(2) : [],
        ts = token[2].slice(2),
        tb = token[3].slice(2),
        ph, th, r;

    if (!tb.length) return null;

    if (ps.length && pb.length && token[0].pseudoSignature == p[0].pseudoSignature) {
        if (token[1] !== p[1]) return;
        // try to join by selectors
        ph = this.getHash(ps);
        th = this.getHash(ts);

        if (this.equalHash(th, ph)) {
            p[3] = p[3].concat(token[3].splice(2));
            return null;
        }
        if (this.okToJoinByProperties(token, p)) {
            // try to join by properties
            r = this.analyze(token, p);
            if (!r.ne1.length && !r.ne2.length) {
                p[2] = this.cleanSelector(p[2].concat(token[2].splice(2)));
                p[2][0].s = translator.translate(cleanInfo(p[2]));
                return null;
            }
        }
    }
};

CSSOCompressor.prototype.okToJoinByProperties = function(r0, r1) {
    var i0 = r0[0], i1 = r1[0];

    // same frozen ruleset
    if (i0.freezeID === i1.freezeID) return true;

    // same pseudo-classes in selectors
    if (i0.pseudoID === i1.pseudoID) return true;

    // different frozen rulesets
    if (i0.freeze && i1.freeze) {
        return this.pseudoSelectorSignature(r0[2], this.allowedPClasses) === this.pseudoSelectorSignature(r1[2], this.allowedPClasses);
    }

    // is it frozen at all?
    return !(i0.freeze || i1.freeze);
};

CSSOCompressor.prototype.allowedPClasses = {
    'after': 1,
    'before': 1
};

CSSOCompressor.prototype.containsOnlyAllowedPClasses = function(selector) {
    var ss;
    for (var i = 2; i < selector.length; i++) {
        ss = selector[i];
        for (var j = 2; j < ss.length; j++) {
            if (ss[j][1] == 'pseudoc' || ss[j][1] == 'pseudoe') {
                if (!(ss[j][2][2] in this.allowedPClasses)) return false;
            }
        }
    }
    return true;
};

CSSOCompressor.prototype.restructureRuleset = function(token, rule, container, i) {
    var p = (i === 2 || container[i - 1][1] === 'unknown') ? null : container[i - 1],
        ps = p ? p[2].slice(2) : [],
        pb = p ? p[3].slice(2) : [],
        tb = token[3].slice(2),
        r, nr;

    if (!tb.length) return null;

    if (ps.length && pb.length && token[0].pseudoSignature == p[0].pseudoSignature) {
        if (token[1] !== p[1]) return;
        // try to join by properties
        r = this.analyze(token, p);

        if (r.eq.length && (r.ne1.length || r.ne2.length)) {
            if (r.ne1.length && !r.ne2.length) { // p in token
                var ns = token[2].slice(2), // TODO: copypaste
                    nss = translator.translate(cleanInfo(token[2])),
                    sl = nss.length + // selector length
                         ns.length - 1, // delims length
                    bl = this.calcLength(r.eq) + // declarations length
                         r.eq.length - 1; // decldelims length
                if (sl < bl) {
                    p[2] = this.cleanSelector(p[2].concat(token[2].slice(2)));
                    token[3].splice(2);
                    token[3] = token[3].concat(r.ne1);
                    return token;
                }
            } else if (r.ne2.length && !r.ne1.length) { // token in p
                var ns = p[2].slice(2),
                    nss = translator.translate(cleanInfo(p[2])),
                    sl = nss.length + // selector length
                         ns.length - 1, // delims length
                    bl = this.calcLength(r.eq) + // declarations length
                         r.eq.length - 1; // decldelims length
                if (sl < bl) {
                    token[2] = this.cleanSelector(p[2].concat(token[2].slice(2)));
                    p[3].splice(2);
                    p[3] = p[3].concat(r.ne2);
                    return token;
                }
            } else { // extract equal block?
                var ns = this.cleanSelector(p[2].concat(token[2].slice(2))),
                    nss = translator.translate(cleanInfo(ns)),
                    rl = nss.length + // selector length
                         ns.length - 1 + // delims length
                         2, // braces length
                    bl = this.calcLength(r.eq) + // declarations length
                         r.eq.length - 1; // decldelims length

                if (bl >= rl) { // ok, it's good enough to extract
                    ns[0].s = nss;
                    nr = [{f:0, l:0}, 'ruleset', ns, [{f:0,l:0}, 'block'].concat(r.eq)];
                    token[3].splice(2);
                    token[3] = token[3].concat(r.ne1);
                    p[3].splice(2);
                    p[3] = p[3].concat(r.ne2);
                    container.splice(i, 0, nr);
                    return nr;
                }
            }
        }
    }
};

CSSOCompressor.prototype.calcLength = function(tokens) {
    var r = 0;
    for (var i = 0; i < tokens.length; i++) r += tokens[i][0].s.length;
    return r;
};

CSSOCompressor.prototype.cleanSelector = function(token) {
    if (token.length === 2) return null;
    var h = {}, s;
    for (var i = 2; i < token.length; i++) {
        s = token[i][0].s;
        if (s in h) token.splice(i, 1), i--;
        else h[s] = 1;
    }

    return token;
};

CSSOCompressor.prototype.analyze = function(r1, r2) {
    var r = { eq: [], ne1: [], ne2: [] };

    if (r1[1] !== r2[1]) return r;

    var b1 = r1[3], b2 = r2[3],
        d1 = b1.slice(2), d2 = b2.slice(2),
        h1, h2, i, x;

    h1 = this.getHash(d1);
    h2 = this.getHash(d2);

    for (i = 0; i < d1.length; i++) {
        x = d1[i];
        if (x[0].s in h2) r.eq.push(x);
        else r.ne1.push(x);
    }

    for (i = 0; i < d2.length; i++) {
        x = d2[i];
        if (!(x[0].s in h1)) r.ne2.push(x);
    }

    return r;
};

CSSOCompressor.prototype.equalHash = function(h0, h1) {
    var k;
    for (k in h0) if (!(k in h1)) return false;
    for (k in h1) if (!(k in h0)) return false;
    return true;
};

CSSOCompressor.prototype.getHash = function(tokens) {
    var r = {};
    for (var i = 0; i < tokens.length; i++) r[tokens[i][0].s] = 1;
    return r;
};

CSSOCompressor.prototype.hashInHash = function(h0, h1) {
    for (var k in h0) if (!(k in h1)) return false;
    return true;
};

CSSOCompressor.prototype.delimSelectors = function(token) {
    for (var i = token.length - 1; i > 2; i--) {
        token.splice(i, 0, [{}, 'delim']);
    }
};

CSSOCompressor.prototype.delimBlocks = function(token) {
    for (var i = token.length - 1; i > 2; i--) {
        token.splice(i, 0, [{}, 'decldelim']);
    }
};

CSSOCompressor.prototype.copyArray = function(a) {
    var r = [], t;

    for (var i = 0; i < a.length; i++) {
        t = a[i];
        if (Array.isArray(t)) r.push(this.copyArray(t));
        else if (typeof t === 'object') r.push(this.copyObject(t));
        else r.push(t);
    }

    return r;
};

CSSOCompressor.prototype.copyObject = function(o) {
    var r = {};
    for (var k in o) r[k] = o[k];
    return r;
};

CSSOCompressor.prototype.pathUp = function(path) {
    return path.substr(0, path.lastIndexOf('/'));
};
var translator = require('./translator.js').translator(),
    cleanInfo = require('./util.js').cleanInfo;

exports.compress = function(tree, ro) {
    return new CSSOCompressor().compress(tree, ro);
};
