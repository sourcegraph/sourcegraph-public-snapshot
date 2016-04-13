'use strict';
var ip = require('ip-regex').v4().source;
var tlds = require('./tlds.json').join('|');

module.exports = function (opts) {
	opts = opts || {};

	var auth = '(?:\\S+(?::\\S*)?@)?';
	var domain = '(?:\\.(?:xn--[a-z0-9\\-]{1,59}|(?:[a-z\\u00a1-\\uffff0-9]+-?){0,62}[a-z\\u00a1-\\uffff0-9]{1,63}))*';
	var host = '(?:xn--[a-z0-9\\-]{1,59}|(?:(?:[a-z\\u00a1-\\uffff0-9]+-?){0,62}[a-z\\u00a1-\\uffff0-9]{1,63}))';
	var path = '(?:\/[^\\s]*)?';
	var port = '(?::\\d{2,5})?';
	var protocol = '(?:(?:(?:\\w)+:)?\/\/)?';
	var tld = '(?:\\.(?:xn--[a-z0-9\\-]{1,59}|' + tlds + '+))';

	var regex = [
		protocol + auth + '(?:' + ip + '|',
		'(?:localhost)|' + host + domain + tld + ')' + port + path
	].join('');

	return opts.exact ? new RegExp('(?:^' + regex + '$)', 'i') :
						new RegExp('(["\'])?' + regex + '\\1', 'ig');
};
