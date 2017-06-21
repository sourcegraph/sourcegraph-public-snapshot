#!/usr/local/bin/node

// Might be good to use an explicit path to node on the shebang line in case
// it isn't in PATH when launched by Chrome.

// Adapted from https://github.com/jdiamond/chrome-native-messaging.

var fs = require('fs');
var exec = require('child_process').exec;

var nativeMessage = require('./index');

var input = new nativeMessage.Input();
var transform = new nativeMessage.Transform(messageHandler);
var output = new nativeMessage.Output();

process.stdin
    .pipe(input)
    .pipe(transform)
    .pipe(output)
    .pipe(process.stdout)
    ;

var subscriptions = {};

var timer = setInterval(function () {
    if (subscriptions.time) {
        output.write({ time: new Date().toISOString() });
    }
}, 1000);

input.on('end', function () {
    clearInterval(timer);
});

function messageHandler(msg, push, done) {
    if (msg.cmd) {
        exec(msg.cmd, function (error, stdout, stderr) {
            push(error);
            done();
        });
    }
}
