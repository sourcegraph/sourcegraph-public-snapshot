var $util = {};

$util.cleanInfo = function(tree) {
    var r = [];
    tree = tree.slice(1);

    tree.forEach(function(e) {
        r.push(Array.isArray(e) ? $util.cleanInfo(e) : e);
    });

    return r;
};

$util.treeToString = function(tree, level) {
    var spaces = $util.dummySpaces(level),
        level = level ? level : 0,
        s = (level ? '\n' + spaces : '') + '[';

    tree.forEach(function(e) {
        s += (Array.isArray(e) ? $util.treeToString(e, level + 1) : e.f !== undefined ? $util.ircToString(e) : ('\'' + e.toString() + '\'')) + ', ';
    });

    return s.substr(0, s.length - 2) + ']';
};

$util.ircToString = function(o) {
    return '{' + o.f + ',' + o.l + '}';
};

$util.dummySpaces = function(num) {
    return '                                                  '.substr(0, num * 2);
};
function srcToCSSP(s, rule, _needInfo) {
var TokenType = {
    StringSQ: 'StringSQ',
    StringDQ: 'StringDQ',
    CommentML: 'CommentML',
    CommentSL: 'CommentSL',

    Newline: 'Newline',
    Space: 'Space',
    Tab: 'Tab',

    ExclamationMark: 'ExclamationMark',         // !
    QuotationMark: 'QuotationMark',             // "
    NumberSign: 'NumberSign',                   // #
    DollarSign: 'DollarSign',                   // $
    PercentSign: 'PercentSign',                 // %
    Ampersand: 'Ampersand',                     // &
    Apostrophe: 'Apostrophe',                   // '
    LeftParenthesis: 'LeftParenthesis',         // (
    RightParenthesis: 'RightParenthesis',       // )
    Asterisk: 'Asterisk',                       // *
    PlusSign: 'PlusSign',                       // +
    Comma: 'Comma',                             // ,
    HyphenMinus: 'HyphenMinus',                 // -
    FullStop: 'FullStop',                       // .
    Solidus: 'Solidus',                         // /
    Colon: 'Colon',                             // :
    Semicolon: 'Semicolon',                     // ;
    LessThanSign: 'LessThanSign',               // <
    EqualsSign: 'EqualsSign',                   // =
    GreaterThanSign: 'GreaterThanSign',         // >
    QuestionMark: 'QuestionMark',               // ?
    CommercialAt: 'CommercialAt',               // @
    LeftSquareBracket: 'LeftSquareBracket',     // [
    ReverseSolidus: 'ReverseSolidus',           // \
    RightSquareBracket: 'RightSquareBracket',   // ]
    CircumflexAccent: 'CircumflexAccent',       // ^
    LowLine: 'LowLine',                         // _
    LeftCurlyBracket: 'LeftCurlyBracket',       // {
    VerticalLine: 'VerticalLine',               // |
    RightCurlyBracket: 'RightCurlyBracket',     // }
    Tilde: 'Tilde',                             // ~

    Identifier: 'Identifier',
    DecimalNumber: 'DecimalNumber'
};

function getTokens(s) {
    var Punctuation,
        urlMode = false,
        blockMode = 0;

    Punctuation = {
        ' ': TokenType.Space,
        '\n': TokenType.Newline,
        '\r': TokenType.Newline,
        '\t': TokenType.Tab,
        '!': TokenType.ExclamationMark,
        '"': TokenType.QuotationMark,
        '#': TokenType.NumberSign,
        '$': TokenType.DollarSign,
        '%': TokenType.PercentSign,
        '&': TokenType.Ampersand,
        '\'': TokenType.Apostrophe,
        '(': TokenType.LeftParenthesis,
        ')': TokenType.RightParenthesis,
        '*': TokenType.Asterisk,
        '+': TokenType.PlusSign,
        ',': TokenType.Comma,
        '-': TokenType.HyphenMinus,
        '.': TokenType.FullStop,
        '/': TokenType.Solidus,
        ':': TokenType.Colon,
        ';': TokenType.Semicolon,
        '<': TokenType.LessThanSign,
        '=': TokenType.EqualsSign,
        '>': TokenType.GreaterThanSign,
        '?': TokenType.QuestionMark,
        '@': TokenType.CommercialAt,
        '[': TokenType.LeftSquareBracket,
    //        '\\': TokenType.ReverseSolidus,
        ']': TokenType.RightSquareBracket,
        '^': TokenType.CircumflexAccent,
        '_': TokenType.LowLine,
        '{': TokenType.LeftCurlyBracket,
        '|': TokenType.VerticalLine,
        '}': TokenType.RightCurlyBracket,
        '~': TokenType.Tilde
    };

    function isDecimalDigit(c) {
        return '0123456789'.indexOf(c) >= 0;
    }

    function throwError(message) {
        throw message;
    }

    var buffer = '',
        tokens = [],
        pos,
        tn = 0,
        ln = 1;

    function _getTokens(s) {
        if (!s) return [];

        tokens = [];

        var c, cn;

        for (pos = 0; pos < s.length; pos++) {
            c = s.charAt(pos);
            cn = s.charAt(pos + 1);

            if (c === '/' && cn === '*') {
                parseMLComment(s);
            } else if (!urlMode && c === '/' && cn === '/') {
                if (blockMode > 0) parseIdentifier(s);
                else parseSLComment(s);
            } else if (c === '"' || c === "'") {
                parseString(s, c);
            } else if (c === ' ') {
                parseSpaces(s)
            } else if (c in Punctuation) {
                pushToken(Punctuation[c], c);
                if (c === '\n' || c === '\r') ln++;
                if (c === ')') urlMode = false;
                if (c === '{') blockMode++;
                if (c === '}') blockMode--;
            } else if (isDecimalDigit(c)) {
                parseDecimalNumber(s);
            } else {
                parseIdentifier(s);
            }
        }

        mark();

        return tokens;
    }

    function pushToken(type, value) {
        tokens.push({ tn: tn++, ln: ln, type: type, value: value });
    }

    function parseSpaces(s) {
        var start = pos;

        for (; pos < s.length; pos++) {
            if (s.charAt(pos) !== ' ') break;
        }

        pushToken(TokenType.Space, s.substring(start, pos));
        pos--;
    }

    function parseMLComment(s) {
        var start = pos;

        for (pos = pos + 2; pos < s.length; pos++) {
            if (s.charAt(pos) === '*') {
                if (s.charAt(pos + 1) === '/') {
                    pos++;
                    break;
                }
            }
        }

        pushToken(TokenType.CommentML, s.substring(start, pos + 1));
    }

    function parseSLComment(s) {
        var start = pos;

        for (pos = pos + 2; pos < s.length; pos++) {
            if (s.charAt(pos) === '\n' || s.charAt(pos) === '\r') {
                pos++;
                break;
            }
        }

        pushToken(TokenType.CommentSL, s.substring(start, pos));
        pos--;
    }

    function parseString(s, q) {
        var start = pos;

        for (pos = pos + 1; pos < s.length; pos++) {
            if (s.charAt(pos) === '\\') pos++;
            else if (s.charAt(pos) === q) break;
        }

        pushToken(q === '"' ? TokenType.StringDQ : TokenType.StringSQ, s.substring(start, pos + 1));
    }

    function parseDecimalNumber(s) {
        var start = pos;

        for (; pos < s.length; pos++) {
            if (!isDecimalDigit(s.charAt(pos))) break;
        }

        pushToken(TokenType.DecimalNumber, s.substring(start, pos));
        pos--;
    }

    function parseIdentifier(s) {
        var start = pos;

        while (s.charAt(pos) === '/') pos++;

        for (; pos < s.length; pos++) {
            if (s.charAt(pos) === '\\') pos++;
            else if (s.charAt(pos) in Punctuation) break;
        }

        var ident = s.substring(start, pos);

        urlMode = urlMode || ident === 'url';

        pushToken(TokenType.Identifier, ident);
        pos--;
    }

    // ====================================
    // second run
    // ====================================

    function mark() {
        var ps = [], // Parenthesis
            sbs = [], // SquareBracket
            cbs = [], // CurlyBracket
            t;

        for (var i = 0; i < tokens.length; i++) {
            t = tokens[i];
            switch(t.type) {
                case TokenType.LeftParenthesis:
                    ps.push(i);
                    break;
                case TokenType.RightParenthesis:
                    if (ps.length) {
                        t.left = ps.pop();
                        tokens[t.left].right = i;
                    }
                    break;
                case TokenType.LeftSquareBracket:
                    sbs.push(i);
                    break;
                case TokenType.RightSquareBracket:
                    if (sbs.length) {
                        t.left = sbs.pop();
                        tokens[t.left].right = i;
                    }
                    break;
                case TokenType.LeftCurlyBracket:
                    cbs.push(i);
                    break;
                case TokenType.RightCurlyBracket:
                    if (cbs.length) {
                        t.left = cbs.pop();
                        tokens[t.left].right = i;
                    }
                    break;
            }
        }
    }

    return _getTokens(s);
}
// version: 1.0.0

function getCSSPAST(_tokens, rule, _needInfo) {

    var tokens,
        pos,
        failLN = 0,
        currentBlockLN = 0,
        needInfo = false;

    var CSSPNodeType,
        CSSLevel,
        CSSPRules;

    CSSPNodeType = {
        IdentType: 'ident',
        AtkeywordType: 'atkeyword',
        StringType: 'string',
        ShashType: 'shash',
        VhashType: 'vhash',
        NumberType: 'number',
        PercentageType: 'percentage',
        DimensionType: 'dimension',
        CdoType: 'cdo',
        CdcType: 'cdc',
        DecldelimType: 'decldelim',
        SType: 's',
        AttrselectorType: 'attrselector',
        AttribType: 'attrib',
        NthType: 'nth',
        NthselectorType: 'nthselector',
        NamespaceType: 'namespace',
        ClazzType: 'clazz',
        PseudoeType: 'pseudoe',
        PseudocType: 'pseudoc',
        DelimType: 'delim',
        StylesheetType: 'stylesheet',
        AtrulebType: 'atruleb',
        AtrulesType: 'atrules',
        AtrulerqType: 'atrulerq',
        AtrulersType: 'atrulers',
        AtrulerType: 'atruler',
        BlockType: 'block',
        RulesetType: 'ruleset',
        CombinatorType: 'combinator',
        SimpleselectorType: 'simpleselector',
        SelectorType: 'selector',
        DeclarationType: 'declaration',
        PropertyType: 'property',
        ImportantType: 'important',
        UnaryType: 'unary',
        OperatorType: 'operator',
        BracesType: 'braces',
        ValueType: 'value',
        ProgidType: 'progid',
        FiltervType: 'filterv',
        FilterType: 'filter',
        CommentType: 'comment',
        UriType: 'uri',
        RawType: 'raw',
        FunctionBodyType: 'functionBody',
        FunktionType: 'funktion',
        FunctionExpressionType: 'functionExpression',
        UnknownType: 'unknown'
    };

    CSSPRules = {
        'ident': function() { if (checkIdent(pos)) return getIdent() },
        'atkeyword': function() { if (checkAtkeyword(pos)) return getAtkeyword() },
        'string': function() { if (checkString(pos)) return getString() },
        'shash': function() { if (checkShash(pos)) return getShash() },
        'vhash': function() { if (checkVhash(pos)) return getVhash() },
        'number': function() { if (checkNumber(pos)) return getNumber() },
        'percentage': function() { if (checkPercentage(pos)) return getPercentage() },
        'dimension': function() { if (checkDimension(pos)) return getDimension() },
//        'cdo': function() { if (checkCDO()) return getCDO() },
//        'cdc': function() { if (checkCDC()) return getCDC() },
        'decldelim': function() { if (checkDecldelim(pos)) return getDecldelim() },
        's': function() { if (checkS(pos)) return getS() },
        'attrselector': function() { if (checkAttrselector(pos)) return getAttrselector() },
        'attrib': function() { if (checkAttrib(pos)) return getAttrib() },
        'nth': function() { if (checkNth(pos)) return getNth() },
        'nthselector': function() { if (checkNthselector(pos)) return getNthselector() },
        'namespace': function() { if (checkNamespace(pos)) return getNamespace() },
        'clazz': function() { if (checkClazz(pos)) return getClazz() },
        'pseudoe': function() { if (checkPseudoe(pos)) return getPseudoe() },
        'pseudoc': function() { if (checkPseudoc(pos)) return getPseudoc() },
        'delim': function() { if (checkDelim(pos)) return getDelim() },
        'stylesheet': function() { if (checkStylesheet(pos)) return getStylesheet() },
        'atruleb': function() { if (checkAtruleb(pos)) return getAtruleb() },
        'atrules': function() { if (checkAtrules(pos)) return getAtrules() },
        'atrulerq': function() { if (checkAtrulerq(pos)) return getAtrulerq() },
        'atrulers': function() { if (checkAtrulers(pos)) return getAtrulers() },
        'atruler': function() { if (checkAtruler(pos)) return getAtruler() },
        'block': function() { if (checkBlock(pos)) return getBlock() },
        'ruleset': function() { if (checkRuleset(pos)) return getRuleset() },
        'combinator': function() { if (checkCombinator(pos)) return getCombinator() },
        'simpleselector': function() { if (checkSimpleselector(pos)) return getSimpleSelector() },
        'selector': function() { if (checkSelector(pos)) return getSelector() },
        'declaration': function() { if (checkDeclaration(pos)) return getDeclaration() },
        'property': function() { if (checkProperty(pos)) return getProperty() },
        'important': function() { if (checkImportant(pos)) return getImportant() },
        'unary': function() { if (checkUnary(pos)) return getUnary() },
        'operator': function() { if (checkOperator(pos)) return getOperator() },
        'braces': function() { if (checkBraces(pos)) return getBraces() },
        'value': function() { if (checkValue(pos)) return getValue() },
        'progid': function() { if (checkProgid(pos)) return getProgid() },
        'filterv': function() { if (checkFilterv(pos)) return getFilterv() },
        'filter': function() { if (checkFilter(pos)) return getFilter() },
        'comment': function() { if (checkComment(pos)) return getComment() },
        'uri': function() { if (checkUri(pos)) return getUri() },
        'raw': function() { if (checkRaw(pos)) return getRaw() },
        'funktion': function() { if (checkFunktion(pos)) return getFunktion() },
        'functionExpression': function() { if (checkFunctionExpression(pos)) return getFunctionExpression() },
        'unknown': function() { if (checkUnknown(pos)) return getUnknown() }
    };

    function fail(token) {
        if (token && token.ln > failLN) failLN = token.ln;
    }

    function throwError() {
        console.error('Please check the validity of the CSS block starting from the line #' + currentBlockLN);
        if (process) process.exit(1);
        throw new Error();
    }

    function _getAST(_tokens, rule, _needInfo) {
        tokens = _tokens;
        needInfo = _needInfo;
        pos = 0;

        markSC();

        return rule ? CSSPRules[rule]() : CSSPRules['stylesheet']();
    }

//any = braces | string | percentage | dimension | number | uri | functionExpression | funktion | ident | unary
    function checkAny(_i) {
        return checkBraces(_i) ||
               checkString(_i) ||
               checkPercentage(_i) ||
               checkDimension(_i) ||
               checkNumber(_i) ||
               checkUri(_i) ||
               checkFunctionExpression(_i) ||
               checkFunktion(_i) ||
               checkIdent(_i) ||
               checkUnary(_i);
    }

    function getAny() {
        if (checkBraces(pos)) return getBraces();
        else if (checkString(pos)) return getString();
        else if (checkPercentage(pos)) return getPercentage();
        else if (checkDimension(pos)) return getDimension();
        else if (checkNumber(pos)) return getNumber();
        else if (checkUri(pos)) return getUri();
        else if (checkFunctionExpression(pos)) return getFunctionExpression();
        else if (checkFunktion(pos)) return getFunktion();
        else if (checkIdent(pos)) return getIdent();
        else if (checkUnary(pos)) return getUnary();
    }

//atkeyword = '@' ident:x -> [#atkeyword, x]
    function checkAtkeyword(_i) {
        var l;

        if (tokens[_i++].type !== TokenType.CommercialAt) return fail(tokens[_i - 1]);

        if (l = checkIdent(_i)) return l + 1;

        return fail(tokens[_i]);
    }

    function getAtkeyword() {
        var startPos = pos;

        pos++;

        return needInfo?
            [{ ln: tokens[startPos].ln }, CSSPNodeType.AtkeywordType, getIdent()]:
            [CSSPNodeType.AtkeywordType, getIdent()];
    }

//attrib = '[' sc*:s0 ident:x sc*:s1 attrselector:a sc*:s2 (ident | string):y sc*:s3 ']' -> this.concat([#attrib], s0, [x], s1, [a], s2, [y], s3)
//       | '[' sc*:s0 ident:x sc*:s1 ']' -> this.concat([#attrib], s0, [x], s1),
    function checkAttrib(_i) {
        if (tokens[_i].type !== TokenType.LeftSquareBracket) return fail(tokens[_i]);

        if (!tokens[_i].right) return fail(tokens[_i]);

        return tokens[_i].right - _i + 1;
    }

    function checkAttrib1(_i) {
        var start = _i;

        _i++;

        var l = checkSC(_i); // s0

        if (l) _i += l;

        if (l = checkIdent(_i)) _i += l; // x
        else return fail(tokens[_i]);

        if (l = checkSC(_i)) _i += l; // s1

        if (l = checkAttrselector(_i)) _i += l; // a
        else return fail(tokens[_i]);

        if (l = checkSC(_i)) _i += l; // s2

        if ((l = checkIdent(_i)) || (l = checkString(_i))) _i += l; // y
        else return fail(tokens[_i]);

        if (l = checkSC(_i)) _i += l; // s3

        if (tokens[_i].type === TokenType.RightSquareBracket) return _i - start;

        return fail(tokens[_i]);
    }

    function getAttrib1() {
        var startPos = pos;

        pos++;

        var a = (needInfo? [{ ln: tokens[startPos].ln }, CSSPNodeType.AttribType] : [CSSPNodeType.AttribType])
                .concat(getSC())
                .concat([getIdent()])
                .concat(getSC())
                .concat([getAttrselector()])
                .concat(getSC())
                .concat([checkString(pos)? getString() : getIdent()])
                .concat(getSC());

        pos++;

        return a;
    }

    function checkAttrib2(_i) {
        var start = _i;

        _i++;

        var l = checkSC(_i);

        if (l) _i += l;

        if (l = checkIdent(_i)) _i += l;

        if (l = checkSC(_i)) _i += l;

        if (tokens[_i].type === TokenType.RightSquareBracket) return _i - start;

        return fail(tokens[_i]);
    }

    function getAttrib2() {
        var startPos = pos;

        pos++;

        var a = (needInfo? [{ ln: tokens[startPos].ln }, CSSPNodeType.AttribType] : [CSSPNodeType.AttribType])
                .concat(getSC())
                .concat([getIdent()])
                .concat(getSC());

        pos++;

        return a;
    }

    function getAttrib() {
        if (checkAttrib1(pos)) return getAttrib1();
        if (checkAttrib2(pos)) return getAttrib2();
    }

//attrselector = (seq('=') | seq('~=') | seq('^=') | seq('$=') | seq('*=') | seq('|=')):x -> [#attrselector, x]
    function checkAttrselector(_i) {
        if (tokens[_i].type === TokenType.EqualsSign) return 1;
        if (tokens[_i].type === TokenType.VerticalLine && (!tokens[_i + 1] || tokens[_i + 1].type !== TokenType.EqualsSign)) return 1;

        if (!tokens[_i + 1] || tokens[_i + 1].type !== TokenType.EqualsSign) return fail(tokens[_i]);

        switch(tokens[_i].type) {
            case TokenType.Tilde:
            case TokenType.CircumflexAccent:
            case TokenType.DollarSign:
            case TokenType.Asterisk:
            case TokenType.VerticalLine:
                return 2;
        }

        return fail(tokens[_i]);
    }

    function getAttrselector() {
        var startPos = pos,
            s = tokens[pos++].value;

        if (tokens[pos] && tokens[pos].type === TokenType.EqualsSign) s += tokens[pos++].value;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.AttrselectorType, s] :
                [CSSPNodeType.AttrselectorType, s];
    }

//atrule = atruler | atruleb | atrules
    function checkAtrule(_i) {
        var start = _i,
            l;

        if (tokens[start].atrule_l !== undefined) return tokens[start].atrule_l;

        if (l = checkAtruler(_i)) tokens[_i].atrule_type = 1;
        else if (l = checkAtruleb(_i)) tokens[_i].atrule_type = 2;
        else if (l = checkAtrules(_i)) tokens[_i].atrule_type = 3;
        else return fail(tokens[start]);

        tokens[start].atrule_l = l;

        return l;
    }

    function getAtrule() {
        switch (tokens[pos].atrule_type) {
            case 1: return getAtruler();
            case 2: return getAtruleb();
            case 3: return getAtrules();
        }
    }

//atruleb = atkeyword:ak tset*:ap block:b -> this.concat([#atruleb, ak], ap, [b])
    function checkAtruleb(_i) {
        var start = _i,
            l;

        if (l = checkAtkeyword(_i)) _i += l;
        else return fail(tokens[_i]);

        if (l = checkTsets(_i)) _i += l;

        if (l = checkBlock(_i)) _i += l;
        else return fail(tokens[_i]);

        return _i - start;
    }

    function getAtruleb() {
        return (needInfo?
                    [{ ln: tokens[pos].ln }, CSSPNodeType.AtrulebType, getAtkeyword()] :
                    [CSSPNodeType.AtrulebType, getAtkeyword()])
                        .concat(getTsets())
                        .concat([getBlock()]);
    }

//atruler = atkeyword:ak atrulerq:x '{' atrulers:y '}' -> [#atruler, ak, x, y]
    function checkAtruler(_i) {
        var start = _i,
            l;

        if (l = checkAtkeyword(_i)) _i += l;
        else return fail(tokens[_i]);

        if (l = checkAtrulerq(_i)) _i += l;

        if (_i < tokens.length && tokens[_i].type === TokenType.LeftCurlyBracket) _i++;
        else return fail(tokens[_i]);

        if (l = checkAtrulers(_i)) _i += l;

        if (_i < tokens.length && tokens[_i].type === TokenType.RightCurlyBracket) _i++;
        else return fail(tokens[_i]);

        return _i - start;
    }

    function getAtruler() {
        var atruler = needInfo?
                        [{ ln: tokens[pos].ln }, CSSPNodeType.AtrulerType, getAtkeyword(), getAtrulerq()] :
                        [CSSPNodeType.AtrulerType, getAtkeyword(), getAtrulerq()];

        pos++;

        atruler.push(getAtrulers());

        pos++;

        return atruler;
    }

//atrulerq = tset*:ap -> [#atrulerq].concat(ap)
    function checkAtrulerq(_i) {
        return checkTsets(_i);
    }

    function getAtrulerq() {
        return (needInfo? [{ ln: tokens[pos].ln }, CSSPNodeType.AtrulerqType] : [CSSPNodeType.AtrulerqType]).concat(getTsets());
    }

//atrulers = sc*:s0 ruleset*:r sc*:s1 -> this.concat([#atrulers], s0, r, s1)
    function checkAtrulers(_i) {
        var start = _i,
            l;

        if (l = checkSC(_i)) _i += l;

        while ((l = checkRuleset(_i)) || (l = checkAtrule(_i)) || (l = checkSC(_i))) {
            _i += l;
        }

        tokens[_i].atrulers_end = 1;

        if (l = checkSC(_i)) _i += l;

        return _i - start;
    }

    function getAtrulers() {
        var atrulers = (needInfo? [{ ln: tokens[pos].ln }, CSSPNodeType.AtrulersType] : [CSSPNodeType.AtrulersType]).concat(getSC()),
            x;

        while (!tokens[pos].atrulers_end) {
            if (checkSC(pos)) {
                atrulers = atrulers.concat(getSC());
            } else if (checkRuleset(pos)) {
                atrulers.push(getRuleset());
            } else {
                atrulers.push(getAtrule());
            }
        }

        return atrulers.concat(getSC());
    }

//atrules = atkeyword:ak tset*:ap ';' -> this.concat([#atrules, ak], ap)
    function checkAtrules(_i) {
        var start = _i,
            l;

        if (l = checkAtkeyword(_i)) _i += l;
        else return fail(tokens[_i]);

        if (l = checkTsets(_i)) _i += l;

        if (_i >= tokens.length) return _i - start;

        if (tokens[_i].type === TokenType.Semicolon) _i++;
        else return fail(tokens[_i]);

        return _i - start;
    }

    function getAtrules() {
        var atrules = (needInfo? [{ ln: tokens[pos].ln }, CSSPNodeType.AtrulesType, getAtkeyword()] : [CSSPNodeType.AtrulesType, getAtkeyword()]).concat(getTsets());

        pos++;

        return atrules;
    }

//block = '{' blockdecl*:x '}' -> this.concatContent([#block], x)
    function checkBlock(_i) {
        if (_i < tokens.length && tokens[_i].type === TokenType.LeftCurlyBracket) return tokens[_i].right - _i + 1;

        return fail(tokens[_i]);
    }

    function getBlock() {
        var block = needInfo? [{ ln: tokens[pos].ln }, CSSPNodeType.BlockType] : [CSSPNodeType.BlockType],
            end = tokens[pos].right;

        pos++;

        while (pos < end) {
            if (checkBlockdecl(pos)) block = block.concat(getBlockdecl());
            else throwError();
        }

        pos = end + 1;

        return block;
    }

//blockdecl = sc*:s0 (filter | declaration):x decldelim:y sc*:s1 -> this.concat(s0, [x], [y], s1)
//          | sc*:s0 (filter | declaration):x sc*:s1 -> this.concat(s0, [x], s1)
//          | sc*:s0 decldelim:x sc*:s1 -> this.concat(s0, [x], s1)
//          | sc+:s0 -> s0

    function checkBlockdecl(_i) {
        var l;

        if (l = _checkBlockdecl0(_i)) tokens[_i].bd_type = 1;
        else if (l = _checkBlockdecl1(_i)) tokens[_i].bd_type = 2;
        else if (l = _checkBlockdecl2(_i)) tokens[_i].bd_type = 3;
        else if (l = _checkBlockdecl3(_i)) tokens[_i].bd_type = 4;
        else return fail(tokens[_i]);

        return l;
    }

    function getBlockdecl() {
        switch (tokens[pos].bd_type) {
            case 1: return _getBlockdecl0();
            case 2: return _getBlockdecl1();
            case 3: return _getBlockdecl2();
            case 4: return _getBlockdecl3();
        }
    }

    //sc*:s0 (filter | declaration):x decldelim:y sc*:s1 -> this.concat(s0, [x], [y], s1)
    function _checkBlockdecl0(_i) {
        var start = _i,
            l;

        if (l = checkSC(_i)) _i += l;

        if (l = checkFilter(_i)) {
            tokens[_i].bd_filter = 1;
            _i += l;
        } else if (l = checkDeclaration(_i)) {
            tokens[_i].bd_decl = 1;
            _i += l;
        } else return fail(tokens[_i]);

        if (_i < tokens.length && (l = checkDecldelim(_i))) _i += l;
        else return fail(tokens[_i]);

        if (l = checkSC(_i)) _i += l;

        return _i - start;
    }

    function _getBlockdecl0() {
        return getSC()
                .concat([tokens[pos].bd_filter? getFilter() : getDeclaration()])
                .concat([getDecldelim()])
                .concat(getSC());
    }

    //sc*:s0 (filter | declaration):x sc*:s1 -> this.concat(s0, [x], s1)
    function _checkBlockdecl1(_i) {
        var start = _i,
            l;

        if (l = checkSC(_i)) _i += l;

        if (l = checkFilter(_i)) {
            tokens[_i].bd_filter = 1;
            _i += l;
        } else if (l = checkDeclaration(_i)) {
            tokens[_i].bd_decl = 1;
            _i += l;
        } else return fail(tokens[_i]);

        if (l = checkSC(_i)) _i += l;

        return _i - start;
    }

    function _getBlockdecl1() {
        return getSC()
                .concat([tokens[pos].bd_filter? getFilter() : getDeclaration()])
                .concat(getSC());
    }

    //sc*:s0 decldelim:x sc*:s1 -> this.concat(s0, [x], s1)
    function _checkBlockdecl2(_i) {
        var start = _i,
            l;

        if (l = checkSC(_i)) _i += l;

        if (l = checkDecldelim(_i)) _i += l;
        else return fail(tokens[_i]);

        if (l = checkSC(_i)) _i += l;

        return _i - start;
    }

    function _getBlockdecl2() {
        return getSC()
                 .concat([getDecldelim()])
                 .concat(getSC());
    }

    //sc+:s0 -> s0
    function _checkBlockdecl3(_i) {
        return checkSC(_i);
    }

    function _getBlockdecl3() {
        return getSC();
    }

//braces = '(' tset*:x ')' -> this.concat([#braces, '(', ')'], x)
//       | '[' tset*:x ']' -> this.concat([#braces, '[', ']'], x)
    function checkBraces(_i) {
        if (_i >= tokens.length ||
            (tokens[_i].type !== TokenType.LeftParenthesis &&
             tokens[_i].type !== TokenType.LeftSquareBracket)
            ) return fail(tokens[_i]);

        return tokens[_i].right - _i + 1;
    }

    function getBraces() {
        var startPos = pos,
            left = pos,
            right = tokens[pos].right;

        pos++;

        var tsets = getTsets();

        pos++;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.BracesType, tokens[left].value, tokens[right].value].concat(tsets) :
                [CSSPNodeType.BracesType, tokens[left].value, tokens[right].value].concat(tsets);
    }

    function checkCDC(_i) {}

    function checkCDO(_i) {}

    // node: Clazz
    function checkClazz(_i) {
        var l;

        if (tokens[_i].clazz_l) return tokens[_i].clazz_l;

        if (tokens[_i].type === TokenType.FullStop) {
            if (l = checkIdent(_i + 1)) {
                tokens[_i].clazz_l = l + 1;
                return l + 1;
            }
        }

        return fail(tokens[_i]);
    }

    function getClazz() {
        var startPos = pos;

        pos++;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.ClazzType, getIdent()] :
                [CSSPNodeType.ClazzType, getIdent()];
    }

    // node: Combinator
    function checkCombinator(_i) {
        if (tokens[_i].type === TokenType.PlusSign ||
            tokens[_i].type === TokenType.GreaterThanSign ||
            tokens[_i].type === TokenType.Tilde) return 1;

        return fail(tokens[_i]);
    }

    function getCombinator() {
        return needInfo?
                [{ ln: tokens[pos].ln }, CSSPNodeType.CombinatorType, tokens[pos++].value] :
                [CSSPNodeType.CombinatorType, tokens[pos++].value];
    }

    // node: Comment
    function checkComment(_i) {
        if (tokens[_i].type === TokenType.CommentML) return 1;

        return fail(tokens[_i]);
    }

    function getComment() {
        var startPos = pos,
            s = tokens[pos].value.substring(2),
            l = s.length;

        if (s.charAt(l - 2) === '*' && s.charAt(l - 1) === '/') s = s.substring(0, l - 2);

        pos++;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.CommentType, s] :
                [CSSPNodeType.CommentType, s];
    }

    // declaration = property:x ':' value:y -> [#declaration, x, y]
    function checkDeclaration(_i) {
        var start = _i,
            l;

        if (l = checkProperty(_i)) _i += l;
        else return fail(tokens[_i]);

        if (_i < tokens.length && tokens[_i].type === TokenType.Colon) _i++;
        else return fail(tokens[_i]);

        if (l = checkValue(_i)) _i += l;
        else return fail(tokens[_i]);

        return _i - start;
    }

    function getDeclaration() {
        var declaration = needInfo?
                [{ ln: tokens[pos].ln }, CSSPNodeType.DeclarationType, getProperty()] :
                [CSSPNodeType.DeclarationType, getProperty()];

        pos++;

        declaration.push(getValue());

        return declaration;
    }

    // node: Decldelim
    function checkDecldelim(_i) {
        if (_i < tokens.length && tokens[_i].type === TokenType.Semicolon) return 1;

        return fail(tokens[_i]);
    }

    function getDecldelim() {
        var startPos = pos;

        pos++;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.DecldelimType] :
                [CSSPNodeType.DecldelimType];
    }

    // node: Delim
    function checkDelim(_i) {
        if (_i < tokens.length && tokens[_i].type === TokenType.Comma) return 1;

        if (_i >= tokens.length) return fail(tokens[tokens.length - 1]);

        return fail(tokens[_i]);
    }

    function getDelim() {
        var startPos = pos;

        pos++;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.DelimType] :
                [CSSPNodeType.DelimType];
    }

    // node: Dimension
    function checkDimension(_i) {
        var ln = checkNumber(_i),
            li;

        if (!ln || (ln && _i + ln >= tokens.length)) return fail(tokens[_i]);

        if (li = checkNmName2(_i + ln)) return ln + li;

        return fail(tokens[_i]);
    }

    function getDimension() {
        var startPos = pos,
            n = getNumber(),
            dimension = needInfo ?
                [{ ln: tokens[pos].ln }, CSSPNodeType.IdentType, getNmName2()] :
                [CSSPNodeType.IdentType, getNmName2()];

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.DimensionType, n, dimension] :
                [CSSPNodeType.DimensionType, n, dimension];
    }

