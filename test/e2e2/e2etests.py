import time
import math
import traceback
import argparse
import os
import sys

from selenium.webdriver.common.keys import Keys
from colors import yellow

from e2etypes import *

def test_repo_jump_to(d):
    wd = d.d

    repo_queries = [
        ("golang/go", "github.com/golang/go", "/github.com/golang/go"),
        ("mux", "github.com/gorilla/mux", "/github.com/gorilla/mux"),
        ("pat", "github.com/gorilla/pat", "/github.com/gorilla/pat"),
    ]

    wd.get(d.sg_url('/github.com/gorilla/mux')) # start on a page with the jump modal active
    for query, result_text, url_path in repo_queries:
        d.active_elem().send_keys("/")
        d.active_elem().send_keys(query)
        Util.wait_for_all_network_indicators_to_be_invisible_with_jiggle(d, jiggle_wait=5)
        wait_for(lambda: len(d.find_search_modal_results(result_text, exact_match=True)) > 0)
        Util.select_search_result_using_arrow_keys(d, result_text, exact_match=True)
        wait_for(lambda: wd.current_url == d.sg_url(url_path), text=('wd.current_url == "%s"' % d.sg_url(url_path)))

def test_github_private_auth_onboarding(d):
    wd = d.d
    username = os.getenv("NEW_USER_GITHUB")
    password = os.getenv("NEW_USER_GITHUB_PASSWORD", "")
    if username is None:
        logf("[%s] skipping test_onboarding because $NEW_USER_GITHUB not set" % yellow("WARN"))
        return

    # Delete user from Auth0 if currently exists
    d.delete_user_if_exists(username)

    # Go to home, click "Sign up"
    wd.get(d.sg_url("/"))
    retry(lambda: wd.find_element_by_link_text("Sign up").click())
    retry(lambda: d.find_button_by_partial_text("Continue with GitHub").click())

    # Type in GitHub login creds
    wd.find_element_by_id("login_field").send_keys(username)
    wd.find_element_by_id("password").send_keys(password)
    d.active_elem().send_keys(Keys.ENTER)

    # Re-authorize application in case GitHub thinks we're a bot (heh heh)
    if len(d.find_buttons_by_partial_text("Authorize application")) > 0:
        d.find_button_by_partial_text("Authorize application").click()

    wait_for(lambda: len(d.find_tokens("Checkers")) > 0)
    wait_for(lambda: wd.find_element_by_id("def-coachmark"), max_wait=4.0)

    # Log out
    Util.log_out(d)

def test_github_public_auth_onboarding(d):
    wd = d.d
    username = os.getenv("NEW_USER_PUBLIC_GITHUB")
    password = os.getenv("NEW_USER_PUBLIC_GITHUB_PASSWORD", "")
    if username is None:
        logf("[%s] skipping test_github_public_auth_onboarding because $NEW_USER_PUBLIC_GITHUB not set" % yellow("WARN"))
        return

    # Delete user from Auth0 if currently exists
    d.delete_user_if_exists(username)

    # Go to home, click "Sign up"
    wd.get(d.sg_url("/"))
    retry(lambda: wd.find_element_by_link_text("Sign up").click())
    retry(lambda: wd.find_element_by_link_text("sign in").click())

    # Type in GitHub login creds
    wd.find_element_by_id("login_field").send_keys(username)
    wd.find_element_by_id("password").send_keys(password)
    d.active_elem().send_keys(Keys.ENTER)

    # Re-authorize application in case GitHub thinks we're a bot (heh heh)
    if len(d.find_buttons_by_partial_text("Authorize application")) > 0:
        d.find_button_by_partial_text("Authorize application").click()

    wait_for(lambda: len(d.find_tokens("Checkers")) > 0)
    wait_for(lambda: wd.find_element_by_id("def-coachmark"), max_wait=4.0)
    wait_for(lambda: wd.find_element_by_id("ref-coachmark"), max_wait=4.0)

    # Log out
    Util.log_out(d)

def test_login_logout(d):
    wd = d.d
    username = os.getenv("USER_GITHUB")
    password = os.getenv("USER_GITHUB_PASSWORD")
    if username is None:
        logf("[%s] skipping test_login_logout because $USER_GITHUB not set" % yellow("WARN"))
        return
    wd.get(d.sg_url("/"))
    Util.log_in(d, username, password)
    wait_for(lambda: wd.current_url == d.sg_url("/"))
    Util.log_out(d)
    wait_for(lambda: wd.current_url == d.sg_url("/"))

