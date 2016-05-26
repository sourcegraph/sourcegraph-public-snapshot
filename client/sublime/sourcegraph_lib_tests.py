import os
import subprocess
import sys
import unittest

try:
    from unittest import mock
except:
    import mock

sys.path.append(os.path.dirname(__file__))
import sourcegraph_lib

from test_data import Tests


def start_default_instance():
    test_settings = sourcegraph_lib.Settings()
    sourcegraph_lib_instance = sourcegraph_lib.Sourcegraph(test_settings)
    sourcegraph_lib_instance.post_load()
    return sourcegraph_lib_instance


def check_output(test_case_class, test_output, expected_output):
    if expected_output is None:
        test_case_class.assertIsNone(test_output)
        return
    test_case_class.assertEqual(expected_output, test_output)


def full_test_filename(test_name, gopath):
    return os.path.join(gopath, 'src', 'github.com', 'luttig', 'sg-live-plugin-tests', 'go_tests', test_name)

def run_go_test(test, sourcegraph_lib_instance):
    full_filename = full_test_filename(test.lookup_args.filename, sourcegraph_lib_instance.settings.ENV['GOPATH'])
    buff = b''
    if test.lookup_args.filename != '' and test.lookup_args.filename != '.go':
        try:
            with open(full_filename, 'r') as test_file:
                buff = test_file.read().encode()
        except:
            pass
    return sourcegraph_lib_instance.get_sourcegraph_request(full_filename, test.lookup_args.cursor_offset, buff, test.lookup_args.selected_token)

class VerifyGoodGopath(unittest.TestCase):
    def test(self):
        sourcegraph_lib_instance = start_default_instance()
        sourcegraph_lib_instance.settings.ENV['GOPATH'] = '/path/does/not/exist'
        validate_output = sourcegraph_lib.validate_settings(sourcegraph_lib_instance.settings)
        self.assertEqual(validate_output, sourcegraph_lib.ERR_GOPATH_UNDEFINED)


class VerifyClearCacheOnHardReload(unittest.TestCase):
    def test(self):
        sourcegraph_lib_instance = start_default_instance()
        sourcegraph_lib_instance.open_channel_os = mock.Mock()
        self.assertIsNone(sourcegraph_lib_instance.EXPORTED_PARAMS_CACHE)
        self.assertEqual(sourcegraph_lib_instance.HAVE_OPENED_CHANNEL, False)

        imported_struct_test = Tests().IMPORTED_STRUCT
        imported_struct_test.lookup_args.filename = full_test_filename(imported_struct_test.lookup_args.filename, sourcegraph_lib_instance.settings.ENV['GOPATH'])
        buff = b''
        try:
            with open(imported_struct_test.lookup_args.filename, 'r') as test_file:
                buff = test_file.read().encode()
        except:
            pass
        imported_struct_test.lookup_args.preceding_selection = buff
        sourcegraph_lib_instance.on_selection_modified_handler(imported_struct_test.lookup_args)
        self.assertEqual(sourcegraph_lib_instance.HAVE_OPENED_CHANNEL, True)
        self.assertEqual(sourcegraph_lib_instance.EXPORTED_PARAMS_CACHE, imported_struct_test.expected_output)

        sourcegraph_lib_instance.open_channel(hard_refresh=True)
        self.assertIsNone(sourcegraph_lib_instance.EXPORTED_PARAMS_CACHE)
        self.assertEqual(sourcegraph_lib_instance.HAVE_OPENED_CHANNEL, True)

        sourcegraph_lib_instance.settings.AUTO = True
        sourcegraph_lib_instance.on_selection_modified_handler(imported_struct_test.lookup_args)
        self.assertEqual(sourcegraph_lib_instance.HAVE_OPENED_CHANNEL, True)
        self.assertEqual(sourcegraph_lib_instance.EXPORTED_PARAMS_CACHE, imported_struct_test.expected_output)


class VerifySyntaxVarieties(unittest.TestCase):
    def test(self):
        syntax_tests = Tests().syntax_tests()
        for test in syntax_tests:
            sourcegraph_lib_instance = start_default_instance()
            test_params = syntax_tests[test]
            test_output = run_go_test(test_params, sourcegraph_lib_instance)
            check_output(self, test_output, test_params.expected_output)


