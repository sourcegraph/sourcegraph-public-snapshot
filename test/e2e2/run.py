import os
import sys
import argparse
import traceback
from colors import green, red, bold

from e2etypes import *
from e2etests import *

def main():
    p = argparse.ArgumentParser()
    p.add_argument("--slow", help="run tests more slowly", action="store_true", default=False)
    p.add_argument("--interactive", help="wait for user to press ENTER after running each test", action="store_true", default=False)
    p.add_argument("--pause-on-err", help="pause on failure, so you can click around to see what happened", action="store_true", default=False)

    p.add_argument("--url", help="the URL of the Sourcegraph instance to be tested. In dev, use http://172.17.0.1:3080", default="https://sourcegraph.com") # 172.17.0.1 is the default value of the docker bridge IP, 10.0.2.2 is the default network gateway in VirtualBox VMs

    p.add_argument("--selenium", help="the address of the Selenium server instance to communicate with", default="http://localhost:4444")
    p.add_argument("--browser", help="the browser type (firefox or chrome)", default="chrome")
    p.add_argument("--filter", help="only run the tests matching this query", default="")

    args = p.parse_args()
    if args.browser.lower() not in ["chrome", "firefox"]:
        sys.stderr.write("browser needs to be chrome or firefox, was %s\n" % args.browser)
        return

    user_agent = "Sourcegraph e2etest-bot"

    tests = [t for t in all_tests if args.filter in t.func_name]
    failed_tests = []
    def success(test_name):
        print '[%s](%s) %s' % (green("PASS"), args.browser, test_name)
    def fail(test_name):
        failed_tests.append(test_name)
        print '[%s](%s) %s' % (red("FAIL"), args.browser, test_name)
    print '\nTest plan:\n%s\n' % '\n'.join(['\t'+f.func_name for f in tests])
    for test in tests:
        print '[%s](%s) %s' % (bold("RUN "), args.browser, test.func_name)
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
        except E2EFatal as e:
            print "E2EFatal: " + str(e)
            fail(test.func_name)
            if args.pause_on_err:
                print "PAUSED on error. Hit ENTER to continue"
                raw_input()
        except E2EError as e:
            print "E2EError: " + str(e)
            fail(test.func_name)
            if args.pause_on_err:
                print "PAUSED on error. Hit ENTER to continue"
                raw_input()
        except Exception as e:
            print "Uncaught exception when running test: "
            traceback.print_exc()
            fail(test.func_name)
            if args.pause_on_err:
                print "PAUSED on error. Hit ENTER to continue"
                raw_input()
        finally:
            if driver is not None:
                driver.close()

    print
    if len(failed_tests) > 0:
        print '%s: %d / %d FAILED\n' % (red("FAILURE"), len(failed_tests), len(tests))
        sys.exit(1)
    else:
        print green('ALL SUCCESS\n')

if __name__ == '__main__':
    main()
