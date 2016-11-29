import os
import sys
import argparse
import traceback
import atexit
import random

from urlparse import urlparse
from colors import green, red, bold
from slackclient import SlackClient

import requests

from e2etypes import *
from e2etests import *

user_agent = "Sourcegraph e2etest-bot"

extension_path = "/browser-ext"

emoji = [
    "dog",
    "cat",
    "mouse",
    "hamster",
    "rabbit",
    "bear",
    "koala",
    "tiger",
    "panda_face",
    "lion_face",
    "cow",
    "pig",
    "frog",
    "octopus",
    "monkey_face",
]

failure_msg_template = """:rotating_light: *TEST FAILED* :rotating_light:
*Test name*: `%s`
*Owner*: %s
*Browser*: %s
*URL*: %s
*Repro*: In directory `$sourcegraph-root/test/e2e2`, run `make test OPT="--pause-on-err --filter=%s" BROWSER=%s SOURCEGRAPH_URL=%s`
*Error*:
```
%s
```
%s:
```
%s
```
(For docs, see https://github.com/sourcegraph/sourcegraph/blob/master/test/e2e2/README.md)
"""

def failure_msg(test_name, owner, browser, url, sgurl, stack_trace, console_log):
    browser_log_title = "*Browser console* (only console.error for Firefox)"
    if browser == 'chrome':
        browser_log_title = "*Browser console* (console.error and console.log for Chrome)"
    return failure_msg_template % (
        test_name, owner,
        browser.capitalize(),
        url,
        test_name, browser, sgurl,
        stack_trace,
        browser_log_title, console_log,
    )

def alert_alertmanager(alertname, exported_name, url, alertmgr_url, oauth_cookie=None):
    alert = [{
        # The `severity: page` is necessary to alert OpsGenie
        "labels": { "alertname": alertname, "exported_name": exported_name, "severity": "page" },
        # The `description` and `details` fields are displayed in OpsGenie
        "annotations": { "description": ("%s: %s" % (alertname, exported_name)), "details": ("See: %s" % url) },
        "generatorURL": url,
    }]
    cookies = None
    if oauth_cookie is not None:
        cookies = { "_oauth2_proxy": oauth_cookie }
    try:
        resp = requests.post(('%s/api/v1/alerts' % alertmgr_url), cookies=cookies, json=alert)
        if resp.status_code == 200:
            logf("created Alertmanager alert: %s", id)
        else:
            logf("[ERROR] failed to create Alertmanager alert, response code: %d", resp.status_code)
    except Exception as err:
        logf("[ERROR] failed to post Alertmanager alert: %s", err.message)

def slack_and_alertmgr(args):
    slack_cli, slack_ch, alertmgr_url, alertmgr_cookie = None, None, None, None
    if args.alert_on_err:
        slack_tok, slack_ch = os.getenv("SLACK_API_TOKEN"), os.getenv("SLACK_WARNING_CHANNEL")
        if not slack_ch or not slack_tok:
            logf("If --alert-on-err is specified, environment variables SLACK_API_TOKEN and SLACK_WARNING_CHANNEL should be set. Exiting.")
            sys.exit(1)
        slack_cli = SlackClient(slack_tok)
        alertmgr_url = os.getenv("ALERT_MANAGER_URL", default="https://alertmanager.sgdev.org")
        alertmgr_cookie = os.getenv("ALERT_MANAGER_OAUTH_COOKIE")
    return slack_cli, slack_ch, alertmgr_url, alertmgr_cookie

# get_browser_log gets the browser logs, filtering out debugging messages
def get_browser_log(driver):
    def include(e):
        if '''%c%s%c LSP %s %s %c%sms''' in e['message']:
            return False
        if "console.groupEnd" in e['message']:
            return False
        return True
    return [e for e in driver.d.get_log('browser') if include(e)]

