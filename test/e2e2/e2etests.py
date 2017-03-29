import time
import math
import traceback
import argparse
import os
import sys
import requests

from selenium.webdriver.common.keys import Keys
from colors import yellow

from e2etypes import *

def test_direct_link_to_repo(d):
    wd = d.d

    wd.get(d.sg_url("/github.com/gorilla/muxy@65b4fd5d316b4b260db61f66726e4859fd0e4889"))
    wait_for(lambda: wd.find_element_by_id("directory_help_message"))
    wait_for(lambda: len(wd.find_elements_by_class_name("monaco-tree-row")) == 10, 5.0)

def test_direct_link_to_directory(d):
    wd = d.d

    wd.get(d.sg_url("/github.com/gorilla/muxy@65b4fd5d316b4b260db61f66726e4859fd0e4889/-/tree/encoder"))
    wait_for(lambda: wd.find_element_by_id("directory_help_message"))
    wait_for(lambda: len(wd.find_elements_by_class_name("monaco-tree-row")) == 11, 5.0)

def test_repo_jump_to(d):
    wd = d.d

    wd.get(d.sg_url('/github.com/gorilla/mux')) # start on a page with the jump modal active
    wait_for(lambda: wd.find_element_by_id("directory_help_message"))
    d.send_keys_like_human("/")
    send_keys_with_retry(d.active_elem, "!golang/go",
                         lambda: len(d.find_search_modal_results("go\ngolang", exact_match=True)) > 0,
                         max_wait=5.0)
    d.find_search_modal_results("go\ngolang", exact_match=True)[0].click()

    wait_for(lambda: wd.current_url == d.sg_url("/github.com/golang/go"), text=('wd.current_url == "%s"' % d.sg_url("/github.com/gorilla/mux")))

def test_login_logout(d):
    wd = d.d
    username = os.getenv("USER_GITHUB")
    password = os.getenv("USER_GITHUB_PASSWORD")
    if username is None:
        logf("[%s] skipping test_login_logout because $USER_GITHUB not set" % yellow("WARN"))
        return
    wd.get(d.sg_url("/"))
    Util.log_in(d, username, password)
    Util.log_out(d)
    wait_for(lambda: wd.current_url == d.sg_url("/"))

def test_godoc_workflow(d):
    wd = d.d

    # Get to NewRouter landing page (presumably after
    # clicking on a godoc link or Google search result).
    wd.get(d.sg_url("/go/github.com/gorilla/mux/-/NewRouter"))

    # Hover over "NewRouter" token
    wait_for(lambda: len(d.find_tokens("NewRouter")) > 0, 10)
    d.hover_token_with_retry("NewRouter",
                             lambda: 'NewRouter' in d.find_tooltip_near_elem(d.find_tokens("NewRouter")[0])[1].text)

    # Open NewRouter in InfoBar
    retry(lambda: d.find_token("NewRouter").click())

    # Wait until local and global refs load.
    wait_for(lambda: len(wd.find_elements_by_id("reference-tree")) == 1, 15)
    wait_for(lambda: len(wd.find_elements_by_class_name("monaco-tree-rows")) > 0)
    wait_for(lambda: len(wd.find_elements_by_class_name("left-right-widget_right")) > 0)
    wait_for(lambda: len(wd.find_elements_by_class_name("uil-default")) == 0, 45)
    wait_for(lambda: len(wd.find_elements_by_class_name("monaco-workspace-badge")) >= 1)
    retry(lambda: wd.find_element_by_class_name("monaco-workspace-badge").click())
    wait_for(lambda: len(wd.find_elements_by_class_name("monaco-workspace-badge")) >= 2)

    # Open preview and scroll downlist without affecting InfoBar
    retry(lambda: wd.find_element_by_class_name("monaco-workspace-badge").click())
    retry(lambda: d.send_keys_like_human(Keys.RIGHT))
    retry(lambda: d.send_keys_like_human(Keys.DOWN))

    # Verify Infobar remained open
    wait_for(lambda: len(wd.find_elements_by_class_name("monaco-tree-rows")) > 0)
    wait_for(lambda: len(wd.find_elements_by_class_name("left-right-widget_right")) > 0)

    # Dismiss InfoBar
    retry(lambda: d.send_keys_like_human(Keys.ESCAPE)) # hide any tooltip that might steal the click

    # Quickopen to "Route"
    retry(lambda: d.send_keys_like_human("/"))
    send_keys_with_retry(d.active_elem, "#Route",
                         lambda: len(d.find_search_modal_results("Route\nmux", exact_match=True)) > 0,
                         max_wait=5.0)

    retry(lambda: d.send_keys_like_human(Keys.ENTER))
    wait_for(lambda: "route.go" in wd.current_url and "/github.com/gorilla/mux" in wd.current_url)

