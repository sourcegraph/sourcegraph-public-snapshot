import sys
import time
import math
import traceback
import urllib
import subprocess
import os
import json

from datetime import datetime

from selenium import webdriver
from selenium.webdriver.common.desired_capabilities import DesiredCapabilities
from selenium.common.exceptions import StaleElementReferenceException, NoSuchElementException, ElementNotVisibleException, WebDriverException
from selenium.webdriver.common.action_chains import ActionChains
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.keys import Keys

# E2EFatal should be raised on any error that should halt further test progress
class E2EFatal(Exception):
    def __init__(self, *args, **kwargs):
        Exception.__init__(self, *args, **kwargs)

# E2EError should be raised on any condition that should trigger a test failure
class E2EError(Exception):
    def __init__(self, *args, **kwargs):
        Exception.__init__(self, *args, **kwargs)

# logf is a convenience logging function that wraps sys.stderr.write,
# but prefixes each line with a timestamp.
def logf(s, *args):
    ts = datetime.fromtimestamp(time.time()).strftime('%Y-%m-%d %H:%M:%S')
    sf = s % args
    sys.stderr.write("%s %s\n" % (ts, sf))

# wait_for waits for condition to be true. If condition does not hold
# immediately, it sleeps in increments of wait_incr until max_wait
# seconds. Its behavior is similar to Selenium's WebDriverWait and it's
# easier to use. This function should be called from e2etests, but NOT
# by methods in the Driver class.
def wait_for(condition, max_wait=2.0, wait_incr=0.1, text=""):
    if len(text) > 0:
        text = '"%s"' % text

    time_waited = 0.0
    while time_waited < max_wait:
        try:
            res = condition()
            if res is None or res:
                return
        except (StaleElementReferenceException, NoSuchElementException, ElementNotVisibleException):
            pass
        time.sleep(max(0, min(wait_incr, max_wait - time_waited)))
        time_waited += wait_incr
    try:
        if not condition():
            raise E2EError("timed out waiting for condition %s" % text)
    except (StaleElementReferenceException, NoSuchElementException, ElementNotVisibleException):
        raise E2EError("timed out waiting for condition %s" % text)

# retry calls fn a maximum of $attempts times, waiting $cooldown
# seconds in between invocations. It returns the return value of the
# first invocation of fn that does not raise an exception. This
# function should be called from e2etests, but NOT by methods in the
# Driver class.
def retry(fn, attempts=3, cooldown=0):
    for i in xrange(0, attempts):
        try:
            return fn()
        except (StaleElementReferenceException, NoSuchElementException, ElementNotVisibleException, WebDriverException, E2EError) as e:
            if i == attempts - 1:
                raise e
        time.sleep(cooldown)

# distance returns the L2 pixel distance between two elements
def distance(e, f):
    dx = e.location['x'] - f.location['x']
    dy = e.location['y'] - f.location['y']
    return math.sqrt((dx*dx) + (dy*dy))