//filter = filterp:x ':' filterv:y -> [#filter, x, y]
    function checkFilter(_i) {
        var start = _i,
            l;

        if (l = checkFilterp(_i)) _i += l;
        else return fail(tokens[_i]);

        if (tokens[_i].type === TokenType.Colon) _i++;
        else return fail(tokens[_i]);

        if (l = checkFilterv(_i)) _i += l;
        else return fail(tokens[_i]);

        return _i - start;
    }

    function getFilter() {
        var filter = needInfo?
                [{ ln: tokens[pos].ln }, CSSPNodeType.FilterType, getFilterp()] :
                [CSSPNodeType.FilterType, getFilterp()];

        pos++;

        filter.push(getFilterv());

        return filter;
    }

//filterp = (seq('-filter') | seq('_filter') | seq('*filter') | seq('-ms-filter') | seq('filter')):t sc*:s0 -> this.concat([#property, [#ident, t]], s0)
    function checkFilterp(_i) {
        var start = _i,
            l,
            x;

        if (_i < tokens.length) {
            if (tokens[_i].value === 'filter') l = 1;
            else {
                x = joinValues2(_i, 2);

                if (x === '-filter' || x === '_filter' || x === '*filter') l = 2;
                else {
                    x = joinValues2(_i, 4);

                    if (x === '-ms-filter') l = 4;
                    else return fail(tokens[_i]);
                }
            }

            tokens[start].filterp_l = l;

            _i += l;

            if (checkSC(_i)) _i += l;

            return _i - start;
        }

        return fail(tokens[_i]);
    }

    function getFilterp() {
        var startPos = pos,
            x = joinValues2(pos, tokens[pos].filterp_l),
            ident = needInfo? [{ ln: tokens[startPos].ln }, CSSPNodeType.IdentType, x] : [CSSPNodeType.IdentType, x];

        pos += tokens[pos].filterp_l;

        return (needInfo? [{ ln: tokens[startPos].ln }, CSSPNodeType.PropertyType, ident] : [CSSPNodeType.PropertyType, ident])
                    .concat(getSC());

    }

