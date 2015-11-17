var fs = require('fs'),
    csso = require('./cssoapi.js'),
    src;

var args = process.argv.slice(2),
    opts = args.length ? getOpts(args, [
            '-v', '--version',
            '-h', '--help',
            '-dp', '--parser',
            '-off', '--restructure-off'
           ], [
            '-r', '--rule',
            '-i', '--input',
            '-o', '--output'
           ]) : null,
    single = opts && opts.single,
    pairs = opts && opts.pairs,
    other = opts && opts.other,
    ro = single && single.contains(['-off', '--restructure-off']),
    inFile = (pairs && (pairs['-i'] || pairs['--input'])) || (other && other[0]),
    outFile = (pairs && (pairs['-o'] || pairs['--output'])) || (other && other[1]),
    rule = pairs && (pairs['-r'] || pairs['--rule']) || 'stylesheet';

if (single && single.contains(['-v', '--version'])) {
    console.log(require(__dirname + '/../package.json').version);
} else if (!inFile || !opts || (single && single.contains(['-h', '--help'])) || other.length > 2) {
    console.log(fs.readFileSync(__dirname + '/../USAGE', 'utf-8'));
} else {
    src = fs.readFileSync(inFile).toString().trim();

    if (single.contains(['-dp', '--parser'])) csso.printTree(csso.cleanInfo(csso.parse(src, rule, true)));
    else {
        if (!outFile) console.log(csso.justDoIt(src, ro, true));
        else fs.writeFileSync(outFile, csso.justDoIt(src, ro, true));
    }
}

// Utils

function getOpts(argv, o_single, o_pairs) {
    var opts = { single : [], pairs : {}, other : [] },
        arg,
        i = 0;

    for (; i < argv.length;) {
        arg = argv[i];
        if (o_single && o_single.indexOf(arg) !== -1 && (!o_pairs || o_pairs.indexOf(arg) === -1)) {
            opts.single.push(arg);
        } else if (o_pairs && o_pairs.indexOf(arg) !== -1 && (!o_single || o_single.indexOf(arg) === -1)) {
            opts.pairs[arg] = argv[++i];
        } else opts.other.push(arg);
        i++;
    }

    opts.single.contains = function(value) {
        if (typeof value === 'string') {
            return this.indexOf(value) !== -1;
        } else {
            for (var i = 0; i < value.length; i++) if (this.indexOf(value[i]) !== -1) return true;
        }
        return false;
    };

    return opts;
}
