"use strict";
var url_1 = require('url');
var HttpProxyAgent = require('http-proxy-agent');
var HttpsProxyAgent = require('https-proxy-agent');
function proxy(options) {
    var noProxy = options.noProxy || process.env.NO_PROXY || process.env.no_proxy;
    if (noProxy === '*') {
        return function (urlStr) { return undefined; };
    }
    var noProxyList = parseNoProxy(noProxy);
    var httpProxy = options.httpProxy || options.proxy || process.env.HTTP_PROXY || process.env.http_proxy;
    var httpsProxy = options.httpsProxy || options.proxy || process.env.HTTPS_PROXY || process.env.https_proxy;
    var httpProxyUrl = httpProxy ? url_1.parse(httpProxy) : undefined;
    var httpsProxyUrl = httpsProxy ? url_1.parse(httpsProxy) : undefined;
    return function (urlStr) {
        var url = url_1.parse(urlStr);
        var ignore = noProxy && urlInNoProxy(url, noProxyList);
        if (url.protocol === 'https:') {
            if (ignore || !httpsProxyUrl) {
                return;
            }
            return new HttpsProxyAgent(httpsProxyUrl);
        }
        if (ignore || !httpProxyUrl) {
            return;
        }
        return new HttpProxyAgent(httpProxyUrl);
    };
}
function formatHostname(hostname) {
    return hostname.replace(/^\.*/, '.').toLowerCase();
}
function parseNoProxy(noProxy) {
    if (!noProxy) {
        return [];
    }
    return noProxy.split(',').map(function (zone) {
        var location = zone.trim().toLowerCase();
        var parts = location.split(':', 2);
        var hostname = formatHostname(parts[0]);
        var port = parts[1];
        return { hostname: hostname, port: port };
    });
}
function urlInNoProxy(url, noProxyList) {
    var hostname = formatHostname(url.hostname);
    var port = url.port || (url.protocol === 'https:' ? '443' : '80');
    return noProxyList.some(function (noProxy) {
        var isMatchedAt = hostname.indexOf(noProxy.hostname);
        var hostnameMatched = isMatchedAt > -1 && isMatchedAt === hostname.length - noProxy.hostname.length;
        if (noProxy.port != null) {
            return hostnameMatched && port === noProxy.port;
        }
        return hostnameMatched;
    });
}
module.exports = proxy;
//# sourceMappingURL=index.js.map