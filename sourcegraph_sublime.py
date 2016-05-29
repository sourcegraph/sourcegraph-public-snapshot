import os
import sys

sys.path.append(os.path.dirname(__file__))
import sourcegraph_lib

import sublime

import sublime_plugin

SETTINGS_FILENAME = 'Sourcegraph.sublime-settings'
GOSUBLIME_SETTINGS_FILENAME = 'GoSublime.sublime-settings'
SUBLIME_VERSION = int(sublime.version())

SG_LIB_INSTANCE = {}


def find_gopath_from_gosublime():
	if sublime.load_settings(GOSUBLIME_SETTINGS_FILENAME).has('env'):
		gosubl_env = sublime.load_settings(GOSUBLIME_SETTINGS_FILENAME).get('env')
		if 'GOPATH' in gosubl_env:
			return gosubl_env['GOPATH'].replace('$HOME', sourcegraph_lib.get_home_path()).replace(':$GS_GOPATH', '')
	return None


def load_settings(settings):
	sg_settings = sourcegraph_lib.Settings()

	if settings.has('LOG_LEVEL'):
		sourcegraph_lib.LOG_LEVEL = settings.get('LOG_LEVEL')
		sourcegraph_lib.log_output('[settings] Found logging setting in Settings file: %s' % sourcegraph_lib.LOG_LEVEL)
	if settings.has('ENABLE_LOOKBACK'):
		sg_settings.ENABLE_LOOKBACK = settings.get('ENABLE_LOOKBACK')
	if settings.has('SG_BASE_URL'):
		sg_settings.SG_BASE_URL = settings.get('SG_BASE_URL').rstrip('/')
	if settings.has('SG_SEND_URL'):
		sg_settings.SG_SEND_URL = settings.get('SG_SEND_URL').rstrip('/')
	if settings.has('SG_LOG_FILE'):
		sourcegraph_lib.SG_LOG_FILE = settings.get('SG_LOG_FILE')
	if settings.has('AUTO'):
		sg_settings.AUTO = settings.get('AUTO')
	if settings.has('GOBIN'):
		sg_settings.GOBIN = settings.get('GOBIN').rstrip(os.sep)
	shell_gopath = sourcegraph_lib.find_gopath_from_shell(sg_settings.ENV.get('SHELL'))
	if settings.has('GOPATH'):
		sg_settings.ENV['GOPATH'] = str(settings.get('GOPATH').rstrip(os.sep)).strip()
		sourcegraph_lib.log_output('[settings] Using GOPATH found in Sublime settings file: %s' % sg_settings.ENV['GOPATH'])
	elif shell_gopath and shell_gopath.rstrip(os.sep).strip() != '':
		sg_settings.ENV['GOPATH'] = shell_gopath.rstrip(os.sep).strip()
		sourcegraph_lib.log_output('[settings] Using GOPATH from shell: %s' % sg_settings.ENV['GOPATH'])
	elif find_gopath_from_gosublime():
		sg_settings.ENV['GOPATH'] = find_gopath_from_gosublime()
		sourcegraph_lib.log_output('[settings] Found GOPATH in GoSublime settings: %s' % sg_settings.ENV['GOPATH'])
	if 'GOPATH' in sg_settings.ENV and sg_settings.ENV.get('GOPATH') != '':
		sg_settings.ENV['GOPATH'] = sg_settings.ENV['GOPATH'].replace('~', sourcegraph_lib.get_home_path())

	global SG_LIB_INSTANCE
	SG_LIB_INSTANCE = sourcegraph_lib.Sourcegraph(sg_settings)
	SG_LIB_INSTANCE.post_load()


def reload_settings():
	old_base_url = SG_LIB_INSTANCE.settings.SG_BASE_URL
	settings = sublime.load_settings(SETTINGS_FILENAME)
	load_settings(settings)
	if SG_LIB_INSTANCE.settings.SG_BASE_URL != old_base_url and SG_LIB_INSTANCE.settings.AUTO:
		SG_LIB_INSTANCE.open_channel(hard_refresh=True)


def plugin_loaded():
	settings = sublime.load_settings(SETTINGS_FILENAME)
	settings.clear_on_change('sg-reload')
	settings.add_on_change('sg-reload', reload_settings)

	gosublime_settings = sublime.load_settings(GOSUBLIME_SETTINGS_FILENAME)
	gosublime_settings.clear_on_change('sg-reload-gosubl')
	gosublime_settings.add_on_change('sg-reload-gosubl', reload_settings)

	load_settings(settings)


def cursor_offset(view):
	row, col = view.rowcol(view.sel()[0].begin())
	symbol_start_offset = sourcegraph_lib.search_for_symbols(view.sel()[0].begin(), view.substr(view.line(view.sel()[0])), row, col, SG_LIB_INSTANCE.settings.ENABLE_LOOKBACK)

	string_before = view.substr(sublime.Region(0, symbol_start_offset))
	string_before.encode('utf-8')
	buffer_before = bytearray(string_before, encoding='utf8')
	return str(len(buffer_before))


def process_selection(view):
	if len(view.sel()) == 0:
		return
	args = sourcegraph_lib.LookupArgs(filename=view.file_name(), cursor_offset=cursor_offset(view), preceding_selection=str.encode(view.substr(sublime.Region(0, view.size()))), selected_token=view.substr(view.extract_scope(view.sel()[0].begin())))
	SG_LIB_INSTANCE.on_selection_modified_handler(args)


class SgOpenLogCommand(sublime_plugin.WindowCommand):
	def run(self):
		self.window.open_file(sourcegraph_lib.SG_LOG_FILE)


class SgManualProcessCommand(sublime_plugin.TextCommand):
	def run(self, edit):
		process_selection(self.view)


class SgAutoProcessCommand(sublime_plugin.EventListener):
	def __init__(self):
		super().__init__()

	def on_selection_modified_async(self, view):
		if SG_LIB_INSTANCE.settings.AUTO:
			process_selection(view)
