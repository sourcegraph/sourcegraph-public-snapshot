"use strict";
var __assign = (this && this.__assign) || function () {
    __assign = Object.assign || function(t) {
        for (var s, i = 1, n = arguments.length; i < n; i++) {
            s = arguments[i];
            for (var p in s) if (Object.prototype.hasOwnProperty.call(s, p))
                t[p] = s[p];
        }
        return t;
    };
    return __assign.apply(this, arguments);
};
exports.__esModule = true;
exports.parseCatalog = void 0;
var catalog_model_1 = require("@backstage/catalog-model");
var plugin_catalog_backend_1 = require("@backstage/plugin-catalog-backend");
var parseCatalog = function (src, providerName) {
    var results = [];
    src.forEach(function (r) {
        var _a;
        var location = {
            "type": "url",
            "target": "".concat(r.repository, "/catalog-info.yaml")
        };
        var yaml = Buffer.from(r.fileContent, "utf8");
        for (var _i = 0, _b = (0, plugin_catalog_backend_1.parseEntityYaml)(yaml, location); _i < _b.length; _i++) {
            var item = _b[_i];
            var parsed = item;
            results.push({
                entity: __assign(__assign({}, parsed.entity), { metadata: __assign(__assign({}, parsed.entity.metadata), { annotations: __assign(__assign({}, parsed.entity.metadata.annotations), (_a = {}, _a[catalog_model_1.ANNOTATION_LOCATION] = "url:".concat(parsed.location.target), _a[catalog_model_1.ANNOTATION_ORIGIN_LOCATION] = providerName, _a)) }) }),
                locationKey: parsed.location.target
            });
        }
    });
    return results;
};
exports.parseCatalog = parseCatalog;