def test_global_refs(d, test):
    wd = d.d

    # Go to repo page
    wd.get(d.sg_url("/%s" % test['repo_rev']))
    wait_for(lambda: wd.find_element_by_id("directory_help_message"))

    # Jump to symbol
    d.send_keys_like_human("/")
    send_keys_with_retry(d.active_elem, '#' + test['symbol'],
                         lambda: len(d.find_search_modal_results(test['symbol'])) > 0,
                         max_wait=5.0)
    d.find_search_modal_results(test['symbol'])[0].click()

    # Wait for sidebar to appear.
    wait_for(lambda: len(wd.find_elements_by_css_selector('[class="sg-sidebar"]')) > 0)

    # Symbol signature
    wait_for(lambda: len(d.find_sidebar_elements_by_tag_name_and_partial_text("div", test["symbol"])) > 0)

    # "Defined in" header
    wait_for(lambda: len(d.find_sidebar_elements_by_tag_name_and_partial_text("p", "Defined in")) > 0)

    # Wait for references to load + un-expand the "Local" references
    wait_for(lambda: len(wd.find_elements_by_id("reference-tree")) == 1, 15)
    wait_for(lambda: len(wd.find_elements_by_class_name("monaco-tree-rows")) > 0)
    wait_for(lambda: len(wd.find_elements_by_class_name("left-right-widget_right")) > 0)
    wait_for(lambda: len(wd.find_elements_by_class_name("uil-default")) == 0, 45) # Wait for loading icon to disappear
    wait_for(lambda: len(wd.find_elements_by_class_name("monaco-workspace-badge")) >= 1)
    retry(lambda: wd.find_element_by_class_name("monaco-workspace-badge").click())

    # Local References
    wait_for(lambda: len(d.find_sidebar_elements_by_tag_name_and_partial_text("div", "Local")) > 0)

    # External References
    wait_for(lambda: len(d.find_sidebar_elements_by_tag_name_and_partial_text("div", "External")) > test["global_min"])

def test_beta_signup(d):
    wd = d.d
    username = os.getenv("USER_GITHUB")
    password = os.getenv("USER_GITHUB_PASSWORD")
    if username is None:
        logf("[%s] skipping test_beta_signup because $USER_GITHUB not set" % yellow("WARN"))
        return

    wd.get(d.sg_url("/"))
    Util.log_in(d, username, password)
    wait_for(lambda: len(d.find_elements_by_tag_name_and_partial_text("div", "My repositories")) > 0)

    wd.get(d.sg_url("/beta"))
    wait_for(lambda: len(d.find_elements_by_tag_name_and_partial_text("div", "Register for beta access")) > 0)

    retry(lambda: wd.execute_script("return arguments[0].scrollIntoView();", d.find_button_by_partial_text("Participate")))

    retry(lambda: wd.find_element_by_name('fullName').send_keys("Bobby Jones"))
    wait_for(lambda: wd.find_element_by_name('fullName').get_attribute("value") == "Bobby Jones")
    retry(lambda: wd.find_element_by_name('company').send_keys("Dutch East India"))
    wait_for(lambda: wd.find_element_by_name('company').get_attribute("value") == "Dutch East India")

    def f():
        checkboxes = wd.find_elements_by_name('editors')
        for checkbox in checkboxes:
            checkbox.click()
    retry(f)

    def f2():
        checkboxes = wd.find_elements_by_name('languages')
        for checkbox in checkboxes:
            checkbox.click()
    retry(f2)

    retry(lambda: wd.find_element_by_name('message').send_keys("Sourcegraph is great"))
    wait_for(lambda: wd.find_element_by_name('message').get_attribute("value") == "Sourcegraph is great")
    retry(lambda: d.find_button_by_partial_text("Participate").click())
    wait_for(lambda: len(d.find_elements_by_tag_name_and_partial_text("p", "We'll contact you at")) > 0)

def test_first_open_jump_to_line(d):
    wd = d.d
    wd.get(d.sg_url("/github.com/gorilla/pat/-/blob/pat.go#L65:18"))
    wait_for(lambda: len([e for e in wd.find_elements_by_css_selector(".line-numbers") if e.text == "65"]) == 1)

def test_browser_extension_app_injection(d):
    wd = d.d
    wd.get("https://github.com/gorilla/mux")
    wait_for(lambda: len(wd.find_elements_by_id("sourcegraph-app-background")) == 1)
    wait_for(lambda: wd.find_element_by_id("sourcegraph-app-background").value_of_css_property("display") == "none")

