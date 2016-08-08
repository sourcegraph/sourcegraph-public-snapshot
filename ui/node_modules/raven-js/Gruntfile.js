module.exports = function(grunt) {
    "use strict";

    var _ = require('lodash');
    var path = require('path');

    var coreFiles = [
        'template/_header.js',
        'vendor/**/*.js',
        'src/**/*.js',
        'template/_footer.js'
    ];

    var plugins = grunt.option('plugins');
    // Create plugin paths and verify hey exist
    plugins = _.map(plugins ? plugins.split(',') : [], function (plugin) {
        var path = 'plugins/' + plugin + '.js';

        if(!grunt.file.exists(path))
            throw new Error("Plugin '" + plugin + "' not found in plugins directory.");

        return path;
    });

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

        var all = [a];

        for (var i = 0; i < a.length; i++) {
            fn(i, a, [], all);
        }

        return all;
    };

    var pluginCombinations = combine(grunt.file.expand('plugins/*.js'));
    var pluginConcatFiles = _.reduce(pluginCombinations, function (dict, comb) {
        var key = _.map(comb, function (plugin) {
            return path.basename(plugin, '.js');
        });
        key.sort();

        var dest = path.join('build/', key.join(','), '/raven.js');
        dict[dest] = coreFiles.concat(comb);

        return dict;
    }, {});

    var gruntConfig = {
        pkg: grunt.file.readJSON('package.json'),
        aws: grunt.file.exists('aws.json') ? grunt.file.readJSON('aws.json'): {},

        clean: ['build'],
        concat: {
            options: {
                separator: '\n',
                banner: grunt.file.read('template/_copyright.js'),
                process: true
            },
            core: {
                src: coreFiles.concat(plugins),
                dest: 'build/raven.js'
            },
            all: {
                files: pluginConcatFiles
            }
        },

        uglify: {
            options: {
                sourceMap: function (dest) {
                    return path.join(path.dirname(dest),
                                     path.basename(dest, '.js')) +
                           '.map';
                },
                sourceMappingURL: function (dest) {
                    return path.basename(dest, '.js') + '.map';
                },
                preserveComments: 'some'
            },
            dist: {
                src: ['build/**/*.js'],
                ext: '.min.js',
                expand: true
            }
        },

        fixSourceMaps: {
            all: ['build/**/*.map']
        },

        jshint: {
            options: {
                jshintrc: '.jshintrc'
            },
            all: ['Gruntfile.js', 'src/**/*.js', 'plugins/**/*.js']
        },

        mocha: {
            all: {
                options: {
                    mocha: {
                        ignoreLeaks: true,
                        grep:        grunt.option('grep')
                    },
                    log:      true,
                    reporter: 'Dot',
                    run:      true
                },
                src: ['test/index.html'],
                nonull: true
            }
        },

        release: {
            options: {
                npm:           false,
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
                    base: 'docs/html',
                    keepalive: true
                }
            }
        },

        copy: {
            dist: {
                expand: true,
                flatten: true,
                cwd: 'build/',
                src: '**',
                dest: 'dist/'
            }
        }
    };

    grunt.initConfig(gruntConfig);

    // Custom Grunt tasks
    grunt.registerTask('version', function() {
        var pkg = grunt.config.get('pkg');
        if (grunt.option('dev')) {
            pkg.release = 'dev';
            pkg.version = grunt.config.get('gitinfo').local.branch.current.shortSHA;
        } else {
            pkg.release = pkg.version;
        }
        grunt.config.set('pkg', pkg);
    });

    grunt.registerMultiTask('fixSourceMaps', function () {
        this.files.forEach(function (f) {
            var result;
            var sources = f.src.filter(function (filepath) {
                if (!grunt.file.exists(filepath)) {
                    grunt.log.warn('Source file "' + filepath + '" not found.');
                    return false;
                } else {
                    return true;
                }
            }).forEach(function (filepath) {
                var base = path.dirname(filepath);
                var sMap = grunt.file.readJSON(filepath);
                sMap.file = path.relative(base, sMap.file);
                sMap.sources = _.map(sMap.sources, path.relative.bind(path, base));

                grunt.file.write(filepath, JSON.stringify(sMap));
                // Print a success message.
                grunt.log.writeln('File "' + filepath + '" fixed.');
            });
        });
    });

    // Grunt contrib tasks
    grunt.loadNpmTasks('grunt-contrib-uglify');
    grunt.loadNpmTasks('grunt-contrib-concat');
    grunt.loadNpmTasks('grunt-contrib-clean');
    grunt.loadNpmTasks('grunt-contrib-jshint');
    grunt.loadNpmTasks('grunt-contrib-connect');
    grunt.loadNpmTasks('grunt-contrib-copy');

    // 3rd party Grunt tasks
    grunt.loadNpmTasks('grunt-mocha');
    grunt.loadNpmTasks('grunt-release');
    grunt.loadNpmTasks('grunt-s3');
    grunt.loadNpmTasks('grunt-gitinfo');

    // Build tasks
    grunt.registerTask('_prep', ['clean', 'gitinfo', 'version']);
    grunt.registerTask('concat.core', ['_prep', 'concat:core']);
    grunt.registerTask('concat.all', ['_prep', 'concat:all']);
    grunt.registerTask('build.core', ['concat.core', 'uglify', 'fixSourceMaps']);
    grunt.registerTask('build.all', ['concat.all', 'uglify', 'fixSourceMaps']);
    grunt.registerTask('build', ['build.all']);
    grunt.registerTask('dist', ['build.core', 'copy:dist']);

    // Test task
    grunt.registerTask('test', ['jshint', 'mocha']);

    // Webserver tasks
    grunt.registerTask('run:test', ['connect:test']);
    grunt.registerTask('run:docs', ['connect:docs']);

    grunt.registerTask('publish', ['test', 'build.all', 's3']);
    grunt.registerTask('default', ['test']);
};
