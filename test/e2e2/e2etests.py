import time
import math
import traceback
import argparse
import os
import sys

from selenium.webdriver.common.keys import Keys

from e2etypes import *

def test_repo_jump_to(d):
    wd = d.d

    repo_queries = [
        ("golang/go", "github.com/golang/go", "/github.com/golang/go"),
        ("mux", "github.com/gorilla/mux", "/github.com/gorilla/mux"),
        ("pat", "github.com/gorilla/pat", "/github.com/gorilla/pat"),
    ]

    def select_search_result_using_arrow_keys(result_text):
        for i in xrange(0, 30):
            d.active_elem().send_keys(Keys.DOWN)
            def f():
                selected = d.find_search_modal_selected_result()
                if result_text in selected.text:
                    d.active_elem().send_keys(Keys.ENTER)
                    return True
                return False
            if retry(f):
                return
        raise E2EFatal("did not find search result '%s'" % result_text)

    wd.get(d.sg_url('/github.com/gorilla/mux')) # start on a page with the jump modal active
    for query, result_text, url_path in repo_queries:
        d.active_elem().send_keys("/")
        d.active_elem().send_keys(query)
        wait_for(d.all_network_indicators_are_invisible)
        select_search_result_using_arrow_keys(result_text)
        wait_for(lambda: wd.current_url == d.sg_url(url_path))

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
    retry(rc)

    # Peek reference
    wait_for(lambda: len(d.find_references_menu_options()) > 0, 5)
    retry(lambda: d.find_references_menu_options()[0].click())

    # Jump to reference
    wait_for(lambda: len(d.find_references_menu_options()) > 0)
    retry(lambda: d.double_click(d.find_references_menu_options()[0]))

    # Dismiss peek view
    retry(lambda: d.active_elem().send_keys(Keys.ESCAPE))

    # Jump to modal to "NewEncryptionClient"
    retry(lambda: d.active_elem().send_keys("/"))
    retry(lambda: d.active_elem().send_keys("NewEncryptionClient"))
    wait_for(d.all_network_indicators_are_invisible)
    retry(lambda: d.active_elem().send_keys(Keys.ENTER))
    wait_for(lambda: "encryption_client.go" in wd.current_url and "/github.com/aws/aws-sdk-go" in wd.current_url)

all_tests = [
    test_repo_jump_to,
    test_golden_workflow,
]
