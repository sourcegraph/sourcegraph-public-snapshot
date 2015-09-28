{
  'targets': [
    {
      'target_name': 'libsass',
      'type': 'static_library',
      'defines': [
         'LIBSASS_VERSION="<!(node -e "process.stdout.write(require(\'../package.json\').libsass)")"'
      ],
      'sources': [
        'libsass/ast.cpp',
        'libsass/base64vlq.cpp',
        'libsass/bind.cpp',
        'libsass/cencode.c',
        'libsass/constants.cpp',
        'libsass/context.cpp',
        'libsass/contextualize.cpp',
        'libsass/contextualize_eval.cpp',
        'libsass/cssize.cpp',
        'libsass/emitter.cpp',
        'libsass/error_handling.cpp',
        'libsass/eval.cpp',
        'libsass/expand.cpp',
        'libsass/extend.cpp',
        'libsass/file.cpp',
        'libsass/functions.cpp',
        'libsass/inspect.cpp',
        'libsass/json.cpp',
        'libsass/lexer.cpp',
        'libsass/listize.cpp',
        'libsass/node.cpp',
        'libsass/output.cpp',
        'libsass/parser.cpp',
        'libsass/plugins.cpp',
        'libsass/position.cpp',
        'libsass/prelexer.cpp',
        'libsass/remove_placeholders.cpp',
        'libsass/sass.cpp',
        'libsass/sass2scss.cpp',
        'libsass/sass_context.cpp',
        'libsass/sass_functions.cpp',
        'libsass/sass_util.cpp',
        'libsass/sass_values.cpp',
        'libsass/source_map.cpp',
        'libsass/to_c.cpp',
        'libsass/to_string.cpp',
        'libsass/units.cpp',
        'libsass/utf8_string.cpp',
        'libsass/util.cpp'
      ],
      'cflags!': [
        '-fno-exceptions'
      ],
      'cflags_cc!': [
        '-fno-exceptions'
      ],
      'cflags_cc': [
        '-fexceptions',
        '-frtti',
      ],
      'direct_dependent_settings': {
        'include_dirs': [ 'libsass' ],
      },
      'conditions': [
        ['OS=="mac"', {
          'xcode_settings': {
            'OTHER_CPLUSPLUSFLAGS': [
              '-std=c++11',
              '-stdlib=libc++'
            ],
            'OTHER_LDFLAGS': [],
            'GCC_ENABLE_CPP_EXCEPTIONS': 'YES',
            'GCC_ENABLE_CPP_RTTI': 'YES',
            'MACOSX_DEPLOYMENT_TARGET': '10.7'
          }
        }],
        ['OS=="win"', {
          'msvs_settings': {
            'VCCLCompilerTool': {
              'AdditionalOptions': [
                '/GR',
                '/EHs'
              ]
            }
          }
        }],
        ['OS!="win"', {
          'cflags_cc+': [
            '-std=c++0x'
          ]
        }]
      ]
    }
  ]
}
