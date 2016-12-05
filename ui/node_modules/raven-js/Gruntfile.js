'use strict';

module.exports = function(grunt) {
    var _ = require('lodash');
    var path = require('path');
    var os = require('os');
    var through = require('through2');
    var proxyquire = require('proxyquireify');
    var versionify = require('browserify-versionify');
    var derequire = require('derequire/plugin');
    var collapser = require('bundle-collapser/plugin');

    var excludedPlugins = [
        'react-native'
    ];

    var plugins = grunt.option('plugins');
    // Create plugin paths and verify they exist
    plugins = _.map(plugins ? plugins.split(',') : [], function (plugin) {
        var p = 'plugins/' + plugin + '.js';

        if(!grunt.file.exists(p))
            throw new Error("Plugin '" + plugin + "' not found in plugins directory.");

        return p;
    });

    // custom browserify transformer to re-write plugins to
    // self-register with Raven via addPlugin
    function AddPluginBrowserifyTransformer() {
        return function (file) {
            return through(function (buf, enc, next) {
                buf = buf.toString('utf8');
                if (/plugins/.test(file)) {
                    buf += "\nrequire('../src/singleton').addPlugin(module.exports);";
                }
                this.push(buf);
                next();
            });
        };
    }

    // Taken from http://dzone.com/snippets/calculate-all-combinations
    var combine = function (a) {
        var fn = function (n, src, got, all) {
            if (n === 0) {
                all.push(got);
                return;
            }

            for (var j = 0; j < src.length; j++) {
                fn(n - 1, src.slice(j + 1), got.concat([src[j]]), all);
            }
        };

        var excluded = _.map(excludedPlugins, function(plugin) {
            return 'plugins/' + plugin + '.js';
        });

        // Remove the plugins that we don't want to build
        a = _.filter(a, function(n) {
            return excluded.indexOf(n) === -1;
        });

        var all = [a];

        for (var i = 0; i < a.length; i++) {
            fn(i, a, [], all);
        }

        return all;
    };

    var plugins = grunt.file.expand('plugins/*.js');

    var cleanedPlugins = plugins.filter(function (plugin) {
        var pluginName = path.basename(plugin, '.js');

        return excludedPlugins.indexOf(pluginName) === -1;
    });

    var pluginSingleFiles = cleanedPlugins.map(function (plugin) {
        var filename = path.basename(plugin);

        var file = {};
        file.src = plugin;
        file.dest = path.join('build', 'plugins', filename);

        return file;
    });

    var pluginCombinations = combine(plugins);
    var pluginConcatFiles = _.reduce(pluginCombinations, function (dict, comb) {
        var key = _.map(comb, function (plugin) {
            return path.basename(plugin, '.js');
        });
        key.sort();

        var dest = path.join('build/', key.join(','), '/raven.js');
        dict[dest] = ['src/singleton.js'].concat(comb);

        return dict;
    }, {});

    var browserifyConfig = {
        options: {
            banner: grunt.file.read('template/_copyright.js'),
                browserifyOptions: {
                standalone: 'Raven' // umd
            },
            transform: [versionify],
            plugin: [
                derequire,
                collapser
            ]
        },
        core: {
            src: 'src/singleton.js',
            dest: 'build/raven.js'
        },
        'plugins-combined': {
            files: pluginConcatFiles,
            options: {
                transform: [
                    [versionify],
                    [new AddPluginBrowserifyTransformer()]
                ]
            }
        },
        test: {
            src: 'test/**/*.test.js',
            dest: 'build/raven.test.js',
            options: {
                browserifyOptions: {
                    debug: false// source maps
                },
                ignore: ['react-native'],
                plugin: [proxyquire.plugin]
            }
        }
    };

    // Create a dedicated entry in browserify config for
    // each individual plugin (each needs a unique `standalone`
    // config)
    var browserifyPluginTaskNames = [];
    pluginSingleFiles.forEach(function (item) {
        var name = item.src
            .replace(/.*\//, '') // everything before slash
            .replace('.js', ''); // extension
        var capsName = name.charAt(0).toUpperCase() + name.slice(1);
        var config = {
            src: item.src,
            dest: item.dest,
            options: {
                browserifyOptions: {
                    // e.g. Raven.Plugins.Angular
                    standalone: 'Raven.Plugins.' + capsName
                }
            }
        };
        browserifyConfig[name] = config;
        browserifyPluginTaskNames.push('browserify:' + name);
    });


    var awsConfigPath = path.join(os.homedir(), '.aws', 'raven-js.json');
    var gruntConfig = {
        pkg: grunt.file.readJSON('package.json'),
        aws: grunt.file.exists(awsConfigPath) ? grunt.file.readJSON(awsConfigPath): {},

        clean: ['build'],

        browserify: browserifyConfig,

        uglify: {
            options: {
                sourceMap: true,

                // Only preserve comments that start with (!)
                preserveComments: /^!/,

                // Minify object properties that begin with _ ("private"
                // methods and values)
                mangleProperties: {
                    regex: /^_/
                },

                compress: {
                    booleans: true,
                    conditionals: true,
                    dead_code: true,
                    join_vars: true,
                    pure_getters: true,
                    sequences: true,
                    unused: true,

                    global_defs: {
                        '__DEV__': false
                    }
                }
            },
            dist: {
                src: ['build/**/*.js'],
                ext: '.min.js',
                expand: true
            }
        },

        eslint: {
            target: ['Gruntfile.js', 'src/**/*.js', 'plugins/**/*.js']
        },

        mocha: {
            options: {
                mocha: {
                    ignoreLeaks: true,
                    grep: grunt.option('grep')
                },
                log: true,
                reporter: 'Dot',
                run: true
            },
            unit: {
                src: ['test/index.html'],
                nonull: true
            },
            integration: {
                src: ['test/integration/index.html'],
                nonull: true
            }
        },

        release: {
            options: {
                npm: false,
                commitMessage: 'Release <%= version %>'
            }
        },

        s3: {
            options: {
                key: '<%= aws.key %>',
                secret: '<%= aws.secret %>',
                bucket: '<%= aws.bucket %>',
                access: 'public-read',
                // Limit concurrency
                maxOperations: 20,
                headers: {
                    // Surrogate-Key header for Fastly to purge by release
                    'x-amz-meta-surrogate-key': '<%= pkg.release %>'
                }
            },
            all: {
                upload: [{
                    src: 'build/**/*',
                    dest: '<%= pkg.release %>/',
                    rel: 'build/'
                }]
            }
        },

        connect: {
            test: {
                options: {
                    port: 8000,
                    debug: true,
                    keepalive: true
                }
            },

            docs: {
                options: {
                    port: 8000,
                    debug: true,
                    base: 'docs/_build/html',
                    keepalive: true
                }
            }
        },

        copy: {
            dist: {
                expand: true,
                flatten: false,
                cwd: 'build/',
                src: '**',
                dest: 'dist/'
            }
        },

        sri: {
            dist: {
                src: ['dist/*.js'],
                options: {
                    dest: 'dist/sri.json',
                    pretty: true
                }
            },
            build: {
                src: ['build/**/*.js'],
                options: {
                    dest: 'build/sri.json',
                    pretty: true
                }
            }
        }
    };

    grunt.initConfig(gruntConfig);

    // Custom Grunt tasks
    grunt.registerTask('version', function() {
        var pkg = grunt.config.get('pkg');

        // Verify version string in source code matches what's in package.json
        var Raven = require('./src/raven');
        if (Raven.prototype.VERSION !== pkg.version) {
            return grunt.util.error('Mismatched version in src/raven.js: ' + Raven.prototype.VERSION +
                ' (should be ' + pkg.version + ')');
        }

        if (grunt.option('dev')) {
            pkg.release = 'dev';
        } else {
            pkg.release = pkg.version;
        }
        grunt.config.set('pkg', pkg);
    });

    // Grunt contrib tasks
    grunt.loadNpmTasks('grunt-contrib-uglify');
    grunt.loadNpmTasks('grunt-contrib-clean');
    grunt.loadNpmTasks('grunt-contrib-connect');
    grunt.loadNpmTasks('grunt-contrib-copy');

    // 3rd party Grunt tasks
    grunt.loadNpmTasks('grunt-browserify');
    grunt.loadNpmTasks('grunt-mocha');
    grunt.loadNpmTasks('grunt-release');
    grunt.loadNpmTasks('grunt-s3');
    grunt.loadNpmTasks('grunt-gitinfo');
    grunt.loadNpmTasks('grunt-sri');
    grunt.loadNpmTasks('grunt-eslint');

    // Build tasks
    grunt.registerTask('_prep', ['clean', 'gitinfo', 'version']);
    grunt.registerTask('browserify.core', ['_prep', 'browserify:core'].concat(browserifyPluginTaskNames));
    grunt.registerTask('browserify.plugins-combined', ['_prep', 'browserify:plugins-combined']);
    grunt.registerTask('build.test', ['_prep', 'browserify:test']);
    grunt.registerTask('build.core', ['browserify.core', 'uglify', 'sri:dist']);
    grunt.registerTask('build.plugins-combined', ['browserify.plugins-combined', 'uglify', 'sri:dist', 'sri:build']);
    grunt.registerTask('build', ['build.plugins-combined']);
    grunt.registerTask('dist', ['build.core', 'copy:dist']);

    // Test task
    grunt.registerTask('test', ['eslint', 'browserify.core', 'browserify:test', 'mocha']);

    // Webserver tasks
    grunt.registerTask('run:test', ['connect:test']);
    grunt.registerTask('run:docs', ['connect:docs']);

    grunt.registerTask('publish', ['test', 'build.plugins-combined', 's3']);
    grunt.registerTask('default', ['test']);
};
