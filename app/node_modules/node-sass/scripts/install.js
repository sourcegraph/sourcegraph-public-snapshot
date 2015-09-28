/*!
 * node-sass: scripts/install.js
 */

var fs = require('fs'),
    mkdir = require('mkdirp'),
    npmconf = require('npmconf'),
    path = require('path'),
    request = require('request'),
    pkg = require('../package.json');

require('../lib/extensions');

/**
 * Download file, if succeeds save, if not delete
 *
 * @param {String} url
 * @param {String} dest
 * @param {Function} cb
 * @api private
 */

function download(url, dest, cb) {
  var reportError = function(err) {
    cb(['Cannot download "', url, '": ',
      typeof err.message === 'string' ? err.message : err].join(''));
  };
  var successful = function(response) {
    return response.statusCode >= 200 && response.statusCode < 300;
  };

  applyProxy({ rejectUnauthorized: false }, function(options) {
    options.headers = {
      'User-Agent': [
        'node/', process.version, ' ',
        'node-sass-installer/', pkg.version
      ].join('')
    };
    try {
      request(url, options, function(err, response) {
        if (err) {
          reportError(err);
        } else if (!successful(response)) {
            reportError(['HTTP error', response.statusCode, response.statusMessage].join(' '));
        } else {
            cb();
        }
      }).on('response', function(response) {
          if (successful(response)) {
            response.pipe(fs.createWriteStream(dest));
          }
      });
    } catch (err) {
      cb(err);
    }
  });
}

/**
 * Get applyProxy settings
 *
 * @param {Object} options
 * @param {Function} cb
 * @api private
 */

function applyProxy(options, cb) {
  npmconf.load({}, function (er, conf) {
    var proxyUrl;

    if (!er) {
      proxyUrl = conf.get('https-proxy') ||
                 conf.get('proxy') ||
                 conf.get('http-proxy');
    }

    var env = process.env;

    options.proxy = proxyUrl ||
                    env.HTTPS_PROXY ||
                    env.https_proxy ||
                    env.HTTP_PROXY ||
                    env.http_proxy;

    cb(options);
  });
}

/**
 * Check and download binary
 *
 * @api private
 */

function checkAndDownloadBinary() {
  try {
    process.sass.getBinaryPath(true);
    return;
  } catch (e) { }

  mkdir(path.dirname(process.sass.binaryPath), function(err) {
    if (err) {
      console.error(err);
      return;
    }

    download(process.sass.binaryUrl, process.sass.binaryPath, function(err) {
      if (err) {
        console.error(err);
        return;
      }

      console.log('Binary downloaded and installed at', process.sass.binaryPath);
    });
  });
}

/**
 * Skip if CI
 */

if (process.env.SKIP_SASS_BINARY_DOWNLOAD_FOR_CI) {
  console.log('Skipping downloading binaries on CI builds');
  return;
}

/**
 * If binary does not exsit, download it
 */

checkAndDownloadBinary();
