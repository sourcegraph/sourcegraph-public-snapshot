'use strict';

var spawn = require('child_process').spawn;
var executable = require('executable');
var spawnSync = require('spawn-sync');

module.exports = function (bin, cmd, cb) {
	if (typeof cmd === 'function') {
		cb = cmd;
		cmd = ['--help'];
	}

	executable(bin, function (err, works) {
		if (err) {
			cb(err);
			return;
		}

		if (!works) {
			cb(new Error('Couldn\'t execute the `' + bin + '` binary. Make sure it has the right permissions.'));
			return;
		}

		var cp = spawn(bin, cmd);

		cp.on('error', cb);
		cp.on('exit', function (code) {
			cb(null, code === 0);
		});
	});
};

module.exports.sync = function (bin, cmd) {
	cmd = cmd || ['--help'];

	if (!executable.sync(bin)) {
		return false;
	}

	return spawnSync(bin, cmd).status === 0;
};
