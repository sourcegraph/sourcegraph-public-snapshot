import time
import math
import traceback

from selenium import webdriver
from selenium.webdriver.common.desired_capabilities import DesiredCapabilities
from selenium.common.exceptions import StaleElementReferenceException, NoSuchElementException, ElementNotVisibleException, WebDriverException
from selenium.webdriver.common.action_chains import ActionChains
from selenium.webdriver.chrome.options import Options

# E2EFatal should be raised on any error that should halt further test progress
class E2EFatal(Exception):
    def __init__(self, *args, **kwargs):
        Exception.__init__(self, *args, **kwargs)

# E2EError should be raised on any condition that should trigger a test failure
class E2EError(Exception):
    def __init__(self, *args, **kwargs):
        Exception.__init__(self, *args, **kwargs)

# wait_for waits for condition to be true. If condition does not hold
# immediately, it sleeps in increments of wait_incr until max_wait
# seconds. Its behavior is similar to Selenium's WebDriverWait and it's
# easier to use.
def wait_for(condition, max_wait=2, wait_incr=0.1):
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
    if not condition():
        raise E2EError("timed out waiting for condition")

# retry calls fn a maximum of $attempts times, waiting $cooldown seconds in between invocations.
# It returns the return value of the first invocation of fn that does not raise an exception.
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
# Clients should prefer to use the convenience methods whenever
# possible, but it is also acceptable to reference the underlying
# Selenium web driver directly.
class Driver(object):
    def __init__(self, wd, sourcegraph_url):
        self.d = wd  # d is the Selenium WebDriver that Driver uses to interact with the browser
        self.sourcegraph_url = sourcegraph_url

    def close(self):
        self.d.close()

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

    def find_search_modal_results(self, results_text):
        results = self.d.find_elements_by_css_selector("[data-class-name~=modal-result]")
        return [r for r in results if results_text in r.text]