def test_golden_workflow(d):
    wd = d.d

    # Get to NewRouter landing page (presumably after
    # clicking on a Google search result).
    wd.get(d.sg_url("/github.com/gorilla/mux/-/info/GoPackage/github.com/gorilla/mux/-/NewRouter"))

    # Click on first usage example
    retry(lambda: wd.find_element_by_partial_link_text("router := mux.NewRouter()").click())

    # Hover over "NewRouter" token
    wait_for(lambda: len(d.find_tokens("NewRouter")) > 0)
    d.hover_token("NewRouter")
    wait_for(lambda: '' in d.find_tooltip_near_elem(d.find_tokens("NewRouter")[0]).text)

    # Right click and peek
    retry(lambda: d.right_click(d.find_token("NewRouter")))
    retry(lambda: d.find_context_menu_option("Peek Definition").click())

    # Click "NewRouter" token
    retry(lambda: d.find_token("NewRouter").click())

    # Dismiss peek view
    retry(lambda: d.active_elem().send_keys(Keys.ESCAPE))

    # Right click and find refs
    def rc():
        retry(lambda: d.right_click(d.find_token("NewRouter")))
        retry(lambda: d.find_context_menu_option("Find All References").click())
        wait_for(lambda: len(d.find_references_menu_options()) > 0, 10)
    retry(rc)

    # Peek reference
    retry(lambda: d.find_references_menu_options()[0].click())

    # Jump to reference
    wait_for(lambda: len(d.find_references_menu_options()) > 0)
    retry(lambda: d.double_click(d.find_references_menu_options()[0]))

    # Dismiss peek view
    retry(lambda: d.active_elem().send_keys(Keys.ESCAPE))

    # Jump to modal to "NewRouter"
    retry(lambda: d.active_elem().send_keys("/"))
    retry(lambda: d.active_elem().send_keys("NewRouter"))
    Util.wait_for_all_network_indicators_to_be_invisible_with_jiggle(d, jiggle_wait=4)
    def e():
        d.active_elem().send_keys(Keys.ENTER)
        wait_for(lambda: "mux.go" in wd.current_url and "/github.com/gorilla/mux" in wd.current_url)
    retry(e)

def test_find_external_refs(d):
    wd = d.d

    tests = [{
        "repo": "github.com/go-gorp/gorp",
        "symbol": "gorp.SqlExecutor",
        "symbol_name": "SqlExecutor",
    }, {
        "repo": "github.com/gorilla/mux",
        "symbol": "mux.Router",
        "symbol_name": "Router",
    }, {
        "repo": "github.com/golang/go",
        "symbol": "json.Marshal",
        "symbol_name": "Marshal",
    }]
    for test in tests:
        # Go to repo page
        wd.get(d.sg_url("/%s" % test['repo']))
        wait_for(lambda: len(d.find_elements_by_tag_name_and_partial_text("div", "FILES")) > 0)

        # Jump to symbol
        d.active_elem().send_keys("/")
        d.active_elem().send_keys(test['symbol'])
        Util.wait_for_all_network_indicators_to_be_invisible_with_jiggle(d, jiggle_wait=4)
        wait_for(lambda: len(d.find_search_modal_results(test['symbol'])) > 0)
        Util.select_search_result_using_arrow_keys(d, test['symbol'])

        # Right click, find external references
        wait_for(lambda: len(d.find_tokens(test['symbol_name'])) > 0, 5) # wait a little longer, to rule out VSCode start-up time
        def rc():
            retry(lambda: d.active_elem().send_keys(Keys.ESCAPE)) # hide any tooltip that might steal the click
            retry(lambda: d.active_elem().send_keys(Keys.UP)) # cursor might steal the click if we don't move it out of the way
            retry(lambda: d.right_click(d.find_token(test['symbol_name'])))
            retry(lambda: d.find_context_menu_option("Find External References").click())
        retry(rc)

        # Description header
        wait_for(lambda: len(d.find_elements_by_tag_name_and_partial_text("h2", "Description")) > 0)

        # Examples header
        wait_for(lambda: len(d.find_elements_by_tag_name_and_partial_text("h2", "Examples")) > 0)

        # External repository references header
        wait_for(lambda: len(d.find_elements_by_tag_name_and_partial_text("h2", "%s is referenced in" % test['symbol_name'])) > 0)

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

    retry(lambda: wd.find_element_by_css_selector('[class^="BetaInterestForm"] input').send_keys("Bobby Jones"))
    wait_for(lambda: wd.find_element_by_css_selector('[class^="BetaInterestForm"] input').get_attribute("value") == "Bobby Jones")

    def f():
        checkboxes = wd.find_elements_by_css_selector('[class^="BetaInterestForm"] input[type="checkbox"]')
        for checkbox in checkboxes:
            checkbox.click()
    retry(f)

    retry(lambda: wd.find_element_by_css_selector('[class^="BetaInterestForm"] textarea').send_keys("Sourcegraph is great"))
    wait_for(lambda: wd.find_element_by_css_selector('[class^="BetaInterestForm"] textarea').get_attribute("value") == "Sourcegraph is great")
    retry(lambda: d.find_button_by_partial_text("Participate").click())
    wait_for(lambda: len(d.find_elements_by_tag_name_and_partial_text("p", "We'll contact you at %s" % username)) > 0)