//filterv = progid+:x -> [#filterv].concat(x)
    function checkFilterv(_i) {
        var start = _i,
            l;

        if (l = checkProgid(_i)) _i += l;
        else return fail(tokens[_i]);

        while (l = checkProgid(_i)) {
            _i += l;
        }

        tokens[start].last_progid = _i;

        if (_i < tokens.length && (l = checkSC(_i))) _i += l;

        if (_i < tokens.length && (l = checkImportant(_i))) _i += l;

        return _i - start;
    }

    function getFilterv() {
        var filterv = needInfo? [{ ln: tokens[pos].ln }, CSSPNodeType.FiltervType] : [CSSPNodeType.FiltervType],
            last_progid = tokens[pos].last_progid;

        while (pos < last_progid) {
            filterv.push(getProgid());
        }

        filterv = filterv.concat(checkSC(pos) ? getSC() : []);

        if (pos < tokens.length && checkImportant(pos)) filterv.push(getImportant());

        return filterv;
    }

//functionExpression = ``expression('' functionExpressionBody*:x ')' -> [#functionExpression, x.join('')],
    function checkFunctionExpression(_i) {
        var start = _i;

        if (!tokens[_i] || tokens[_i++].value !== 'expression') return fail(tokens[_i - 1]);

        if (!tokens[_i] || tokens[_i].type !== TokenType.LeftParenthesis) return fail(tokens[_i]);

        return tokens[_i].right - start + 1;
    }

    function getFunctionExpression() {
        var startPos = pos;

        pos++;

        var e = joinValues(pos + 1, tokens[pos].right - 1);

        pos = tokens[pos].right + 1;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.FunctionExpressionType, e] :
                [CSSPNodeType.FunctionExpressionType, e];
    }