def run_tests(args, tests):
    failed_tests = []
    slack_cli, slack_ch, alertmgr_url, alertmgr_cookie = slack_and_alertmgr(args)

    def success(test_name):
        logf('[%s](%s) %s' % (green("PASS"), args.browser, test_name))

    def fail(test_name, owner, exception, driver):
        logf('[%s](%s) %s' % (red("FAIL"), args.browser, test_name))
        traceback.print_exc(30)
        console_log_msgs = get_browser_log(driver)
        if len(console_log_msgs) > 0:
            console_log = '\n'.join([('[%s] %s' % (e['level'], e['message'])) for e in console_log_msgs])
        else:
            console_log = "(None)"
        logf('Browser log:\n%s', console_log)
        if args.alert_on_err:
            msg = failure_msg(test_name, owner, args.browser, driver.d.current_url, args.url, traceback.format_exc(30), console_log)
            screenshot = driver.d.get_screenshot_as_png()
            resp = slack_cli.api_call("files.upload", channels=slack_ch, initial_comment=msg, file=screenshot, filename="screenshot.png")
            slack_alert_link = resp['file']['permalink']
            exported_name = '%s,%s' % (test_name, args.browser)
            alert_alertmanager("E2E failure", exported_name, slack_alert_link, alertmgr_url, oauth_cookie=alertmgr_cookie)
        if args.pause_on_err:
            print("""
#################################################################################################
PAUSED on error. You are now in the Python debugger (https://docs.python.org/2/library/pdb.html).
You can do things like `driver.d.find_element_by_id("my-id").click()`.
Type "continue" to continue.
#################################################################################################
""")
            import pdb; pdb.set_trace()

    logf('')
    logf('Starting test run with test plan:\n%s' % '\n'.join(['\t'+f[0].func_name for f in tests]))

    for test in tests:
        testfunc, owner = test
        if args.browser == "firefox" and "test_browser_extension" in testfunc.func_name:
            continue

        for i in xrange(0, args.tries_before_err):
            logf('[%s](%s) %s (attempt %d/%d)' % (bold("RUN "), args.browser, testfunc.func_name, i + 1, args.tries_before_err))
            try:
                driver, wd = None, None
                if args.browser == "chrome":
                    opt = DesiredCapabilities.CHROME.copy()
                    opt['chromeOptions'] = { "args": ["--user-agent=%s" % user_agent, "--load-extension=%s" % extension_path] }
                    opt['loggingPrefs'] = { 'browser': 'INFO' }
                    wd = webdriver.Remote(
                        command_executor=('%s/wd/hub' % args.selenium),
                        desired_capabilities=opt,
                    )
                elif args.browser == "firefox":
                    profile = webdriver.FirefoxProfile()
                    profile.set_preference('general.useragent.override', user_agent)
                    opt = DesiredCapabilities.FIREFOX.copy()
                    opt['loggingPrefs'] = { 'browser': 'SEVERE' }
                    wd = webdriver.Remote(
                        command_executor=('%s/wd/hub' % args.selenium),
                        desired_capabilities=opt,
                        browser_profile=profile,
                    )
                if args.slow:
                    wd.implicitly_wait(0.5)
                actual_user_agent = wd.execute_script("return navigator.userAgent")
                if actual_user_agent != user_agent:
                    raise Exception('user agent should be "%s", but was "%s"' % (user_agent, actual_user_agent))

                driver = Driver(wd, args.url)
                driver.d.maximize_window()
                driver.d.delete_all_cookies()
                testfunc(driver)
                success(testfunc.func_name)
                if args.interactive:
                    print("ENTER to continue ")
                    raw_input()
                break # on success, don't retry
            except (E2EError, E2EFatal, Exception) as e:
                if i == args.tries_before_err - 1: # if this is the last attempt, signal failure
                    test_name = testfunc.func_name
                    fail(test_name, owner, e, driver)
                    failed_tests.append(test_name)
            finally:
                if driver is not None:
                    driver.quit()
    if len(failed_tests) > 0:
        logf('Test run results: %s' % red("%d / %d FAILED" % (len(failed_tests), len(tests))))
    else:
        logf('Test run results: %s' % green('ALL SUCCESS'))
    return failed_tests

def main():
    p = argparse.ArgumentParser()
    p.add_argument("--slow", help="run tests more slowly", action="store_true", default=False)
    p.add_argument("--interactive", help="wait for user to press ENTER after running each test", action="store_true", default=False)
    p.add_argument("--pause-on-err", help="pause on failure, so you can click around to see what happened", action="store_true", default=False)

    p.add_argument("--url", help="the URL of the Sourcegraph instance to be tested. In dev, use http://172.17.0.1:3080", default="https://sourcegraph.com", type=str) # 172.17.0.1 is the default value of the docker bridge IP, 10.0.2.2 is the default network gateway in VirtualBox VMs

    p.add_argument("--selenium", help="the address of the Selenium server instance to communicate with", default="http://localhost:4444", type=str)
    p.add_argument("--browser", help="the browser type (firefox or chrome)", default="chrome", type=str)
    p.add_argument("--filter", help="only run the tests matching this query", default="", type=str)
    p.add_argument("--alert-on-err", help="send alert to Alertmanager and Slack on error. If this is true, the following environment variables should also be set: SLACK_API_TOKEN, SLACK_WARNING_CHANNEL", action="store_true", default=False)
    p.add_argument("--loop", help="loop continuously", action="store_true", default=False)
    p.add_argument("--tries-before-err", help="the number of times a test is tried before signaling failure", default=1, type=int)

    args = p.parse_args()
    if args.browser.lower() not in ["chrome", "firefox"]:
        sys.stderr.write("browser needs to be chrome or firefox, was %s\n" % args.browser)
        return

    tests = [t for t in all_tests if args.filter in t[0].func_name]

    if args.alert_on_err:
        slack_cli, slack_ch, _, __ = slack_and_alertmgr(args)
        animal_emoji = random.choice(emoji)
        animal_name = animal_emoji.replace('_face', '')
        slack_cli.api_call("chat.postMessage", channel=slack_ch, text=""":%s: Hi, I'm the end-to-end test %s for %s! I'll run the following tests in a loop and post errors to this channel until I retire:
```
%s
```
""" % (animal_emoji, animal_name, args.browser.capitalize(), '\n'.join([t[0].func_name for t in tests])))
        def die_msg():
            slack_cli.api_call("chat.postMessage", channel=slack_ch, text=":%s: *->* :skull: The end-to-end test %s for %s has died." % (animal_emoji, animal_name, args.browser.capitalize()))
        atexit.register(die_msg)

    if args.loop:
        logf("Looping forever...")
        while True:
            run_tests(args, tests)
    else:
        failed_tests = run_tests(args, tests)
        if len(failed_tests) > 0:
            sys.exit(1)

if __name__ == '__main__':
    random.seed()
    main()
