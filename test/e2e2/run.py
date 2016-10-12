import os
import sys
import argparse
import traceback
from urlparse import urlparse
from colors import green, red, bold
from slackclient import SlackClient

from e2etypes import *
from e2etests import *

user_agent = "Sourcegraph e2etest-bot"

failure_msg_template = """:rotating_light: *TEST FAILED* :rotating_light:
*Test name*: `%s`
*Browser*: %s
*URL*: %s
*Repro*: `cd test/e2e2 && make TV=1 OPT="--pause-on-err" BROWSERS=%s TEST=%s SOURCEGRAPH_URL=%s`
*Error*:
```
%s
```
(For docs, see https://github.com/sourcegraph/sourcegraph/blob/master/test/e2e2/README.md)
"""

def failure_msg(test_name, browser, url, stack_trace):
    u = urlparse(url)
    sgurl = "%s://%s" % (u.scheme, u.hostname)
    if u.port:
        sgurl += ":%d" % u.port
    return failure_msg_template % (
        test_name, browser, url,
        browser.lower(), test_name, sgurl,
        stack_trace,
    )

def run_tests(args, tests):
    failed_tests = []

    slack_cli, slack_ch, opsgenie_key = None, None, None
    if args.alert_on_err:
        slack_tok, slack_ch, opsgenie_key = os.getenv("SLACK_API_TOKEN"), os.getenv("SLACK_WARNING_CHANNEL"), os.getenv("OPSGENIE_KEY")
        if not slack_ch or not slack_tok or not opsgenie_key:
            print "if --alert-on-err is specified, environment variables SLACK_API_TOKEN, SLACK_WARNING_CHANNEL, and OPSGENIE_KEY should be set"
            sys.exit(1)
        slack_cli = SlackClient(slack_tok)

    def success(test_name):
        print '[%s](%s) %s' % (green("PASS"), args.browser, test_name)

    def fail(test_name, exception, driver):
        print '[%s](%s) %s' % (red("FAIL"), args.browser, test_name)
        traceback.print_exc(30)
        if args.alert_on_err:
            msg = failure_msg(test_name, args.browser.capitalize(), driver.d.current_url, traceback.format_exc(30))
            screenshot = driver.d.get_screenshot_as_png()
            slack_cli.api_call("files.upload", channels=slack_ch, initial_comment=msg, file=screenshot, filename="screenshot.png")
        if args.pause_on_err:
            print "PAUSED on error. Hit ENTER to continue"
            raw_input()

    print '\nTest plan:\n%s\n' % '\n'.join(['\t'+f.func_name for f in tests])

    for test in tests:
        for i in xrange(0, args.tries_before_err):
            print '[%s](%s) %s (attempt %d/%d)' % (bold("RUN "), args.browser, test.func_name, i + 1, args.tries_before_err)
            try:
                driver = None
                if args.browser == "chrome":
                    opt = DesiredCapabilities.CHROME.copy()
                    opt['chromeOptions'] = { "args": ["--user-agent=%s" % user_agent] }
                elif args.browser == "firefox":
                    opt = DesiredCapabilities.FIREFOX.copy()
                    opt['general.useragent.override'] = user_agent

                wd = webdriver.Remote(
                    command_executor=('%s/wd/hub' % args.selenium),
                    desired_capabilities=opt,
                )
                if args.slow:
                    wd.implicitly_wait(0.5)

                driver = Driver(wd, args.url)
                driver.d.maximize_window()
                test(driver)
                success(test.func_name)
                if args.interactive:
                    print "ENTER to continue "
                    raw_input()
                break # on success, don't retry
            except (E2EError, E2EFatal, Exception) as e:
                if i == args.tries_before_err - 1: # if this is the last attempt, signal failure
                    test_name = test.func_name
                    fail(test_name, e, driver)
                    failed_tests.append(test_name)
            finally:
                if driver is not None:
                    driver.close()
    print
    if len(failed_tests) > 0:
        print '%s: %d / %d FAILED\n' % (red("FAILURE"), len(failed_tests), len(tests))
    else:
        print green('ALL SUCCESS\n')
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
    p.add_argument("--alert-on-err", help="send alert to OpsGenie and Slack on error. If this is true, the following environment variables should also be set: SLACK_API_TOKEN, SLACK_WARNING_CHANNEL, OPSGENIE_KEY", action="store_true", default=False)
    p.add_argument("--loop", help="loop continuously", action="store_true", default=False)
    p.add_argument("--tries-before-err", help="the number of times a test is tried before signaling failure", default=1, type=int)

    args = p.parse_args()
    if args.browser.lower() not in ["chrome", "firefox"]:
        sys.stderr.write("browser needs to be chrome or firefox, was %s\n" % args.browser)
        return

    tests = [t for t in all_tests if args.filter in t.func_name]
    if args.loop:
        print "Looping forever..."
        while True:
            run_tests(args, tests)
    else:
        failed_tests = run_tests(args, tests)
        if len(failed_tests) > 0:
            sys.exit(1)

if __name__ == '__main__':
    main()
