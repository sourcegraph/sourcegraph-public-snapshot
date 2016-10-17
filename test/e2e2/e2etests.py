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
        wait_for(d.all_network_indicators_are_invisible, max_wait=4)
        wait_for(lambda: len(d.find_search_modal_results(result_text, exact_match=True)) > 0)
        Util.select_search_result_using_arrow_keys(d, result_text, exact_match=True)
        wait_for(lambda: wd.current_url == d.sg_url(url_path), text=('wd.current_url == "%s"' % d.sg_url(url_path)))

def test_onboarding(d):
    wd = d.d
    username = os.getenv("NEW_USER_GITHUB")
    password = os.getenv("NEW_USER_GITHUB_PASSWORD", "")

    # Delete user from Auth0 if currently exists
    if username is None:
        logf("[%s] skipping test_onboarding because $NEW_USER_GITHUB not set" % yellow("WARN"))
        return
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

    # Skip the Chrome extension install
    retry(lambda: wd.find_element_by_link_text("Skip").click())
    wait_for(lambda: d.find_elements_by_tag_name_and_partial_text("div", "Find a repository"))

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

    # Get to NewKMSKeyGenerator landing page (presumably after
    # clicking on a Google search result).
    wd.get(d.sg_url("/github.com/aws/aws-sdk-go/-/info/GoPackage/github.com/aws/aws-sdk-go/service/s3/s3crypto/-/NewKMSKeyGenerator"))

    # Click on first usage example
    retry(lambda: wd.find_element_by_partial_link_text("handler := s3crypto.NewKMSKeyGenerator").click())

    # Hover over "NewEncryptionClient" token
    wait_for(lambda: len(d.find_tokens("NewEncryptionClient")) > 0)
    d.hover_token("NewEncryptionClient")
    wait_for(lambda: '' in d.find_tooltip_near_elem(d.find_tokens("NewEncryptionClient")[0]).text)

    # Right click and peek
    retry(lambda: d.right_click(d.find_token("NewEncryptionClient")))
    retry(lambda: d.find_context_menu_option("Peek Definition").click())

    # Click "NewEncryptionClient" token
    retry(lambda: d.find_token("NewEncryptionClient").click())

    # Dismiss peek view
    retry(lambda: d.active_elem().send_keys(Keys.ESCAPE))

    # Right click and find refs
    def rc():
        retry(lambda: d.right_click(d.find_token("NewEncryptionClient")))
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

    # Jump to modal to "NewEncryptionClient"
    retry(lambda: d.active_elem().send_keys("/"))
    retry(lambda: d.active_elem().send_keys("NewEncryptionClient"))
    wait_for(d.all_network_indicators_are_invisible, max_wait=4)
    retry(lambda: d.active_elem().send_keys(Keys.ENTER))
    wait_for(lambda: "encryption_client.go" in wd.current_url and "/github.com/aws/aws-sdk-go" in wd.current_url)

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
        "repo": "github.com/aws/aws-sdk-go",
        "symbol": "s3crypto.NewKMSKeyGenerator",
        "symbol_name": "NewKMSKeyGenerator",
    }]
    for test in tests:
        # Go to repo page
        wd.get(d.sg_url("/%s" % test['repo']))
        wait_for(lambda: len(d.find_elements_by_tag_name_and_partial_text("div", "FILES")) > 0)

        # Jump to symbol
        d.active_elem().send_keys("/")
        d.active_elem().send_keys(test['symbol_name'])
        wait_for(d.all_network_indicators_are_invisible, max_wait=4)
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

        # Definition header
        wait_for(lambda: len(d.find_elements_by_tag_name_and_partial_text("h2", "Definition")) > 0)

def test_beta_signup(d):
    wd = d.d
    username = os.getenv("USER_GITHUB")
    password = os.getenv("USER_GITHUB_PASSWORD")

    wd.get(d.sg_url("/"))
    Util.log_in(d, username, password)

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

all_tests = [
    test_onboarding,
    test_login_logout,
    test_repo_jump_to,
    test_golden_workflow,
    test_find_external_refs,
    test_beta_signup,
]