def test_browser_extension_hover_j2d_blob(d):
    wd = d.d
    wd.get("https://github.com/gorilla/mux/blob/757bef944d0f21880861c2dd9c871ca543023cba/mux.go")
    wait_for(lambda: len(wd.find_elements_by_class_name("sourcegraph-app-annotator")) == 1)

    # hover over a token, get a tooltip (may be "Loading...")
    wait_for(lambda: len(wd.find_elements_by_id("text-node-17-6")) == 1)
    retry(lambda: d.hover_elem(wd.find_element_by_id("text-node-17-6")))
    wait_for(lambda: len(wd.find_elements_by_class_name("sg-tooltip")) == 1)

    # wait for the token to be clickable (textDocument/defnition is resolved)
    wait_for(lambda: len([e for e in wd.find_elements_by_class_name("sg-tooltip") if e.text == "func NewRouter() *Router\nNewRouter returns a new router instance."]) == 1, 10)

    # click and wait for page navigation
    retry(lambda: wd.find_element_by_id("text-node-17-6").click())
    d.verify_new_tab_opened("https://sourcegraph.com/github.com/gorilla/mux@757bef944d0f21880861c2dd9c871ca543023cba/-/blob/mux.go#L17:6-17:15")

def test_browser_extension_hover_j2d_unified_pull_request(d):
    wd = d.d
    wd.get("https://github.com/gorilla/mux/pull/205/files?diff=unified")
    wait_for(lambda: len(wd.find_elements_by_class_name("sourcegraph-app-annotator")) == 2)

    tests = [{
        # addition
        "node": "text-node-17-5",
        "hover": "var contextSet func(r *Request, key interface{}, val interface{}) *Request",
        "j2d_location": "https://sourcegraph.com/github.com/captncraig/mux@acfc892941192f90aadd4f452a295bf39fc5f7ed/-/blob/mux.go#L17:5-17:15",
    }, {
        # deletion
        "node": "text-node-88-3",
        "hover": "func setVars(r *Request, val interface{})",
        "j2d_location": "https://sourcegraph.com/github.com/gorilla/mux@9c068cf16d982f8bd444b8c352acbeec34c4fe5b/-/blob/mux.go#L326:6-326:13",
    }, {
        # unmodified
        "node": "text-node-24-6",
        "hover": "func NewRouter() *Router\nNewRouter returns a new router instance.",
        "j2d_location": "https://sourcegraph.com/github.com/captncraig/mux@acfc892941192f90aadd4f452a295bf39fc5f7ed/-/blob/mux.go#L24:6-24:15",
    }]
    for test in tests:
        # hover over a token, get a tooltip (may be "Loading...")
        wait_for(lambda: len(wd.find_elements_by_id(test["node"])) == 1)
        retry(lambda: d.hover_elem(wd.find_element_by_id(test["node"])))
        wait_for(lambda: len(wd.find_elements_by_class_name("sg-tooltip")) == 1)

        # wait for the token to be clickable (textDocument/defnition is resolved)
        wait_for(lambda: len([e for e in wd.find_elements_by_class_name("sg-tooltip") if e.text == test["hover"]]) == 1, 10)

        # click and wait for page navigation
        retry(lambda: wd.find_element_by_id(test["node"]).click())
        d.verify_new_tab_opened(test["j2d_location"])

        # refresh location after j2d for next test
        wd.get("https://github.com/gorilla/mux/pull/205/files?diff=unified")

