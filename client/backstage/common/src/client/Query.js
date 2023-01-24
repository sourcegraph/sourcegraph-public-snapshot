"use strict";
var __makeTemplateObject = (this && this.__makeTemplateObject) || function (cooked, raw) {
    if (Object.defineProperty) { Object.defineProperty(cooked, "raw", { value: raw }); } else { cooked.raw = raw; }
    return cooked;
};
exports.__esModule = true;
exports.AuthenticatedUserQuery = exports.UserQuery = exports.SearchQuery = void 0;
var graphql_request_1 = require("graphql-request");
var auth_1 = require("@sourcegraph/shared/src/auth");
var SearchQuery = /** @class */ (function () {
    function SearchQuery(query) {
        this.query = query;
    }
    SearchQuery.prototype.Marshal = function (data) {
        var results = new Array();
        for (var v in data.search.results.results) {
            var _a = v, repository = _a.repository, fileContent = _a.file.fileContent;
            results.push({ repository: repository, fileContent: fileContent });
        }
        return results;
    };
    SearchQuery.prototype.vars = function () {
        return { search: this.query };
    };
    SearchQuery.prototype.gql = function () {
        return (0, graphql_request_1.gql)(templateObject_1 || (templateObject_1 = __makeTemplateObject(["\n      query ($search: String!) {\n        search(query: $search) {\n          results {\n            __typename\n            ... on FileMatch {\n              repository\n            }\n            file {\n              content\n            }\n          }\n        }\n      }\n    "], ["\n      query ($search: String!) {\n        search(query: $search) {\n          results {\n            __typename\n            ... on FileMatch {\n              repository\n            }\n            file {\n              content\n            }\n          }\n        }\n      }\n    "])));
    };
    return SearchQuery;
}());
exports.SearchQuery = SearchQuery;
var UserQuery = /** @class */ (function () {
    function UserQuery() {
    }
    UserQuery.prototype.Marshal = function (data) {
        if ("currentUser" in data) {
            return [data.currentUser.username];
        }
        throw new Error("username not found");
    };
    UserQuery.prototype.vars = function () {
        return "";
    };
    UserQuery.prototype.gql = function () {
        return (0, graphql_request_1.gql)(templateObject_2 || (templateObject_2 = __makeTemplateObject(["\n    query {\n      currentUser {\n        username\n      }\n    }\n    "], ["\n    query {\n      currentUser {\n        username\n      }\n    }\n    "])));
    };
    return UserQuery;
}());
exports.UserQuery = UserQuery;
var AuthenticatedUserQuery = /** @class */ (function () {
    function AuthenticatedUserQuery() {
    }
    AuthenticatedUserQuery.prototype.gql = function () {
        return auth_1.currentAuthStateQuery;
    };
    AuthenticatedUserQuery.prototype.vars = function () {
        return "";
    };
    AuthenticatedUserQuery.prototype.Marshal = function (data) {
        return [data.currentUser];
    };
    return AuthenticatedUserQuery;
}());
exports.AuthenticatedUserQuery = AuthenticatedUserQuery;
var templateObject_1, templateObject_2;