//funktion = ident:x '(' functionBody:y ')' -> [#funktion, x, y]
    function checkFunktion(_i) {
        var start = _i,
            l = checkIdent(_i);

        if (!l) return fail(tokens[_i]);

        _i += l;

        if (_i >= tokens.length || tokens[_i].type !== TokenType.LeftParenthesis) return fail(tokens[_i - 1]);

        return tokens[_i].right - start + 1;
    }

    function getFunktion() {
        var startPos = pos,
            ident = getIdent();

        pos++;

        var body = ident[needInfo? 2 : 1] !== 'not'?
            getFunctionBody() :
            getNotFunctionBody(); // ok, here we have CSS3 initial draft: http://dev.w3.org/csswg/selectors3/#negation

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.FunktionType, ident, body] :
                [CSSPNodeType.FunktionType, ident, body];
    }

    function getFunctionBody() {
        var startPos = pos,
            body = [],
            x;

        while (tokens[pos].type !== TokenType.RightParenthesis) {
            if (checkTset(pos)) {
                x = getTset();
                if ((needInfo && typeof x[1] === 'string') || typeof x[0] === 'string') body.push(x);
                else body = body.concat(x);
            } else if (checkClazz(pos)) {
                body.push(getClazz());
            } else {
                throwError();
            }
        }

        pos++;

        return (needInfo?
                    [{ ln: tokens[startPos].ln }, CSSPNodeType.FunctionBodyType] :
                    [CSSPNodeType.FunctionBodyType]
                ).concat(body);
    }

    function getNotFunctionBody() {
        var startPos = pos,
            body = [],
            x;

        while (tokens[pos].type !== TokenType.RightParenthesis) {
            if (checkSimpleselector(pos)) {
                body.push(getSimpleSelector());
            } else {
                throwError();
            }
        }

        pos++;

        return (needInfo?
                    [{ ln: tokens[startPos].ln }, CSSPNodeType.FunctionBodyType] :
                    [CSSPNodeType.FunctionBodyType]
                ).concat(body);
    }

    // node: Ident
    function checkIdent(_i) {
        var start = _i,
            wasIdent = false;

        // start char / word
        if (_i < tokens.length &&
            (tokens[_i].type === TokenType.HyphenMinus ||
            tokens[_i].type === TokenType.LowLine ||
            tokens[_i].type === TokenType.Identifier ||
            tokens[_i].type === TokenType.DollarSign ||
            tokens[_i].type === TokenType.Asterisk)) _i++;
        else return fail(tokens[_i]);

        wasIdent = tokens[_i - 1].type === TokenType.Identifier;

        for (; _i < tokens.length; _i++) {
            if (tokens[_i].type !== TokenType.HyphenMinus &&
                tokens[_i].type !== TokenType.LowLine) {
                    if (tokens[_i].type !== TokenType.Identifier &&
                        (tokens[_i].type !== TokenType.DecimalNumber || !wasIdent)
                        ) break;
                    else wasIdent = true;
            }
        }

        if (!wasIdent && tokens[start].type !== TokenType.Asterisk) return fail(tokens[_i]);

        tokens[start].ident_last = _i - 1;

        return _i - start;
    }

    function getIdent() {
        var startPos = pos,
            s = joinValues(pos, tokens[pos].ident_last);

        pos = tokens[pos].ident_last + 1;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.IdentType, s] :
                [CSSPNodeType.IdentType, s];
    }