def test_browser_extension_hover_j2d_split_pull_request(d):
    wd = d.d
    wd.get("https://github.com/gorilla/mux/pull/205/files?diff=split")
    wait_for(lambda: len(wd.find_elements_by_class_name("sourcegraph-app-annotator")) == 2)
    tests = [{
        # addition
        "node": "text-node-17-5",
        "hover": "var contextSet func(r *Request, key interface{}, val interface{}) *Request",
        "j2d_location": "https://sourcegraph.com/github.com/captncraig/mux@acfc892941192f90aadd4f452a295bf39fc5f7ed/-/blob/mux.go#L17:5-17:15",
    }, {
        # deletion
        "node": "text-node-88-3",
        "hover": "func setVars(r *Request, val interface{})",
        "j2d_location": "https://sourcegraph.com/github.com/gorilla/mux@9c068cf16d982f8bd444b8c352acbeec34c4fe5b/-/blob/mux.go#L326:6-326:13",
    }, {
        # unmodified
        "node": "text-node-18-6",
        "hover": "func NewRouter() *Router\nNewRouter returns a new router instance.",
        "j2d_location": "https://sourcegraph.com/github.com/gorilla/mux@9c068cf16d982f8bd444b8c352acbeec34c4fe5b/-/blob/mux.go#L18:6-18:15",
    }]
    for test in tests:
        # hover over a token, get a tooltip (may be "Loading...")
        wait_for(lambda: len(wd.find_elements_by_id(test["node"])) == 1)
        retry(lambda: d.hover_elem(wd.find_element_by_id(test["node"])))
        wait_for(lambda: len(wd.find_elements_by_class_name("sg-tooltip")) == 1)

        # wait for the token to be clickable (textDocument/defnition is resolved)
        wait_for(lambda: len([e for e in wd.find_elements_by_class_name("sg-tooltip") if e.text == test["hover"]]) == 1, 10)

        # click and wait for page navigation
        retry(lambda: wd.find_element_by_id(test["node"]).click())
        d.verify_new_tab_opened(test["j2d_location"])

        # refresh location after j2d for next test
        wd.get("https://github.com/gorilla/mux/pull/205/files?diff=split")

def test_java_symbol(dr):
    wd = dr.d
    # Go to JUnit repo page
    wd.get(dr.sg_url("/github.com/junit-team/junit4"))
    wait_for(lambda: wd.find_element_by_id("directory_help_message"), 5)
    wait_for(lambda: len(wd.find_elements_by_class_name("monaco-tree-row")) > 0, 5)

    # Symbol search for "testfailure"
    dr.send_keys_like_human("/")
    send_keys_with_retry(dr.active_elem, "#testfailure",
                         lambda: len(dr.find_search_modal_results("TestFailure\njunit.framework", exact_match=True)) > 0,
                         max_wait=5.0)
    wait_for(lambda: len(dr.find_search_modal_results("TestFailure\njunit.framework", exact_match=True)) > 0, 30.0)
    # Click on "TestFailure junit.framework"
    dr.find_search_modal_results("TestFailure\njunit.framework", exact_match=True)[0].click()
    # Wait for URL to change
    wait_for(lambda: "/TestFailure.java#" in wd.current_url, max_wait=10.0, text=('file is TestFailure.java'))
    # Verify info bar is opened
    wait_for(lambda: len(wd.find_elements_by_id("reference-tree")) == 1, 15)
    wait_for(lambda: len(wd.find_elements_by_class_name("monaco-tree-rows")) > 1)
    wait_for(lambda: len(wd.find_elements_by_class_name("left-right-widget_right")) > 0)

def test_java_hover(dr):
    wd = dr.d
    # Go to JUnit repo page
    wd.get(dr.sg_url("/github.com/junit-team/junit4/-/blob/src/main/java/junit/framework/TestFailure.java"))
    # Hover over "TestFailure" token
    wait_for(lambda: len(dr.find_tokens(" TestFailure ")) > 0, 10)
    retry(lambda: dr.hover_token(" TestFailure"))
    wait_for(lambda: 'TestFailure' in dr.find_tooltip_near_elem(dr.find_tokens(" TestFailure ")[0])[1].text)

def test_java_def(dr):
    wd = dr.d
    # Go to JUnit repo page
    wd.get(dr.sg_url("/github.com/junit-team/junit4/-/blob/src/main/java/junit/framework/TestFailure.java"))
    # Click "TestFailure" token and wait until side panel loaded.
    wait_for(lambda: len(dr.find_tokens(" TestFailure ")) > 0, 10)
    click_with_retry(lambda: dr.find_token(" TestFailure "),
                     lambda: len(wd.find_elements_by_id("reference-tree")) == 1, max_wait=15)
    # Click "Throwables" token and wait until side panel reloaded.
    wait_for(lambda: len(dr.find_tokens("Throwables")) > 0, 10)
    click_with_retry(lambda: dr.find_token("Throwables"),
                     lambda: 'Throwables' in wd.find_elements_by_id("reference-tree")[0].text,
                     max_wait=15)
    # Click "Jump to definition"
    dr.find_jump_to_definition_button().click()
    # Wait for URL to change
    wait_for(lambda: "/Throwables.java#" in wd.current_url, max_wait=10.0, text=('file is Throwables.java'))
    # Check if page properly loaded
    wait_for(lambda: len(dr.find_tokens("Throwables")) > 0, 10)

