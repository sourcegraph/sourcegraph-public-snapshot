/*!
 * node-sass: lib/extensions.js
 */

var flags = {},
    fs = require('fs'),
    pkg = require('../package.json'),
    path = require('path');

/**
 * Collect Arguments
 *
 * @param {Array} args
 * @api private
 */

function collectArguments(args) {
  for (var i = 0; i < args.length; i += 2) {
    if (args[i].lastIndexOf('--', 0) !== 0) {
      --i;
      continue;
    }

    flags[args[i]] = args[i + 1];
  }
}

/**
 * Get Runtime Info
 *
 * @api private
 */

function getRuntimeInfo() {
  var execPath = fs.realpathSync(process.execPath); // resolve symbolic link

  var runtime = execPath
               .split(/[\\/]+/).pop()
               .split('.').shift();

  runtime = runtime === 'nodejs' ? 'node' : runtime;

  return {
    name: runtime,
    execPath: execPath
  };
}

/**
 * Get binary name.
 * If environment variable SASS_BINARY_NAME or
 * process argument --binary-name is provided,
 * return it as is, otherwise make default binary
 * name: {platform}-{arch}-{v8 version}.node
 *
 * @api private
 */

function getBinaryName() {
  var binaryName;

  if (flags['--sass-binary-name']) {
    binaryName = flags['--sass-binary-name'];
  } else if (process.env.SASS_BINARY_NAME) {
    binaryName = process.env.SASS_BINARY_NAME;
  } else if (pkg.nodeSassConfig && pkg.nodeSassConfig.binaryName) {
    binaryName = pkg.nodeSassConfig.binaryName;
  } else {
    binaryName = [process.platform, '-',
                  process.arch, '-',
                  process.versions.modules].join('');
  }

  return [binaryName, 'binding.node'].join('_');
}

/**
 * Determine the URL to fetch binary file from.
 * By default feth from the node-sass distribution
 * site on GitHub.
 *
 * The default URL can be overriden using
 * the environment variable SASS_BINARY_SITE
 * or a command line option --sass-binary-site:
 *
 *   node scripts/install.js --sass-binary-site http://example.com/
 *
 * The URL should to the mirror of the repository
 * laid out as follows:
 *
 * SASS_BINARY_SITE/
 *
 *  v3.0.0
 *  v3.0.0/freebsd-x64-14_binding.node
 *  ....
 *  v3.0.0
 *  v3.0.0/freebsd-ia32-11_binding.node
 *  v3.0.0/freebsd-x64-42_binding.node
 *  ... etc. for all supported versions and platforms
 *
 * @api private
 */

function getBinaryUrl() {
  var site = flags['--sass-binary-site'] ||
             process.env.SASS_BINARY_SITE  ||
             pkg.nodeSassConfig.binarySite;
	return [site, 'v' + pkg.version, sass.binaryName].join('/');
}


collectArguments(process.argv.slice(2));

var sass = process.sass = {};

sass.binaryName = getBinaryName();
sass.binaryUrl = getBinaryUrl();
sass.runtime = getRuntimeInfo();

/**
 * Get binary path.
 * If environment variable SASS_BINARY_PATH or
 * process argument --sass-binary-path is provided,
 * select it by appending binary name, otherwise
 * make default binary path using binary name.
 * Once the primary selection is made, check if
 * callers wants to throw if file not exists before
 * returning.
 *
 * @param {Boolean} throwIfNotExists
 * @api private
 */

sass.getBinaryPath = function(throwIfNotExists) {
  var binaryPath;

  if (flags['--sass-binary-path']) {
    binaryPath = flags['--sass-binary-path'];
  } else if (process.env.SASS_BINARY_PATH) {
    binaryPath = process.env.SASS_BINARY_PATH;
  } else if (pkg.nodeSassConfig && pkg.nodeSassConfig.binaryPath) {
    binaryPath = pkg.nodeSassConfig.binaryPath;
  } else {
    binaryPath = path.join(__dirname, '..', 'vendor', sass.binaryName.replace(/_/, '/'));
  }

  if (!fs.existsSync(binaryPath) && throwIfNotExists) {
    throw new Error(['`libsass` bindings not found in ', binaryPath, '. Try reinstalling `node-sass`?'].join(''));
  }

  return binaryPath;
};

sass.binaryPath = sass.getBinaryPath();
