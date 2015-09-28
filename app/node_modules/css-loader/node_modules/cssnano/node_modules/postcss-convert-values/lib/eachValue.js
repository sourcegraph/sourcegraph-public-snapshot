'use strict';

var balanced = require('balanced-match');
var list = require('css-list');

module.exports = function eachValue (value, callback) {
    return list.map(value, function (value, type) {
        var name,
            match,
            index;

        if (type === null) {
            return callback(value);
        }

        if (type === 'func') {
            index = value.indexOf('(');
            name = value.substring(0, index);
            if (~name.indexOf('calc')) {
                match = balanced('(', ')', value);
                if (match) {
                    return name + '(' + eachValue(match.body, callback) + ')';
                }
            }
        }
    });
};