def test_java_cross_repo(dr):
    wd = dr.d
    # Go to JUnit repo page
    wd.get(dr.sg_url("/github.com/google/guava/-/blob/guava/src/com/google/common/collect/RangeGwtSerializationDependencies.java"))
    # Wait for page to load
    wait_for(lambda: len(dr.find_tokens("collect")) > 0, max_wait=10, text="wait for page load")
    # Click in editor
    wd.find_elements_by_css_selector(".monaco-editor")[0].click()
    # Click "Serializable" and wait until side panel reloaded
    click_with_retry(lambda: dr.find_token("Serializable"),
                     lambda: 'Serializable' in wd.find_elements_by_id("reference-tree")[0].text,
                     max_wait=15)
    # Click "Jump to definition"
    dr.find_jump_to_definition_button().click()
    # Wait for URL to change
    wait_for(lambda: "/Serializable.java#" in wd.current_url, max_wait=10.0, text=('file is Serializable.java'))
    # Check if page properly loaded
    wait_for(lambda: len(dr.find_tokens(" Serializable ")) > 0, 10)

def test_java_global_usages(dr):
    wd = dr.d
    # Go to JUnit repo page
    wd.get(dr.sg_url("/github.com/junit-team/junit4/-/blob/src/test/java/junit/tests/AllTests.java"))
    # Wait for page to load
    wait_for(lambda: len(dr.find_tokens("")) > 0, max_wait=10, text="wait for page load")
    click_with_retry(lambda: dr.find_token(" Test "),
                     lambda: 'Test' in wd.find_elements_by_id("reference-tree")[0].text,
                     max_wait=15)
    # Wait for references to load + un-expand the "Local" references
    wait_for(lambda: len(wd.find_elements_by_id("reference-tree")) == 1, 15)
    wait_for(lambda: len(wd.find_elements_by_class_name("monaco-tree-rows")) > 0)
    wait_for(lambda: len(wd.find_elements_by_class_name("left-right-widget_right")) > 0)
    wait_for(lambda: len(wd.find_elements_by_class_name("uil-default")) == 0, 45) # Wait for loading icon to disappear
    wait_for(lambda: len(wd.find_elements_by_class_name("monaco-workspace-badge")) >= 1)
    retry(lambda: wd.find_element_by_class_name("monaco-workspace-badge").click())
    # Local References
    wait_for(lambda: len(dr.find_sidebar_elements_by_tag_name_and_partial_text("div", "Local")) > 0)
    # External References
    wait_for(lambda: len(dr.find_sidebar_elements_by_tag_name_and_partial_text("div", "External")) > 0)

def ensure_test_data(sourcegraph_url):
    print("Ensuring test repositories are cloned")
    start = time.time()
    while True:
        req = requests.post("%s/.api/repos-ensure" % sourcegraph_url, json=test_repos)
        if len(req.json()) == 0:
            break
        if time.time() - start > 60:
            raise Exception("timed out waiting for test data to be ensured")
        time.sleep(1)
    print("All test repositories are cloned")

test_repos = [
    "github.com/gorilla/muxy",
    "github.com/gorilla/mux",
    "github.com/golang/go",
    "github.com/gorilla/pat",
    "github.com/captncraig/mux",
    "github.com/junit-team/junit4",
    "github.com/google/guava",
];

all_tests = [
    (test_login_logout, "@kingy"),
    (test_repo_jump_to, "@john"),
    (test_godoc_workflow, "@john"),
    (test_direct_link_to_repo, "@nick"),
    (test_direct_link_to_directory, "@nick"),
    (test_beta_signup, "@kingy"),
    (test_first_open_jump_to_line, "@nico"),
    (test_browser_extension_app_injection, "@john"),
    (test_browser_extension_hover_j2d_blob, "@john"),
    (test_browser_extension_hover_j2d_unified_pull_request, "@john"),
    (test_browser_extension_hover_j2d_split_pull_request, "@john"),
    (test_java_symbol, "@the.other.aaron"),
    (test_java_hover, "@the.other.aaron"),
    (test_java_def, "@the.other.aaron"),
    (test_java_cross_repo, "@the.other.aaron"),
    # (test_java_global_usages, "@the.other.aaron"), # broken
]

global_ref_tests = [{
    "repo_rev": "github.com/golang/go@go1.7.3", # non-default branch
    "symbol": "Context",
    "global_min": 5,
}, {
    "repo_rev": "github.com/gorilla/mux",
    "symbol": "Router",
    "global_min": 5,
}]
for test in global_ref_tests:
    def test_global_refs_wrap(d):
	return test_global_refs(d, test)
    test_global_refs_wrap.func_name = test_global_refs.func_name + '_' + test['symbol'].lower()
    all_tests.append((test_global_refs_wrap, '@stephen'))