class VerifyGoPathEmptyError(unittest.TestCase):
    def test(self):
        sourcegraph_lib_instance = start_default_instance()
        sourcegraph_lib_instance.settings.ENV['GOPATH'] = ''
        test = Tests.GOPATH_EMPTY
        test_output = run_go_test(test, sourcegraph_lib_instance)
        check_output(self, test_output, test.expected_output)

        error_message = sourcegraph_lib_instance.add_gopath_to_path()
        self.assertEqual(error_message.title, sourcegraph_lib.ERR_GOPATH_UNDEFINED.title)
        self.assertEqual(error_message.description, sourcegraph_lib.ERR_GOPATH_UNDEFINED.description)


# When godefinfo binary not yet installed
class VerifyGodefInfoInstallError(unittest.TestCase):
    def test(self):
        sourcegraph_lib_instance = start_default_instance()
        sourcegraph_lib_instance.settings.ENV['GOPATH'] = '/'
        sourcegraph_lib_instance.settings.ENV['PATH'] = '/path/that/does/not/contain/godefinfo'
        test = Tests.GODEFINFO_INSTALL
        test_output = run_go_test(test, sourcegraph_lib_instance)
        check_output(self, test_output, test.expected_output)


class VerifyGoBinaryError(unittest.TestCase):
    def test(self):
        sourcegraph_lib_instance = start_default_instance()
        sourcegraph_lib_instance.settings.GOBIN = '/path/not/to/gobin'
        test_output = sourcegraph_lib.godefinfo_auto_install(sourcegraph_lib_instance.settings.GOBIN, sourcegraph_lib_instance.settings.ENV)
        self.assertEqual(test_output.Error, sourcegraph_lib.ERR_GO_BINARY.title)
        self.assertEqual(test_output.Fix, sourcegraph_lib.ERR_GO_BINARY.description)


class ValidateSettings(unittest.TestCase):
    def test(self):
        settings = sourcegraph_lib.Settings()
        validation_output = sourcegraph_lib.validate_settings(settings)
        if validation_output is not None:
            self.fail('Failed settings validation: %s' % validation_output.title)


class VerifyGodefinfoAutoUpdate(unittest.TestCase):
    def git_commit(self, git_dir):
        get_commit_process = subprocess.Popen(['git', 'rev-parse', 'HEAD'], cwd=git_dir, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        commit, stderr = get_commit_process.communicate()
        self.assertEqual(stderr, b'')
        return commit

    def test(self):
        test_settings = sourcegraph_lib.Settings()
        sourcegraph_lib_instance = start_default_instance()
        godefinfo_dir = os.path.join(sourcegraph_lib_instance.settings.ENV['GOPATH'], 'src', 'github.com', 'sqs', 'godefinfo')
        current_commit = self.git_commit(godefinfo_dir)

        reset_commit_process = subprocess.Popen(['git', 'reset', '--hard', 'HEAD~'], cwd=godefinfo_dir, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        reset_commit, stderr = reset_commit_process.communicate()
        self.assertEqual(stderr, b'')

        old_commit = self.git_commit(godefinfo_dir)
        self.assertNotEqual(current_commit, old_commit)

        sourcegraph_lib.Sourcegraph(test_settings).post_load()
        self.assertEqual(current_commit, self.git_commit(godefinfo_dir))


# When godef is called on package that user doesn't have, send error message to user on how to install it
class VerifyGoGetError(unittest.TestCase):
    def test(self):
        return  # TODO
        sourcegraph_lib_instance = start_default_instance()
        package_name = 'asdf'
        test = Tests.BAD_IMPORT_PATH
        test.expected_output = sourcegraph_lib.ExportedParams(Error=sourcegraph_lib.ERR_GO_GET(package_name).title, Fix=sourcegraph_lib.ERR_GO_GET(package_name).description)
        test_output = run_go_test(test, sourcegraph_lib_instance)
        check_output(self, test_output, test.expected_output)

if __name__ == '__main__':
    unittest.main()