//important = '!' sc*:s0 seq('important') -> [#important].concat(s0)
    function checkImportant(_i) {
        var start = _i,
            l;

        if (tokens[_i++].type !== TokenType.ExclamationMark) return fail(tokens[_i - 1]);

        if (l = checkSC(_i)) _i += l;

        if (tokens[_i].value !== 'important') return fail(tokens[_i]);

        return _i - start + 1;
    }

    function getImportant() {
        var startPos = pos;

        pos++;

        var sc = getSC();

        pos++;

        return (needInfo? [{ ln: tokens[startPos].ln }, CSSPNodeType.ImportantType] : [CSSPNodeType.ImportantType]).concat(sc);
    }

    // node: Namespace
    function checkNamespace(_i) {
        if (tokens[_i].type === TokenType.VerticalLine) return 1;

        return fail(tokens[_i]);
    }

    function getNamespace() {
        var startPos = pos;

        pos++;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.NamespaceType] :
                [CSSPNodeType.NamespaceType];
    }

//nth = (digit | 'n')+:x -> [#nth, x.join('')]
//    | (seq('even') | seq('odd')):x -> [#nth, x]
    function checkNth(_i) {
        return checkNth1(_i) || checkNth2(_i);
    }

    function checkNth1(_i) {
        var start = _i;

        for (; _i < tokens.length; _i++) {
            if (tokens[_i].type !== TokenType.DecimalNumber && tokens[_i].value !== 'n') break;
        }

        if (_i !== start) {
            tokens[start].nth_last = _i - 1;
            return _i - start;
        }

        return fail(tokens[_i]);
    }

    function getNth() {
        var startPos = pos;

        if (tokens[pos].nth_last) {
            var n = needInfo?
                        [{ ln: tokens[startPos].ln }, CSSPNodeType.NthType, joinValues(pos, tokens[pos].nth_last)] :
                        [CSSPNodeType.NthType, joinValues(pos, tokens[pos].nth_last)];

            pos = tokens[pos].nth_last + 1;

            return n;
        }

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.NthType, tokens[pos++].value] :
                [CSSPNodeType.NthType, tokens[pos++].value];
    }

    function checkNth2(_i) {
        if (tokens[_i].value === 'even' || tokens[_i].value === 'odd') return 1;

        return fail(tokens[_i]);
    }

//nthf = ':' seq('nth-'):x (seq('child') | seq('last-child') | seq('of-type') | seq('last-of-type')):y -> (x + y)
    function checkNthf(_i) {
        var start = _i,
            l = 0;

        if (tokens[_i++].type !== TokenType.Colon) return fail(tokens[_i - 1]); l++;

        if (tokens[_i++].value !== 'nth' || tokens[_i++].value !== '-') return fail(tokens[_i - 1]); l += 2;

        if ('child' === tokens[_i].value) {
            l += 1;
        } else if ('last-child' === tokens[_i].value +
                                    tokens[_i + 1].value +
                                    tokens[_i + 2].value) {
            l += 3;
        } else if ('of-type' === tokens[_i].value +
                                 tokens[_i + 1].value +
                                 tokens[_i + 2].value) {
            l += 3;
        } else if ('last-of-type' === tokens[_i].value +
                                      tokens[_i + 1].value +
                                      tokens[_i + 2].value +
                                      tokens[_i + 3].value +
                                      tokens[_i + 4].value) {
            l += 5;
        } else return fail(tokens[_i]);

        tokens[start + 1].nthf_last = start + l - 1;

        return l;
    }

    function getNthf() {
        pos++;

        var s = joinValues(pos, tokens[pos].nthf_last);

        pos = tokens[pos].nthf_last + 1;

        return s;
    }

