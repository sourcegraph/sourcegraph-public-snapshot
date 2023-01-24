"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __generator = (this && this.__generator) || function (thisArg, body) {
    var _ = { label: 0, sent: function() { if (t[0] & 1) throw t[1]; return t[1]; }, trys: [], ops: [] }, f, y, t, g;
    return g = { next: verb(0), "throw": verb(1), "return": verb(2) }, typeof Symbol === "function" && (g[Symbol.iterator] = function() { return this; }), g;
    function verb(n) { return function (v) { return step([n, v]); }; }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while (g && (g = 0, op[0] && (_ = 0)), _) try {
            if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [op[0] & 2, t.value];
            switch (op[0]) {
                case 0: case 1: t = op; break;
                case 4: _.label++; return { value: op[1], done: false };
                case 5: _.label++; y = op[1]; op = [0]; continue;
                case 7: op = _.ops.pop(); _.trys.pop(); continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) { _ = 0; continue; }
                    if (op[0] === 3 && (!t || (op[1] > t[0] && op[1] < t[3]))) { _.label = op[1]; break; }
                    if (op[0] === 6 && _.label < t[1]) { _.label = t[1]; t = op; break; }
                    if (t && _.label < t[2]) { _.label = t[2]; _.ops.push(op); break; }
                    if (t[2]) _.ops.pop();
                    _.trys.pop(); continue;
            }
            op = body.call(thisArg, _);
        } catch (e) { op = [6, e]; y = 0; } finally { f = t = 0; }
        if (op[0] & 5) throw op[1]; return { value: op[0] ? op[1] : void 0, done: true };
    }
};
exports.__esModule = true;
exports.createService = void 0;
var graphql_request_1 = require("graphql-request");
var Query_1 = require("./Query");
var createService = function (config) {
    var endpoint = config.endpoint, token = config.token, sudoUsername = config.sudoUsername;
    var base = new BaseClient(endpoint, token, sudoUsername || "");
    return new SourcegraphClient(base);
};
exports.createService = createService;
var SourcegraphClient = /** @class */ (function () {
    function SourcegraphClient(client) {
        this.Users = this;
        this.Search = this;
        this.client = client;
    }
    SourcegraphClient.prototype.SearchQuery = function (query) {
        return __awaiter(this, void 0, void 0, function () {
            var q, data;
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0:
                        q = new Query_1.SearchQuery(query);
                        return [4 /*yield*/, this.client.fetch(q)];
                    case 1:
                        data = _a.sent();
                        return [2 /*return*/, q.Marshal(data)];
                }
            });
        });
    };
    SourcegraphClient.prototype.CurrentUsername = function () {
        return __awaiter(this, void 0, void 0, function () {
            var q, data;
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0:
                        q = new Query_1.UserQuery();
                        return [4 /*yield*/, this.client.fetch(q)];
                    case 1:
                        data = _a.sent();
                        return [2 /*return*/, data[0]];
                }
            });
        });
    };
    SourcegraphClient.prototype.GetAuthenticatedUser = function () {
        return __awaiter(this, void 0, void 0, function () {
            var q, data;
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0:
                        q = new Query_1.AuthenticatedUserQuery();
                        return [4 /*yield*/, this.client.fetch(q)];
                    case 1:
                        data = _a.sent();
                        return [2 /*return*/, data[0]];
                }
            });
        });
    };
    return SourcegraphClient;
}());
var BaseClient = /** @class */ (function () {
    function BaseClient(baseUrl, token, sudoUsername) {
        var authz = (sudoUsername === null || sudoUsername === void 0 ? void 0 : sudoUsername.length) > 0 ? "token-sudo user=\"".concat(sudoUsername, "\",token=\"").concat(token, "\"") : "token ".concat(token);
        var apiUrl = "".concat(baseUrl, "/.api/graphql");
        this.client = new graphql_request_1.GraphQLClient(apiUrl, {
            headers: {
                'X-Requested-With': "Sourcegraph - Backstage plugin DEV",
                Authorization: authz
            }
        });
    }
    BaseClient.prototype.fetch = function (q) {
        return __awaiter(this, void 0, void 0, function () {
            var data;
            return __generator(this, function (_a) {
                switch (_a.label) {
                    case 0: return [4 /*yield*/, this.client.request(q.gql(), q.vars())];
                    case 1:
                        data = _a.sent();
                        return [2 /*return*/, q.Marshal(data)];
                }
            });
        });
    };
    return BaseClient;
}());