def test_first_open_jump_to_line(d):
    wd = d.d
    wd.get(d.sg_url("/github.com/gorilla/pat/-/blob/pat.go#L65:18"))
    wait_for(lambda: len([e for e in wd.find_elements_by_css_selector(".line-numbers") if e.text == "65"]) == 1)

def test_browser_extension_app_injection(d):
    wd = d.d
    wd.get("https://github.com/gorilla/mux")
    wait_for(lambda: len(wd.find_elements_by_id("sourcegraph-app-bootstrap")) == 1)
    wait_for(lambda: wd.find_element_by_id("sourcegraph-app-bootstrap").value_of_css_property("display") == "none")
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
    d.verify_new_tab_opened("https://sourcegraph.com/github.com/gorilla/mux@757bef944d0f21880861c2dd9c871ca543023cba/-/blob/mux.go#L17")

def test_browser_extension_hover_j2d_unified_pull_request(d):
    wd = d.d
    wd.get("https://github.com/gorilla/mux/pull/205/files?diff=unified")
    wait_for(lambda: len(wd.find_elements_by_class_name("sourcegraph-app-annotator")) == 2)

    tests = [{
        # addition
        "node": "text-node-17-5",
        "hover": "var contextSet func(r *Request, key interface{}, val interface{}) *Request",
        "j2d_location": "https://sourcegraph.com/github.com/captncraig/mux@acfc892941192f90aadd4f452a295bf39fc5f7ed/-/blob/mux.go#L17",
    }, {
        # deletion
        "node": "text-node-88-3",
        "hover": "func setVars(r *Request, val interface{})",
        "j2d_location": "https://sourcegraph.com/github.com/gorilla/mux@9c068cf16d982f8bd444b8c352acbeec34c4fe5b/-/blob/mux.go#L326",
    }, {
        # unmodified
        "node": "text-node-24-6",
        "hover": "func NewRouter() *Router\nNewRouter returns a new router instance.",
        "j2d_location": "https://sourcegraph.com/github.com/captncraig/mux@acfc892941192f90aadd4f452a295bf39fc5f7ed/-/blob/mux.go#L24",
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
        "j2d_location": "https://sourcegraph.com/github.com/captncraig/mux@acfc892941192f90aadd4f452a295bf39fc5f7ed/-/blob/mux.go#L17",
    }, {
        # deletion
        "node": "text-node-88-3",
        "hover": "func setVars(r *Request, val interface{})",
        "j2d_location": "https://sourcegraph.com/github.com/gorilla/mux@9c068cf16d982f8bd444b8c352acbeec34c4fe5b/-/blob/mux.go#L326",
    }, {
        # unmodified
        "node": "text-node-18-6",
        "hover": "func NewRouter() *Router\nNewRouter returns a new router instance.",
        "j2d_location": "https://sourcegraph.com/github.com/gorilla/mux@9c068cf16d982f8bd444b8c352acbeec34c4fe5b/-/blob/mux.go#L18",
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

all_tests = [
    # (test_github_private_auth_onboarding, "@kingy"), # TODO(king): re-enable after flakiness fixed
    # (test_github_public_auth_onboarding, "@kingy"), # TODO(king): re-enable after flakiness fixed
    (test_login_logout, "@beyang"),
    (test_repo_jump_to, "@nico"),
    (test_golden_workflow, "@matt"),
    (test_find_external_refs, "@stephen"),
    (test_beta_signup, "@kingy"),
    (test_first_open_jump_to_line, "@nico"),
    (test_browser_extension_app_injection, "@john"),
    (test_browser_extension_hover_j2d_blob, "@john"),
    (test_browser_extension_hover_j2d_unified_pull_request, "@john"),
    (test_browser_extension_hover_j2d_split_pull_request, "@john"),
]
