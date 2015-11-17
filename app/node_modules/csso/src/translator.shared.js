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