//nthselector = nthf:x '(' (sc | unary | nth)*:y ')' -> [#nthselector, [#ident, x]].concat(y)
    function checkNthselector(_i) {
        var start = _i,
            l;

        if (l = checkNthf(_i)) _i += l;
        else return fail(tokens[_i]);

        if (tokens[_i].type !== TokenType.LeftParenthesis || !tokens[_i].right) return fail(tokens[_i]);

        l++;

        var rp = tokens[_i++].right;

        while (_i < rp) {
            if (l = checkSC(_i)) _i += l;
            else if (l = checkUnary(_i)) _i += l;
            else if (l = checkNth(_i)) _i += l;
            else return fail(tokens[_i]);
        }

        return rp - start + 1;
    }

    function getNthselector() {
        var startPos = pos,
            nthf = needInfo?
                    [{ ln: tokens[pos].ln }, CSSPNodeType.IdentType, getNthf()] :
                    [CSSPNodeType.IdentType, getNthf()],
            ns = needInfo?
                    [{ ln: tokens[pos].ln }, CSSPNodeType.NthselectorType, nthf] :
                    [CSSPNodeType.NthselectorType, nthf];

        pos++;

        while (tokens[pos].type !== TokenType.RightParenthesis) {
            if (checkSC(pos)) ns = ns.concat(getSC());
            else if (checkUnary(pos)) ns.push(getUnary());
            else if (checkNth(pos)) ns.push(getNth());
        }

        pos++;

        return ns;
    }

    // node: Number
    function checkNumber(_i) {
        if (_i < tokens.length && tokens[_i].number_l) return tokens[_i].number_l;

        if (_i < tokens.length && tokens[_i].type === TokenType.DecimalNumber &&
            (!tokens[_i + 1] ||
             (tokens[_i + 1] && tokens[_i + 1].type !== TokenType.FullStop))
        ) return (tokens[_i].number_l = 1, tokens[_i].number_l); // 10

        if (_i < tokens.length &&
             tokens[_i].type === TokenType.DecimalNumber &&
             tokens[_i + 1] && tokens[_i + 1].type === TokenType.FullStop &&
             (!tokens[_i + 2] || (tokens[_i + 2].type !== TokenType.DecimalNumber))
        ) return (tokens[_i].number_l = 2, tokens[_i].number_l); // 10.

        if (_i < tokens.length &&
            tokens[_i].type === TokenType.FullStop &&
            tokens[_i + 1].type === TokenType.DecimalNumber
        ) return (tokens[_i].number_l = 2, tokens[_i].number_l); // .10

        if (_i < tokens.length &&
            tokens[_i].type === TokenType.DecimalNumber &&
            tokens[_i + 1] && tokens[_i + 1].type === TokenType.FullStop &&
            tokens[_i + 2] && tokens[_i + 2].type === TokenType.DecimalNumber
        ) return (tokens[_i].number_l = 3, tokens[_i].number_l); // 10.10

        return fail(tokens[_i]);
    }

    function getNumber() {
        var s = '',
            startPos = pos,
            l = tokens[pos].number_l;

        for (var i = 0; i < l; i++) {
            s += tokens[pos + i].value;
        }

        pos += l;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.NumberType, s] :
                [CSSPNodeType.NumberType, s];
    }

    // node: Operator
    function checkOperator(_i) {
        if (_i < tokens.length &&
            (tokens[_i].type === TokenType.Solidus ||
            tokens[_i].type === TokenType.Comma ||
            tokens[_i].type === TokenType.Colon ||
            tokens[_i].type === TokenType.EqualsSign)) return 1;

        return fail(tokens[_i]);
    }

    function getOperator() {
        return needInfo?
                [{ ln: tokens[pos].ln }, CSSPNodeType.OperatorType, tokens[pos++].value] :
                [CSSPNodeType.OperatorType, tokens[pos++].value];
    }

    // node: Percentage
    function checkPercentage(_i) {
        var x = checkNumber(_i);

        if (!x || (x && _i + x >= tokens.length)) return fail(tokens[_i]);

        if (tokens[_i + x].type === TokenType.PercentSign) return x + 1;

        return fail(tokens[_i]);
    }

    function getPercentage() {
        var startPos = pos,
            n = getNumber();

        pos++;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.PercentageType, n] :
                [CSSPNodeType.PercentageType, n];
    }

//progid = sc*:s0 seq('progid:DXImageTransform.Microsoft.'):x letter+:y '(' (m_string | m_comment | ~')' char)+:z ')' sc*:s1
//                -> this.concat([#progid], s0, [[#raw, x + y.join('') + '(' + z.join('') + ')']], s1),
    function checkProgid(_i) {
        var start = _i,
            l,
            x;

        if (l = checkSC(_i)) _i += l;

        if ((x = joinValues2(_i, 6)) === 'progid:DXImageTransform.Microsoft.') {
            _start = _i;
            _i += 6;
        } else return fail(tokens[_i - 1]);

        if (l = checkIdent(_i)) _i += l;
        else return fail(tokens[_i]);

        if (l = checkSC(_i)) _i += l;

        if (tokens[_i].type === TokenType.LeftParenthesis) {
            tokens[start].progid_end = tokens[_i].right;
            _i = tokens[_i].right + 1;
        } else return fail(tokens[_i]);

        if (l = checkSC(_i)) _i += l;

        return _i - start;
    }

    function getProgid() {
        var startPos = pos,
            progid_end = tokens[pos].progid_end;

        return (needInfo? [{ ln: tokens[startPos].ln }, CSSPNodeType.ProgidType] : [CSSPNodeType.ProgidType])
                .concat(getSC())
                .concat([_getProgid(progid_end)])
                .concat(getSC());
    }

    function _getProgid(progid_end) {
        var startPos = pos,
            x = joinValues(pos, progid_end);

        pos = progid_end + 1;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.RawType, x] :
                [CSSPNodeType.RawType, x];
    }

//property = ident:x sc*:s0 -> this.concat([#property, x], s0)
    function checkProperty(_i) {
        var start = _i,
            l;

        if (l = checkIdent(_i)) _i += l;
        else return fail(tokens[_i]);

        if (l = checkSC(_i)) _i += l;
        return _i - start;
    }

    function getProperty() {
        var startPos = pos;

        return (needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.PropertyType, getIdent()] :
                [CSSPNodeType.PropertyType, getIdent()])
            .concat(getSC());
    }

    function checkPseudo(_i) {
        return checkPseudoe(_i) ||
               checkPseudoc(_i);
    }

    function getPseudo() {
        if (checkPseudoe(pos)) return getPseudoe();
        if (checkPseudoc(pos)) return getPseudoc();
    }

    function checkPseudoe(_i) {
        var l;

        if (tokens[_i++].type !== TokenType.Colon) return fail(tokens[_i - 1]);

        if (tokens[_i++].type !== TokenType.Colon) return fail(tokens[_i - 1]);

        if (l = checkIdent(_i)) return l + 2;

        return fail(tokens[_i]);
    }

    function getPseudoe() {
        var startPos = pos;

        pos += 2;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.PseudoeType, getIdent()] :
                [CSSPNodeType.PseudoeType, getIdent()];
    }

//pseudoc = ':' (funktion | ident):x -> [#pseudoc, x]
    function checkPseudoc(_i) {
        var l;

        if (tokens[_i++].type !== TokenType.Colon) return fail(tokens[_i - 1]);

        if ((l = checkFunktion(_i)) || (l = checkIdent(_i))) return l + 1;

        return fail(tokens[_i]);
    }

    function getPseudoc() {
        var startPos = pos;

        pos++;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.PseudocType, checkFunktion(pos)? getFunktion() : getIdent()] :
                [CSSPNodeType.PseudocType, checkFunktion(pos)? getFunktion() : getIdent()];
    }

    //ruleset = selector*:x block:y -> this.concat([#ruleset], x, [y])
    function checkRuleset(_i) {
        var start = _i,
            l;

        if (tokens[start].ruleset_l !== undefined) return tokens[start].ruleset_l;

        while (l = checkSelector(_i)) {
            _i += l;
        }

        if (l = checkBlock(_i)) _i += l;
        else return fail(tokens[_i]);

        tokens[start].ruleset_l = _i - start;

        return _i - start;
    }

    function getRuleset() {
        var ruleset = needInfo? [{ ln: tokens[pos].ln }, CSSPNodeType.RulesetType] : [CSSPNodeType.RulesetType];

        while (!checkBlock(pos)) {
            ruleset.push(getSelector());
        }

        ruleset.push(getBlock());

        return ruleset;
    }

    // node: S
    function checkS(_i) {
        if (tokens[_i].ws) return tokens[_i].ws_last - _i + 1;

        return fail(tokens[_i]);
    }

    function getS() {
        var startPos = pos,
            s = joinValues(pos, tokens[pos].ws_last);

        pos = tokens[pos].ws_last + 1;

        return needInfo? [{ ln: tokens[startPos].ln }, CSSPNodeType.SType, s] : [CSSPNodeType.SType, s];
    }

    function checkSC(_i) {
        var l,
            lsc = 0;

        while (_i < tokens.length) {
            if (!(l = checkS(_i)) && !(l = checkComment(_i))) break;
            _i += l;
            lsc += l;
        }

        if (lsc) return lsc;

        if (_i >= tokens.length) return fail(tokens[tokens.length - 1]);

        return fail(tokens[_i]);
    }

    function getSC() {
        var sc = [];

        while (pos < tokens.length) {
            if (checkS(pos)) sc.push(getS());
            else if (checkComment(pos)) sc.push(getComment());
            else break;
        }

        return sc;
    }

    //selector = (simpleselector | delim)+:x -> this.concat([#selector], x)
    function checkSelector(_i) {
        var start = _i,
            l;

        if (_i < tokens.length) {
            while (l = checkSimpleselector(_i) || checkDelim(_i)) {
                _i += l;
            }

            tokens[start].selector_end = _i - 1;

            return _i - start;
        }
    }

    function getSelector() {
        var selector = needInfo? [{ ln: tokens[pos].ln }, CSSPNodeType.SelectorType] : [CSSPNodeType.SelectorType],
            selector_end = tokens[pos].selector_end;

        while (pos <= selector_end) {
            selector.push(checkDelim(pos) ? getDelim() : getSimpleSelector());
        }

        return selector;
    }

    // node: Shash
    function checkShash(_i) {
        if (tokens[_i].type !== TokenType.NumberSign) return fail(tokens[_i]);

        var l = checkNmName(_i + 1);

        if (l) return l + 1;

        return fail(tokens[_i]);
    }

    function getShash() {
        var startPos = pos;

        pos++;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.ShashType, getNmName()] :
                [CSSPNodeType.ShashType, getNmName()];
    }