# Driver is driver that tests should use to interact with the browser.
# It provides convenience methods on top of the Selenium web driver.
#
# Methods in this class should specify "atomic" actions and therefore
# avoid calling wait_for and retry. If you find yourself using either
# wait_for or retry in a Driver method, STOP and move the logic into a
# method in the Util class.
class Driver(object):
    def __init__(self, wd, sourcegraph_url):
        self.d = wd  # d is the Selenium WebDriver that Driver uses to interact with the browser
        self.sourcegraph_url = sourcegraph_url

    def quit(self):
        self.d.quit()

    def sg_url(self, urlpath):
        if len(urlpath) == 0 or urlpath[0] != "/":
            raise ValueError("urlpath should begin with '/'")
        return "%s%s" % (self.sourcegraph_url, urlpath)

    # active_elem provides a robust way to get an element that has focus and is visible.
    # Visibility is important because the Firefox server will not permit interaction with
    # an element that's not visible.
    def active_elem(self):
        e = self.d.switch_to.active_element
        if e.is_displayed():
            return e
        return self.d.find_element_by_tag_name('body')

    def all_network_indicators_are_invisible(self):
        return len(self.d.find_elements_by_css_selector(".uil-default")) == 0

    def right_click(self, elem):
        ActionChains(self.d).context_click(elem).perform()

    def double_click(self, elem):
        ActionChains(self.d).double_click(elem).perform()

    def find_context_menu_option(self, option_text):
        menu = self.d.find_element_by_css_selector(".monaco-menu")
        peek_items = [e for e in menu.find_elements_by_css_selector(".action-label") if option_text in e.text]
        if peek_items is None or len(peek_items) != 1:
            raise E2EError('expected exactly one "%s" option in menu, but found %d' % (option_text, len(peek_items)))
        return peek_items[0]

    def hover_token(self, token_text):
        ActionChains(self.d).move_to_element(self.find_token(token_text)).perform()

    def hover_elem(self, elem):
        ActionChains(self.d).move_to_element(elem).perform()

    def find_tokens(self, tok_text):
        return [e for e in self.d.find_elements_by_css_selector(".token.identifier.go") if tok_text in e.text]

    def find_token(self, tok_text, select_any=True):
        candidates = self.find_tokens(tok_text)
        if len(candidates) == 0:
            raise E2EError('no tokens found with "%s"', tok_text)
        elif len(candidates) > 1 and not select_any:
            raise E2EError('more than one token found with "%s"', tok_text)
        return candidates[0]

    def find_references_menu_options(self):
        refs_menu = self.d.find_element_by_css_selector(".ref-tree.inline")
        return refs_menu.find_elements_by_css_selector(".referenceMatch")

    def find_tooltip_near_elem(self, elem):
        tt = self.d.find_element_by_css_selector(".cdr.hoverHighlight")
        dist = distance(tt, elem)
        if dist <= 100:
            return tt
        raise E2EError("no tooltips within 5px of element %s#%s, nearest was %d" % (tt.tag_name, tt.id, dist))

    def find_search_modal_selected_result(self):
        return self.d.find_element_by_css_selector("[data-class-name~=modal-result-selected]")

    def find_search_modal_results(self, results_text, exact_match=False):
        results = self.d.find_elements_by_css_selector("[data-class-name~=modal-result]")
        if exact_match:
            return [r for r in results if results_text == r.text]
        return [r for r in results if results_text in r.text]

    def find_button_by_partial_text(self, text):
        btns = self.find_buttons_by_partial_text(text)
        if len(btns) == 0:
            raise E2EError('expected, but didn\'t find button with text "%s"' % text)
        return btns[0]

    def find_buttons_by_partial_text(self, text):
        return [e for e in self.d.find_elements_by_tag_name("button") if text in e.text]

    def find_elements_by_tag_name_and_partial_text(self, tag_name, text):
        return [e for e in self.d.find_elements_by_tag_name(tag_name) if text in e.text]

    def delete_user_if_exists(self, username):
        auth0_tok = os.getenv("AUTH0_TOKEN")
        query = urllib.quote('nickname:"%s"' % username)
        auth0_users_out = subprocess.check_output('curl -H "Authorization: Bearer %s" https://sourcegraph.auth0.com/api/v2/users?q=%s' % (auth0_tok, query), shell=True)
        auth0_users = json.loads(auth0_users_out)
        if isinstance(auth0_users, list) and len(auth0_users) == 0:
            return
        if not isinstance(auth0_users, list):
            raise Exception("expected Auth0 result to be a list, but was %s", str(auth0_users))
        if not (isinstance(auth0_users[0], dict) and 'user_id' in auth0_users[0]):
            raise Exception("expected Auth0 result to contain user_id but did not, result was %s", str(auth0_users[0]))
        if len(auth0_users) > 1:
            raise Exception("expected at most 1 Auth0 user with username %s but found %d" % (username, len(auth0_users)))
        auth0_user_id = auth0_users[0]['user_id']
        subprocess.check_output('curl -H "Authorization: Bearer %s" -X DELETE  https://sourcegraph.auth0.com/api/v2/users/%s' % (auth0_tok, urllib.quote(auth0_user_id)), shell=True)

    def verify_new_tab_opened(self, location):
        main_window = self.d.current_window_handle
        wait_for(lambda: len(self.d.window_handles) == 2)
        retry(lambda: self.d.switch_to.window(self.d.window_handles[1]))
        wait_for(lambda: self.d.current_url == location)
        self.d.close()
        retry(lambda: self.d.switch_to.window(main_window))

# Util contains static methods that define more compound actions
# than what are available in Driver methods.
class Util(object):

    @staticmethod
    def log_in(d, username, password):
        wd = d.d
        wd.find_element_by_link_text("Log in").click()
        d.find_button_by_partial_text("Continue with GitHub").click()
        wd.find_element_by_id("login_field").send_keys(username)
        wd.find_element_by_id("password").send_keys(password)
        d.active_elem().send_keys(Keys.ENTER)
        if len(d.find_buttons_by_partial_text("Authorize application")) > 0:
            d.find_button_by_partial_text("Authorize application").click()

    @staticmethod
    def log_out(d):
        wd = d.d
        wd.find_element_by_id("global-nav").find_element_by_css_selector('[class*="popover"]').click()
        wd.find_element_by_partial_link_text("Sign out").click()

    @staticmethod
    def select_search_result_using_arrow_keys(d, result_text, exact_match=False):
        for i in xrange(0, 2):
            for i in xrange(0, 20):
                def f():
                    selected = d.find_search_modal_selected_result()
                    if (exact_match and result_text == selected.text) or (not exact_match and result_text in selected.text):
                        d.active_elem().send_keys(Keys.ENTER)
                        return True
                    return False
                if retry(f):
                    return
                d.active_elem().send_keys(Keys.DOWN)
            for i in xrange(0, 40): # network events might have changed the list, so try one more time from the top
                d.active_elem().send_keys(Keys.UP)
        raise E2EError("did not find search result '%s'" % result_text)
