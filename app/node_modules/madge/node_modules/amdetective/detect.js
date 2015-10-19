var fs = require('fs'),
    amdetective = require('./index.js');

console.log('Reading file from first argument: ' + process.argv[2]);
console.log(amdetective(fs.readFileSync(process.argv[2]).toString()));