//simpleselector = (nthselector | combinator | attrib | pseudo | clazz | shash | any | sc | namespace)+:x -> this.concatContent([#simpleselector], [x])
    function checkSimpleselector(_i) {
        var start = _i,
            l;

        while (_i < tokens.length) {
            if (l = _checkSimpleSelector(_i)) _i += l;
            else break;
        }

        if (_i - start) return _i - start;

        if (_i >= tokens.length) return fail(tokens[tokens.length - 1]);

        return fail(tokens[_i]);
    }

    function _checkSimpleSelector(_i) {
        return checkNthselector(_i) ||
               checkCombinator(_i) ||
               checkAttrib(_i) ||
               checkPseudo(_i) ||
               checkClazz(_i) ||
               checkShash(_i) ||
               checkAny(_i) ||
               checkSC(_i) ||
               checkNamespace(_i);
    }

    function getSimpleSelector() {
        var ss = needInfo? [{ ln: tokens[pos].ln }, CSSPNodeType.SimpleselectorType] : [CSSPNodeType.SimpleselectorType],
            t;

        while (pos < tokens.length && _checkSimpleSelector(pos)) {
            t = _getSimpleSelector();

            if ((needInfo && typeof t[1] === 'string') || typeof t[0] === 'string') ss.push(t);
            else ss = ss.concat(t);
        }

        return ss;
    }

    function _getSimpleSelector() {
        if (checkNthselector(pos)) return getNthselector();
        else if (checkCombinator(pos)) return getCombinator();
        else if (checkAttrib(pos)) return getAttrib();
        else if (checkPseudo(pos)) return getPseudo();
        else if (checkClazz(pos)) return getClazz();
        else if (checkShash(pos)) return getShash();
        else if (checkAny(pos)) return getAny();
        else if (checkSC(pos)) return getSC();
        else if (checkNamespace(pos)) return getNamespace();
    }

    // node: String
    function checkString(_i) {
        if (_i < tokens.length &&
            (tokens[_i].type === TokenType.StringSQ || tokens[_i].type === TokenType.StringDQ)
        ) return 1;

        return fail(tokens[_i]);
    }

    function getString() {
        var startPos = pos;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.StringType, tokens[pos++].value] :
                [CSSPNodeType.StringType, tokens[pos++].value];
    }

    //stylesheet = (cdo | cdc | sc | statement)*:x -> this.concat([#stylesheet], x)
    function checkStylesheet(_i) {
        var start = _i,
            l;

        while (_i < tokens.length) {
            if (l = checkSC(_i)) _i += l;
            else {
                currentBlockLN = tokens[_i].ln;
                if (l = checkAtrule(_i)) _i += l;
                else if (l = checkRuleset(_i)) _i += l;
                else if (l = checkUnknown(_i)) _i += l;
                else throwError();
            }
        }

        return _i - start;
    }

    function getStylesheet(_i) {
        var t,
            stylesheet = needInfo? [{ ln: tokens[pos].ln }, CSSPNodeType.StylesheetType] : [CSSPNodeType.StylesheetType];

        while (pos < tokens.length) {
            if (checkSC(pos)) stylesheet = stylesheet.concat(getSC());
            else {
                currentBlockLN = tokens[pos].ln;
                if (checkRuleset(pos)) stylesheet.push(getRuleset());
                else if (checkAtrule(pos)) stylesheet.push(getAtrule());
                else if (checkUnknown(pos)) stylesheet.push(getUnknown());
                else throwError();
            }
        }

        return stylesheet;
    }

//tset = vhash | any | sc | operator
    function checkTset(_i) {
        return checkVhash(_i) ||
               checkAny(_i) ||
               checkSC(_i) ||
               checkOperator(_i);
    }

    function getTset() {
        if (checkVhash(pos)) return getVhash();
        else if (checkAny(pos)) return getAny();
        else if (checkSC(pos)) return getSC();
        else if (checkOperator(pos)) return getOperator();
    }

    function checkTsets(_i) {
        var start = _i,
            l;

        while (l = checkTset(_i)) {
            _i += l;
        }

        return _i - start;
    }

    function getTsets() {
        var tsets = [],
            x;

        while (x = getTset()) {
            if ((needInfo && typeof x[1] === 'string') || typeof x[0] === 'string') tsets.push(x);
            else tsets = tsets.concat(x);
        }

        return tsets;
    }

    // node: Unary
    function checkUnary(_i) {
        if (_i < tokens.length &&
            (tokens[_i].type === TokenType.HyphenMinus ||
            tokens[_i].type === TokenType.PlusSign)
        ) return 1;

        return fail(tokens[_i]);
    }

    function getUnary() {
        var startPos = pos;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.UnaryType, tokens[pos++].value] :
                [CSSPNodeType.UnaryType, tokens[pos++].value];
    }

    // node: Unknown
    function checkUnknown(_i) {
        if (_i < tokens.length && tokens[_i].type === TokenType.CommentSL) return 1;

        return fail(tokens[_i]);
    }

    function getUnknown() {
        var startPos = pos;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.UnknownType, tokens[pos++].value] :
                [CSSPNodeType.UnknownType, tokens[pos++].value];
    }

