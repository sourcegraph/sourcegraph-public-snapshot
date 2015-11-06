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
