"use strict";
var bundle = require('./bin-bundle');
var init = require('./bin-init');
var uninstall = require('./bin-uninstall');
var install = require('./bin-install');
var list = require('./bin-list');
var search = require('./bin-search');
var open = require('./bin-open');
var view = require('./bin-view');
var prune = require('./bin-prune');
exports.aliases = {
    i: install,
    in: install,
    install: install,
    r: uninstall,
    rm: uninstall,
    un: uninstall,
    remove: uninstall,
    uninstall: uninstall,
    init: init,
    ls: list,
    ll: list,
    la: list,
    list: list,
    bundle: bundle,
    search: search,
    open: open,
    view: view,
    info: view,
    prune: prune
};
//# sourceMappingURL=aliases.js.map