//    uri = seq('url(') sc*:s0 string:x sc*:s1 ')' -> this.concat([#uri], s0, [x], s1)
//        | seq('url(') sc*:s0 (~')' ~m_w char)*:x sc*:s1 ')' -> this.concat([#uri], s0, [[#raw, x.join('')]], s1),
    function checkUri(_i) {
        var start = _i,
            l;

        if (_i < tokens.length && tokens[_i++].value !== 'url') return fail(tokens[_i - 1]);

        if (!tokens[_i] || tokens[_i].type !== TokenType.LeftParenthesis) return fail(tokens[_i]);

        return tokens[_i].right - start + 1;
    }

    function getUri() {
        var startPos = pos,
            uriExcluding = {};

        pos += 2;

        uriExcluding[TokenType.Space] = 1;
        uriExcluding[TokenType.Tab] = 1;
        uriExcluding[TokenType.Newline] = 1;
        uriExcluding[TokenType.LeftParenthesis] = 1;
        uriExcluding[TokenType.RightParenthesis] = 1;

        if (checkUri1(pos)) {
            var uri = (needInfo? [{ ln: tokens[startPos].ln }, CSSPNodeType.UriType] : [CSSPNodeType.UriType])
                        .concat(getSC())
                        .concat([getString()])
                        .concat(getSC());

            pos++;

            return uri;
        } else {
            var uri = (needInfo? [{ ln: tokens[startPos].ln }, CSSPNodeType.UriType] : [CSSPNodeType.UriType])
                        .concat(getSC()),
                l = checkExcluding(uriExcluding, pos),
                raw = needInfo?
                        [{ ln: tokens[pos].ln }, CSSPNodeType.RawType, joinValues(pos, pos + l)] :
                        [CSSPNodeType.RawType, joinValues(pos, pos + l)];

            uri.push(raw);

            pos += l + 1;

            uri = uri.concat(getSC());

            pos++;

            return uri;
        }
    }

    function checkUri1(_i) {
        var start = _i,
            l = checkSC(_i);

        if (l) _i += l;

        if (tokens[_i].type !== TokenType.StringDQ && tokens[_i].type !== TokenType.StringSQ) return fail(tokens[_i]);

        _i++;

        if (l = checkSC(_i)) _i += l;

        return _i - start;
    }

    // value = (sc | vhash | any | block | atkeyword | operator | important)+:x -> this.concat([#value], x)
    function checkValue(_i) {
        var start = _i,
            l;

        while (_i < tokens.length) {
            if (l = _checkValue(_i)) _i += l;
            else break;
        }

        if (_i - start) return _i - start;

        return fail(tokens[_i]);
    }

    function _checkValue(_i) {
        return checkSC(_i) ||
               checkVhash(_i) ||
               checkAny(_i) ||
               checkBlock(_i) ||
               checkAtkeyword(_i) ||
               checkOperator(_i) ||
               checkImportant(_i);
    }

    function getValue() {
        var ss = needInfo? [{ ln: tokens[pos].ln }, CSSPNodeType.ValueType] : [CSSPNodeType.ValueType],
            t;

        while (pos < tokens.length && _checkValue(pos)) {
            t = _getValue();

            if ((needInfo && typeof t[1] === 'string') || typeof t[0] === 'string') ss.push(t);
            else ss = ss.concat(t);
        }

        return ss;
    }

    function _getValue() {
        if (checkSC(pos)) return getSC();
        else if (checkVhash(pos)) return getVhash();
        else if (checkAny(pos)) return getAny();
        else if (checkBlock(pos)) return getBlock();
        else if (checkAtkeyword(pos)) return getAtkeyword();
        else if (checkOperator(pos)) return getOperator();
        else if (checkImportant(pos)) return getImportant();
    }

    // node: Vhash
    function checkVhash(_i) {
        if (_i >= tokens.length || tokens[_i].type !== TokenType.NumberSign) return fail(tokens[_i]);

        var l = checkNmName2(_i + 1);

        if (l) return l + 1;

        return fail(tokens[_i]);
    }

    function getVhash() {
        var startPos = pos;

        pos++;

        return needInfo?
                [{ ln: tokens[startPos].ln }, CSSPNodeType.VhashType, getNmName2()] :
                [CSSPNodeType.VhashType, getNmName2()];
    }

    function checkNmName(_i) {
        var start = _i;

        // start char / word
        if (tokens[_i].type === TokenType.HyphenMinus ||
            tokens[_i].type === TokenType.LowLine ||
            tokens[_i].type === TokenType.Identifier ||
            tokens[_i].type === TokenType.DecimalNumber) _i++;
        else return fail(tokens[_i]);

        for (; _i < tokens.length; _i++) {
            if (tokens[_i].type !== TokenType.HyphenMinus &&
                tokens[_i].type !== TokenType.LowLine &&
                tokens[_i].type !== TokenType.Identifier &&
                tokens[_i].type !== TokenType.DecimalNumber) break;
        }

        tokens[start].nm_name_last = _i - 1;

        return _i - start;
    }

    function getNmName() {
        var s = joinValues(pos, tokens[pos].nm_name_last);

        pos = tokens[pos].nm_name_last + 1;

        return s;
    }

    function checkNmName2(_i) {
        var start = _i;

        if (tokens[_i].type === TokenType.Identifier) return 1;
        else if (tokens[_i].type !== TokenType.DecimalNumber) return fail(tokens[_i]);

        _i++;

        if (!tokens[_i] || tokens[_i].type !== TokenType.Identifier) return 1;

        return 2;
    }

    function getNmName2() {
        var s = tokens[pos].value;

        if (tokens[pos++].type === TokenType.DecimalNumber &&
                pos < tokens.length &&
                tokens[pos].type === TokenType.Identifier
        ) s += tokens[pos++].value;

        return s;
    }

    function checkExcluding(exclude, _i) {
        var start = _i;

        while(_i < tokens.length) {
            if (exclude[tokens[_i++].type]) break;
        }

        return _i - start - 2;
    }

    function joinValues(start, finish) {
        var s = '';

        for (var i = start; i < finish + 1; i++) {
            s += tokens[i].value;
        }

        return s;
    }

    function joinValues2(start, num) {
        if (start + num - 1 >= tokens.length) return;

        var s = '';

        for (var i = 0; i < num; i++) {
            s += tokens[start + i].value;
        }

        return s;
    }

    function markSC() {
        var ws = -1, // whitespaces
            sc = -1, // ws and comments
            t;

        for (var i = 0; i < tokens.length; i++) {
            t = tokens[i];
            switch (t.type) {
                case TokenType.Space:
                case TokenType.Tab:
                case TokenType.Newline:
                    t.ws = true;
                    t.sc = true;

                    if (ws === -1) ws = i;
                    if (sc === -1) sc = i;

                    break;
                case TokenType.CommentML:
                    if (ws !== -1) {
                        tokens[ws].ws_last = i - 1;
                        ws = -1;
                    }

                    t.sc = true;

                    break;
                default:
                    if (ws !== -1) {
                        tokens[ws].ws_last = i - 1;
                        ws = -1;
                    }

                    if (sc !== -1) {
                        tokens[sc].sc_last = i - 1;
                        sc = -1;
                    }
            }
        }

        if (ws !== -1) tokens[ws].ws_last = i - 1;
        if (sc !== -1) tokens[sc].sc_last = i - 1;
    }

    return _getAST(_tokens, rule, _needInfo);
}

    return getCSSPAST(getTokens(s), rule, _needInfo);
}
var translator = new CSSOTranslator(),
    cleanInfo = $util.cleanInfo;
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
function CSSOTranslator() {}

CSSOTranslator.prototype.translate = function(tree) {
//    console.trace('--------');
//    console.log(tree);
    return this._t(tree);
};

CSSOTranslator.prototype._m_simple = {
    'unary': 1, 'nth': 1, 'combinator': 1, 'ident': 1, 'number': 1, 's': 1,
    'string': 1, 'attrselector': 1, 'operator': 1, 'raw': 1, 'unknown': 1
};

CSSOTranslator.prototype._m_composite = {
    'simpleselector': 1, 'dimension': 1, 'selector': 1, 'property': 1, 'value': 1,
    'filterv': 1, 'progid': 1, 'ruleset': 1, 'atruleb': 1, 'atrulerq': 1, 'atrulers': 1,
    'stylesheet': 1
};

CSSOTranslator.prototype._m_primitive = {
    'cdo': 'cdo', 'cdc': 'cdc', 'decldelim': ';', 'namespace': '|', 'delim': ','
};

CSSOTranslator.prototype._t = function(tree) {
    var t = tree[0];
    if (t in this._m_primitive) return this._m_primitive[t];
    else if (t in this._m_simple) return this._simple(tree);
    else if (t in this._m_composite) return this._composite(tree);
    return this[t](tree);
};

CSSOTranslator.prototype._composite = function(t, i) {
    var s = '';
    i = i === undefined ? 1 : i;
    for (; i < t.length; i++) s += this._t(t[i]);
    return s;
};

CSSOTranslator.prototype._simple = function(t) {
    return t[1];
};

CSSOTranslator.prototype.percentage = function(t) {
    return this._t(t[1]) + '%';
};

CSSOTranslator.prototype.comment = function(t) {
    return '/*' + t[1] + '*/';
};

CSSOTranslator.prototype.clazz = function(t) {
    return '.' + this._t(t[1]);
};

CSSOTranslator.prototype.atkeyword = function(t) {
    return '@' + this._t(t[1]);
};

CSSOTranslator.prototype.shash = function(t) {
    return '#' + t[1];
};

CSSOTranslator.prototype.vhash = function(t) {
    return '#' + t[1];
};

CSSOTranslator.prototype.attrib = function(t) {
    return '[' + this._composite(t) + ']';
};

CSSOTranslator.prototype.important = function(t) {
    return '!' + this._composite(t) + 'important';
};

CSSOTranslator.prototype.nthselector = function(t) {
    return ':' + this._simple(t[1]) + '(' + this._composite(t, 2) + ')';
};

CSSOTranslator.prototype.funktion = function(t) {
    return this._simple(t[1]) + '(' + this._composite(t[2]) + ')';
};

CSSOTranslator.prototype.declaration = function(t) {
    return this._t(t[1]) + ':' + this._t(t[2]);
};

CSSOTranslator.prototype.filter = function(t) {
    return this._t(t[1]) + ':' + this._t(t[2]);
};

CSSOTranslator.prototype.block = function(t) {
    return '{' + this._composite(t) + '}';
};

CSSOTranslator.prototype.braces = function(t) {
    return t[1] + this._composite(t, 3) + t[2];
};

CSSOTranslator.prototype.atrules = function(t) {
    return this._composite(t) + ';';
};

CSSOTranslator.prototype.atruler = function(t) {
    return this._t(t[1]) + this._t(t[2]) + '{' + this._t(t[3]) + '}';
};

CSSOTranslator.prototype.pseudoe = function(t) {
    return '::' + this._t(t[1]);
};

CSSOTranslator.prototype.pseudoc = function(t) {
    return ':' + this._t(t[1]);
};

CSSOTranslator.prototype.uri = function(t) {
    return 'url(' + this._composite(t) + ')';
};

CSSOTranslator.prototype.functionExpression = function(t) {
    return 'expression(' + t[1] + ')';
};
