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

    def select_search_result_using_arrow_keys(result_text):
        for i in xrange(0, 2):
            for i in xrange(0, 20):
                def f():
                    selected = d.find_search_modal_selected_result()
                    if result_text == selected.text:
                        d.active_elem().send_keys(Keys.ENTER)
                        return True
                    return False
                if retry(f):
                    return
                d.active_elem().send_keys(Keys.DOWN)
            for i in xrange(0, 40): # network events might have changed the list, so try one more time from the top
                d.active_elem().send_keys(Keys.UP)
        raise E2EFatal("did not find search result '%s'" % result_text)

    wd.get(d.sg_url('/github.com/gorilla/mux')) # start on a page with the jump modal active
    for query, result_text, url_path in repo_queries:
        d.active_elem().send_keys("/")
        d.active_elem().send_keys(query)
        wait_for(d.all_network_indicators_are_invisible, max_wait=4)
        wait_for(lambda: len(d.find_search_modal_results(result_text, exact_match=True)) > 0)
        select_search_result_using_arrow_keys(result_text)
        wait_for(lambda: wd.current_url == d.sg_url(url_path), text=('wd.current_url == "%s"' % d.sg_url(url_path)))

def test_onboarding(d):
    wd = d.d
    username = os.getenv("NEW_USER_GITHUB")
    password = os.getenv("NEW_USER_GITHUB_PASSWORD", "")
    if username is None:
        logf("[%s] skipping test_onboarding because $NEW_USER_GITHUB not set" % yellow("WARN"))
        return
    d.delete_user_if_exists(username)

    wd.get(d.sg_url("/"))
    retry(lambda: wd.find_element_by_link_text("Sign up").click())
    retry(lambda: d.find_button_by_partial_text("Continue with GitHub").click())

    wd.find_element_by_id("login_field").send_keys(username)
    wd.find_element_by_id("password").send_keys(password)
    d.active_elem().send_keys(Keys.ENTER)

    if len(d.find_buttons_by_partial_text("Authorize application")) > 0:
        d.find_button_by_partial_text("Authorize application").click()

    retry(lambda: wd.find_element_by_link_text("Skip").click())
    wait_for(lambda: d.find_elements_by_tag_name_and_partial_text("div", "Find a repository"))

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

    # Search Google for "NewKMSKeyGenerator"
    wd.get("https://www.google.com/search?q=NewKMSKeyGenerator")

    # Click on Sourcegraph result
    retry(lambda: wd.find_element_by_partial_link_text("Sourcegraph").click())

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

all_tests = [
    test_onboarding,
    test_login_logout,
    test_repo_jump_to,
    test_golden_workflow,
]
