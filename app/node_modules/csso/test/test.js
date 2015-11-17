var fs = require('fs');
var path = require('path');
var assert = require('assert');
var csso = require('../lib/cssoapi.js');
var utils = require('../lib/util.js');

var funcs = {
    'p': function parse(src, match) {
        return csso.treeToString(clean(src, match));
    },
    'l': function translate(src, match) {
        return csso.translate(clean(src, match));
    },
    'cl': function translate(src, match) {
        return csso.translate(clean(src, match, true));
    }
};

function parse(src, match) {
    return csso.parse(src, match, true);
}

function clean(src, match, compress) {
    var out = parse(src, match);
    if (compress) {
        out = csso.compress(out);
    }

    return utils.cleanInfo(out);
}

function readFile(path) {
    return fs.readFileSync(path).toString().trim();
}

function runTest(files, filePrefix, rule) {
    var src = readFile(filePrefix + '.css');
    var basename = path.relative(__dirname, filePrefix);

    var test = function(fn) {
        it('in ' + basename, function() {
            assert.equal(funcs[fn](src, rule), readFile(filePrefix + '.' + fn));
        });
    }
    for (var fn in funcs) {
        if (fn in files) {
            test(fn);
        }
    }
}

function runTestsInDir(dir, rule) {
    var files = {};
    fs.readdirSync(dir).forEach(function(f) {
        var ext = path.extname(f);
        if (ext) {
            var basename = path.basename(f, ext);
            if (!files[basename]) {
                files[basename] = {};
            }

            files[basename][ext.substring(1)] = 1;
        }
    });

    for (var k in files) {
        runTest(files[k], path.join(dir, k), rule);
    }
}

describe('CSSO', function() {
    var testDir = path.join(__dirname, 'data');
    fs.readdirSync(testDir).forEach(function(ruleDir) {
        var dir = path.join(testDir, ruleDir);
        var stat = fs.statSync(dir);
        if (!stat.isDirectory()) {
            return;
        }

        runTestsInDir(dir, ruleDir.substring(5));
    });
});
