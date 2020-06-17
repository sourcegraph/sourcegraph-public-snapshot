<style>

ul.r {
    list-style-position: outside !important;
    padding-left: 20px !important;
}

li.r {
    margin-top:10px !important;
    list-style:none !important;
}

.r {
    background-color: #ffffff !important;
    border: 0px !important;
    padding: 0px !important;
    margin: 0px !important;
    border-collapse: collapse !important;
    vertical-align: top !important;
}

th.r {
    text-align: left !important;
    background-color: #ece9d8 !important;
    border: 1px solid #aca899 !important;
    padding: 3px !important;
}

td.r {
    background-color: #ffffff !important;
    text-align: left !important;
    vertical-align: top !important;
    border: 1px solid #aca899 !important;
    padding: 3px !important;
}

.markdown-body code {
    background-color: #ece9d8;
    color: black;
}

img.r {
  border: 1px solid !important;
  border-radius: 3px !important;
  width: 18px !important;
  margin-right: 8px !important;
  vertical-align:bottom !important;
}

.ts {
    border: 0px !important;
    padding: 0px !important;
    margin: 0px !important;
    border-collapse: collapse !important;
    vertical-align: top !important;
    width: 16px !important;
    height: 24px !important;
    background-image: url(../img/div-ts.png);
    background-size: 16px 512px !important;
}

.te {
    border: 0px !important;
    padding: 0px !important;
    margin: 0px !important;
    border-collapse: collapse !important;
    vertical-align: top !important;
    width: 16px !important;
    height: 24px !important;
    background-image: url(../img/div-te.png);
    background-size: 16px 512px !important;
}

.le {
    border: 0px !important;
    padding: 0px !important;
    margin: 0px !important;
    border-collapse: collapse !important;
    vertical-align: top !important;
    width: 16px !important;
    height: 24px !important;
    background-image: url(../img/div-le.png);
    background-size: 16px 512px !important;
}

.ke {
    border: 0px !important;
    padding: 0px !important;
    margin: 0px !important;
    border-collapse: collapse !important;
    vertical-align: top !important;
    width: 16px !important;
    height: 24px !important;
    background-image: url(../img/div-ke.png);
    background-size: 16px 512px !important;
}


.ls {
    border: 0px !important;
    padding: 0px !important;
    margin: 0px !important;
    border-collapse: collapse !important;
    vertical-align: top !important;
    width: 16px !important;
    height: 24px !important;
    background-image: url(../img/div-ls.png);
    background-size: 16px 512px !important;
}

.ks {
    border: 0px !important;
    padding: 0px !important;
    margin: 0px !important;
    border-collapse: collapse !important;
    vertical-align: top !important;
    width: 16px !important;
    height: 24px !important;
    background-image: url(../img/div-ks.png);
    background-size: 16px 512px !important;
}


.d {
    border: 0px !important;
    padding: 0px !important;
    margin: 0px !important;
    border-collapse: collapse !important;
    vertical-align: top !important;
    min-width: 16px !important;
    height: 24px !important;
    background-image: url(../img/div-d.png);
    background-size: 1024px 512px !important;
}

.e {
    border: 0px !important;
    padding: 0px !important;
    margin: 0px !important;
    border-collapse: collapse !important;
    vertical-align: top !important;
    min-width: 16px !important;
    height: 24px !important;
    background-image: url(../img/div-e.png);
    background-size: 1024px 512px !important;
}

.c {
    padding: 1px 3px !important;
    margin: 0px 0px !important;
    border: 2px solid !important;
    -moz-border-radius: 0.4em !important;
    -webkit-border-radius: 0.4em !important;
    -khtml-border-radius: 0.4em !important;
    border-radius: 0.4em !important;
    background-color: #fff !important;
    color: black;
}

</style>

# Sourcegraph Search Query Language

This page provides a complete visual breakdown of the Sourcegraph Search Query
Language and some helpful examples to get you started. The railroad diagrams
show how to combine pieces of syntax. Read them from left-to-right by following
the lines. When a line splits it means there are multiple options available.
When it is possible to repeat a previous syntax, you'll see a line lead into
a box that looks like this:

<table class="r">
  <tbody>
    <tr class="r">
      <table class="r">
        <tbody>
          <tr class="r">
            <td class="ts"></td>
            <td class="d"> </td>
            <td class="te"></td>
          </tr>
          <tr class="r">
            <td class="ls"></td>
            <td class="d"><code class="c">...</code></td>
            <td class="le"></td>
          </tr>
        </tbody>
      </table>
    </tr>
  </tbody>
</table>

## Basic query

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="ts"></td>
                <td class="d"><code class="c"><a href="#search-pattern">search pattern</a></code></td>
                <td class="te"></td>
              </tr>
              <tr class="r">
                <td class="ls"></td>
                <td class="d"><code class="c"><a href="#parameter">parameter</a></code></td>
                <td class="le"></td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="ts"></td>
                <td class="d"> </td>
                <td class="te"></td>
              </tr>
              <tr class="r">
                <td class="ls"></td>
                <td class="d"><code class="c">...</code></td>
                <td class="le"></td>
              </tr>
            </tbody>
          </table>
        </td>
      </tr>
    </tbody>
  </table>
</div>

## Expression

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="d"></td>
                <td class="d"><code class="c"><a href="#basic-query">basic query</a></code></td>
                <td class="d"></td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="ts"></td>
                <td class="d"> </td>
                <td class="d"></td>
              </tr>
              <tr class="r">
                <td class="ks"></td>
                <td class="d"><code class="c">AND</code></td>
                <td class="te"></td>
              </tr>
              <tr class="r">
                <td class="ls"></td>
                <td class="d"><code class="c">OR</code></td>
                <td class="le"></td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="d"></td>
              </tr>
              <tr class="r">
                <td class="d"><code class="c"><a href="#basic-query">basic query</a></code></td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="d"></td>
                <td class="d"> </td>
                <td class="d"></td>
              </tr>
              <tr class="r">
                <td class="ts"></td>
                <td class="d"> </td>
                <td class="te"></td>
              </tr>
              <tr class="r">
                <td class="ls"></td>
                <td class="d"><code class="c">...</code></td>
                <td class="le"></td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="te"></td>
              </tr>
              <tr class="r">
                <td class="le"> </td>
              </tr>
            </tbody>
          </table>
        </td>
      </tr>
    </tbody>
  </table>
</div>

For example, `foo or bar and bar` means `foo or (bar and baz)`.

Expressions are the basic building blocks for search queries. Typical queries
contain a search pattern and some parameters to narrow search. For
example, `testroute repo:gorilla/mux`.

## Search pattern

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="ts"></td>
                <td class="d"><code class="c"><a href="#string">string</a></code></td>
                <td class="te"></td>
              </tr>
              <tr class="r">
                <td class="ls"></td>
                <td class="d"><code class="c"><a href="#quoted-string">quoted string</a></code></td>
                <td class="le"></td>
              </tr>
            </tbody>
          </table>
        </td>
      </tr>
    </tbody>
  </table>
</div>

A pattern to search. By default the pattern is searched literally. The kind of search may be toggled, in which case the pattern matches differently:
<ul class="r">
    <li class="r"><img class="r" src="../img/regex.png">Interpret the pattern as a regular expression. We support [RE2 syntax](https://golang.org/s/re2syntax). If the pattern is a [quoted string](#quoted-string) we search for it literally.</li>
    <li class="r"><img class="r" src="../img/structural.png">Interpret the pattern as a [structural search](https://docs.sourcegraph.com/user/search/structural) pattern.</li>
</ul>


## Parameter

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r"></tr>
      <tr class="r">
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="d">
                  <table class="r">
                    <tbody>
                      <tr class="r">
                        <td class="d">
                          <table class="r">
                            <tbody>
                              <tr class="r">
                                <td class="ts"></td>
                                <td class="d"><code class="c"><a href="#repo">repo</a></code></td>
                                <td class="te"></td>
                              </tr>
                              <tr class="r">
                                <td class="ks"></td>
                                <td class="d"><code class="c"><a href="#file">file</a></code></td>
                                <td class="ke"></td>
                              </tr>
                              <tr class="r">
                                <td class="ks"></td>
                                <td class="d"><code class="c"><a href="#content">content</a></code></td>
                                <td class="ke"></td>
                              </tr>
                              <tr class="r">
                                <td class="ks"></td>
                                <td class="d"><code class="c"><a href="#language">language</a></code></td>
                                <td class="ke"></td>
                              </tr>
                              <tr class="r">
                                <td class="ks"></td>
                                <td class="d"><code class="c"><a href="#type">type</a></code></td>
                                <td class="ke"></td>
                              </tr>
                              <tr class="r">
                                <td class="ks"></td>
                                <td class="d"><code class="c"><a href="#case">case</a></code></td>
                                <td class="ke"></td>
                              </tr>
                              <tr class="r">
                                <td class="ks"></td>
                                <td class="d"><code class="c"><a href="#fork">fork<a/></code></td>
                                <td class="ke"></td>
                              </tr>
                              <tr class="r">
                                <td class="ks"></td>
                                <td class="d"><code class="c"><a href="#archived">archived</a></code></td>
                                <td class="ke"></td>
                              </tr>
                              <tr class="r">
                                <td class="ks"></td>
                                <td class="d"><code class="c"><a href="#repogroup">repogroup</a></code></td>
                                <td class="ke">
                                </td>
                                <tr class="r">
                                  <td class="ks"></td>
                                  <td class="d"><code class="c"><a href="#repo-has-file">repo has file</a></code></td>
                                  <td class="ke"></td>
                                </tr>
                                <tr class="r">
                                  <td class="ks"></td>
                                  <td class="d"><code class="c"><a href="#repo-has-commit-after">repo has commit after</a></code></td>
                                  <td class="ke"></td>
                                </tr>
                                <tr class="r">
                                  <td class="ks"></td>
                                  <td class="d"><code class="c"><a href="#count">count</a></code></td>
                                  <td class="ke"></td>
                                </tr>
                                <tr class="r">
                                  <td class="ks"></td>
                                  <td class="d"><code class="c"><a href="#timeout">timeout</a></code></td>
                                  <td class="ke"></td>
                                </tr>
                                <tr class="r">
                                  <td class="ks"></td>
                                  <td class="d"><code class="c"><a href="#visibility">visibility</a></code></td>
                                  <td class="ke"></td>
                                </tr>
                                <tr class="r">
                                  <td class="ls"></td>
                                  <td class="d"><code class="c"><a href="#pattern-type">pattern type</a></code></td>
                                  <td class="le"></td>
                                </tr>
                            </tbody>
                          </table>
                        </td>
                        </tr>
                    </tbody>
                  </table>
                </td>
                </tr>
            </tbody>
          </table>
        </td>
        </tr>
    </tbody>
  </table>
</div>

Search parameters allow you to narrow search queries and modify search behavior.

### Repo

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="ts"></td>
                <td class="d">&nbsp;</td>
                <td class="te"></td>
              </tr>
              <tr class="r">
                <td class="ls"></td>
                <td class="d"><code class="c">–</code></td>
                <td class="le"></td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="d"><table class="r">
            <tbody>
              <tr class="r">
                <td class="ts"></td>
                <td class="d"><code class="c">repo:</code></td>
                <td class="te">
                </td>
              </tr>
              <tr class="r">
                <td class="ls"></td>
                <td class="d">
                  <table class="r">
                    <tbody>
                      <tr class="r">
                        <td class="d"><code class="c">r:</code></td>
                        <td class="d"></td>
                        <td class="d"></td>
                      </tr>
                    </tbody>
                  </table>
                </td>
                <td class="le">
                </td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="d"><code class="c"><a href="#regular-expression">regular expression</a></code></td>
        <td class="d">
          <td class="d">
            <table class="r">
              <tbody>
                <tr class="r">
                  <td class="ts"></td>
                  <td class="d">&nbsp;</td>
                  <td class="te"></td>
                </tr>
                <tr class="r">
                  <td class="ls"></td>
                  <td class="d"><code class="c">@<a href="#revision">revision</a></code></td>
                  <td class="le"></td>
                </tr>
              </tbody>
            </table>
          </td>
      </tr>
    </tbody>
  </table>
</div>

Search repositories that match the regular expression.

A `-` before `repo` excludes the repository. By default
the repository will be searched at the `HEAD` commit of the default
branch. You can optionally change the [revision](#revision).

### Revision

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="ts"></td>
                <td class="d"><code class="c">branch name</code></td>
                <td class="te"></td>
              </tr>
              <tr class="r">
                <td class="ks"></td>
                <td class="d"><code class="c">commit hash</code></td>
                <td class="ke"></td>
              </tr>
              <tr class="r">
                <td class="ls"></td>
                <td class="d"><code class="c">git tag</code></td>
                <td class="le"></td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="d">
          <td class="d">
            <table class="r">
              <tbody>
                <tr class="r">
                  <td class="ts"></td>
                  <td class="d">&nbsp;</td>
                  <td class="te"></td>
                </tr>
                <tr class="r">
                  <td class="ls"></td>
                  <td class="d">
                    <table class="r">
                      <tbody>
                        <tr class="r">
                          <td class="d"><code class="c">:<a href="#revision">revision</a></code></td>
                          <td class="d">
                            <table class="r">
                              <tbody>
                                <tr class="r">
                                  <td class="ts"></td>
                                  <td class="d">&nbsp;</td>
                                  <td class="te"></td>
                                </tr>
                                <tr class="r">
                                  <td class="ls"></td>
                                  <td class="d"><code class="c">...</code></td>
                                  <td class="le"></td>
                                </tr>
                              </tbody>
                            </table>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </td>
                  <td class="le"></td>
                </tr>
              </tbody>
            </table>
          </td>
      </tr>
    </tbody>
  </table>
</div>

Search a repository at a given revision. For example, a branch name, commit hash, or git tag.

### File

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="ts"></td>
                <td class="d">&nbsp;</td>
                <td class="te"></td>
              </tr>
              <tr class="r">
                <td class="ls"></td>
                <td class="d"><code class="c">–</code></td>
                <td class="le"></td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="d"><table class="r">
            <tbody>
              <tr class="r">
                <td class="ts"></td>
                <td class="d"><code class="c">file:</code></td>
                <td class="te">
                </td>
              </tr>
              <tr class="r">
                <td class="ls"></td>
                <td class="d">
                  <table class="r">
                    <tbody>
                      <tr class="r">
                        <td class="d"><code class="c">f:</code></td>
                        <td class="d"></td>
                        <td class="d"></td>
                      </tr>
                    </tbody>
                  </table>
                </td>
                <td class="le">
                </td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="d"><code class="c"><a href="#regular-expression">regular expression</a></code></td>
        <td class="d">
      </tr>
    </tbody>
  </table>
</div>

Search files whose full path matches the regular expression. A `-` before `file`
excludes the file from being searched.

### Language

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="ts"></td>
        <td class="d"><code class="c">language:</code></td>
        <td class="te">
        </td>
      </tr>
      <tr class="r">
        <td class="ks"></td>
        <td class="d"><code class="c">lang:</code></td>
        <td class="ke">
        </td>
      </tr>
      <tr class="r">
        <td class="ls"></td>
        <td class="d"><code class="c">l:</code></td>
        <td class="le">
        </td>
      </tr>
    </tbody>
  </table>
</tbody>
</table>
</div>

Only search files in the specified programming language, like `typescript` or
`python`.

### Content

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d"></td>
        <td class="d"><code class="c">content:</code></td>
        <td class="d"></td>
        <td class="d"><code class="c"><a href="#quoted-string">quoted string</a></code></td>
        <td class="d">
      </tr>
    </tbody>
  </table>
</div>

Set the search pattern to search using a dedicated parameter.

Useful, for example, when searching a string
like `content:"repo:foo"` that may conflict with the syntax of
parameters in this Sourcegraph language.

### Type

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="d">
                  <table class="r">
                    <tbody>
                      <tr class="r">
                        <td class="d"></td>
                        <td class="d"><code class="c">type:</code></td>
                        <td class="d">
                          <table class="r">
                            <tbody>
                              <tr class="r">
                                <td class="d">
                                  <table class="r">
                                    <tbody>
                                      <tr class="r">
                                        <td class="ts"></td>
                                        <td class="d"><code class="c">symbol</code></td>
                                        <td class="te"></td>
                                      </tr>
                                      <tr class="r">
                                        <td class="ks"></td>
                                        <td class="d"><code class="c">repo</code></td>
                                        <td class="ke"></td>
                                      </tr>
                                      <tr class="r">
                                        <td class="ks"></td>
                                        <td class="d"><code class="c">path</code></td>
                                        <td class="ke"></td>
                                      </tr>
                                      <tr class="r">
                                        <td class="ks"></td>
                                        <td class="d">
                                          <table class="r">
                                            <tbody>
                                              <tr class="r">
                                                <td class="d"><code class="c">file</code></td>
                                                <td class="d"></td>
                                                <td class="d"></td>
                                              </tr>
                                            </tbody>
                                          </table>
                                        </td>
                                </td>
                                <td class="le"></td>
                                </tr>
                                </tbody>
                                </table>
                        </td>
                        <td class="d"></td>
                        <td class="d"></td>
                        <td class="d"></td>
                        <td class="d"></td>
                        <td class="d"></td>
                        <td class="d"></td>
                        <td class="d"></td>
                        <td class="d"></td>
                        <td class="d"></td>
                        </tr>
                        </tbody>
                        </table>
                </td>
                <td class="te"></td>
                </tr>
                <tr class="r">
                  <td class="e"></td>
                  <td class="e"></td>
                  <td class="d">
                    <table class="r">
                      <tbody>
                        <tr class="r">
                          <td class="d">
                            <table class="r">
                              <tbody>
                                <tr class="r">
                                  <td class="ks"></td>
                                  <td class="d"><code class="c">commit</code></td>
                                  <td class="te"></td>
                                </tr>
                                <tr class="r">
                                  <td class="ls"></td>
                                  <td class="d"><code class="c">diff</code></td>
                                  <td class="le"></td>
                                </tr>
                              </tbody>
                            </table>
                          </td>
                          <td class="d"><code class="c"><a href="#commit-parameter"> commit parameter </a></code></td>
                          <td class="d"></td>
                        </tr>
                      </tbody>
                    </table>
                  </td>
                  <td class="le"></td>
                </tr>
                </tbody>
                </table>
        </td>
        </tr>
        </tbody>
        </table>
</td>
</tr>
</tbody>
</table>
</div>

Set whether the search pattern should only conduct a search of a certain type (e.g., only files or repos), or to perform special commit or diff searches.

### Case

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d"></td>
        <td class="d"><code class="c">case:</code></td>
        <td class="d"></td>
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="ts">
                  <td class="d"><code class="c">yes</code></td>
                  <td class="te">
              </tr>
              <tr class="r">
                <td class="ls">
                  <td class="d"><code class="c">no</code></td>
                  <td class="le">
              </tr>
            </tbody>
          </table>
          </td>
      </tr>
    </tbody>
  </table>
</div>

Set whether the search pattern should be treated case-sensitively.


### Fork

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d"></td>
        <td class="d"><code class="c">fork:</code></td>
        <td class="d"></td>
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="ts">
                  <td class="d"><code class="c">yes</code></td>
                  <td class="te">
              </tr>
              <tr class="r">
                <td class="ks">
                  <td class="d"><code class="c">no</code></td>
                  <td class="ke">
              </tr>
              <tr class="r">
                <td class="ls">
                  <td class="d"><code class="c">only</code></td>
                  <td class="le">
              </tr>
            </tbody>
          </table>
          </td>
      </tr>
    </tbody>
  </table>
</div>

Set to `yes` if repository forks should be included or `only` if only forks
should be searched. Respository forks are excluded by default.

### Archived

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d"></td>
        <td class="d"><code class="c">archived:</code></td>
        <td class="d"></td>
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="ts">
                  <td class="d"><code class="c">yes</code></td>
                  <td class="te">
              </tr>
              <tr class="r">
                <td class="ks">
                  <td class="d"><code class="c">no</code></td>
                  <td class="ke">
              </tr>
              <tr class="r">
                <td class="ls">
                  <td class="d"><code class="c">only</code></td>
                  <td class="le">
              </tr>
            </tbody>
          </table>
          </td>
      </tr>
    </tbody>
  </table>
</div>

Set to `yes` if archived repositories should be included or `only` if only
archives should be searched. Archived repositories are excluded by default.

### Repo group

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="ts"></td>
        <td class="d"><code class="c">repogroup:</code></td>
        <td class="te">
        </td>
      </tr>
      <tr class="r">
        <td class="ls"></td>
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="d"><code class="c">g:</code></td>
                <td class="d"></td>
                <td class="d"></td>
                <td class="d"></td>
                <td class="d"></td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="le">
        </td>
      </tr>
    </tbody>
  </table>
</div>

Only include results from the named group of repositories (defined by the server
admin). Same as using [repo](#repo) that matches all of the group’s
repositories. Use [repo](#repo) unless you know that the group
exists.


### Repo has file

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="ts"></td>
                <td class="d">&nbsp;</td>
                <td class="te"></td>
              </tr>
              <tr class="r">
                <td class="ls"></td>
                <td class="d"><code class="c">–</code></td>
                <td class="le"></td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="d"><table class="r">
            <tbody>
              <tr class="r">
                <td class="d"></td>
                <td class="d"><code class="c">repohasfile:</code></td>
                <td class="d">
                </td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="d"><code class="c"><a href="#regular-expression">regular expression</a></code></td>
        <td class="d">
      </tr>
    </tbody>
  </table>
</div>

Only include results from repositories that contain a matching file. This
keyword is a pure filter, so it requires at least one other search term in the
query. Note: this filter currently only works on text matches and file path
matches.

### Repo has commit after

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d"><table class="r">
            <tbody>
              <tr class="r">
                <td class="d"></td>
                <td class="d"><code class="c">repohascommitafter:</code></td>
                <td class="d">
                </td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="d"><code class="c"><a href="#quoted-string">quoted string</a></code></td>
        <td class="d">
      </tr>
    </tbody>
  </table>
</div>

Filter out stale repositories that don’t contain commits past the specified time
frame. This parameter is experimental.

### Count

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d"><table class="r">
            <tbody>
              <tr class="r">
                <td class="d"></td>
                <td class="d"><code class="c">count:</code></td>
                <td class="d">
                </td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="d"><code class="c">number</code></td>
        <td class="d">
      </tr>
    </tbody>
  </table>
</div>

Retrieve at least N results. By default, Sourcegraph stops searching early and
returns if it finds a full page of results. This is desirable for most
interactive searches. To wait for all results, or to see results beyond the
first page, use the count: keyword with a larger N.

### Timeout

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d"><table class="r">
            <tbody>
              <tr class="r">
                <td class="d"></td>
                <td class="d"><code class="c">timeout:</code></td>
                <td class="d">
                </td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="d"><code class="c">time value</code></td>
        <td class="d">
      </tr>
    </tbody>
  </table>
</div>

Set a search timeout. The time value is a string like 10s or 100ms, which is
parsed by the Go time
package's [ParseDuration](https://golang.org/pkg/time/#ParseDuration).
By default the timeout is set to 10 seconds, and the search will optimize for
returning results as soon as possible. The timeout value cannot be set longer
than 1 minute.

### Visibility

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d"></td>
        <td class="d"><code class="c">visibility:</code></td>
        <td class="d"></td>
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="ts">
                  <td class="d"><code class="c">any</code></td>
                  <td class="te">
              </tr>
              <tr class="r">
                <td class="ks">
                  <td class="d"><code class="c">public</code></td>
                  <td class="ke">
              </tr>
              <tr class="r">
                <td class="ls">
                  <td class="d"><code class="c">private</code></td>
                  <td class="le">
              </tr>
            </tbody>
          </table>
          </td>
      </tr>
    </tbody>
  </table>
</div>

Filter results to only public or private repositories. The default is to include
both private and public repositories.

### Pattern type

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d"></td>
        <td class="d"><code class="c">patterntype:</code></td>
        <td class="d"></td>
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="ts">
                  <td class="d"><code class="c">literal</code></td>
                  <td class="te">
              </tr>
              <tr class="r">
                <td class="ks">
                  <td class="d"><code class="c">regexp</code></td>
                  <td class="ke">
              </tr>
              <tr class="r">
                <td class="ls">
                  <td class="d"><code class="c">structural</code></td>
                  <td class="le">
              </tr>
            </tbody>
          </table>
          </td>
      </tr>
    </tbody>
  </table>
</div>

Set whether the pattern should run a literal search, regular expression search,
or a structural search pattern. This keyword is available as a commannd-line and
accessibility option, and equivalent to the visual toggles
in [search pattern](#search-pattern).

## Regular expression

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="ts"></td>
                <td class="d"><code class="c"><a href="#string">string</a></code></td>
                <td class="te"></td>
              </tr>
              <tr class="r">
                <td class="ls"></td>
                <td class="d"><code class="c"><a href="#quoted-string">quoted string</a></code></td>
                <td class="le"></td>
              </tr>
            </tbody>
          </table>
        </td>
      </tr>
    </tbody>
  </table>
</div>

A string that is interpreted as a <a href="https://golang.org/s/re2syntax">RE2</a> regular expression.

## String

<div name="r">
  <tr class="r">
    <td class="d"><code class="c">string</code></td>
  </tr>
</div>

An unquoted string is any contiguous sequence of characters not containing whitespace.


## Quoted string

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="ts"></td>
        <td class="d"><code class="c">"any string"</code></td>
        <td class="te"></td>
      </tr>
      <tr class="r">
        <td class="ls"></td>
        <td class="d"><code class="c">'any string'</code></td>
        <td class="le"></td>
      </tr>
    </tbody>
  </table>
</div>

Any string, including whitespace, may be quoted with single `'` or double `"`
quotes. Quotes can be escaped with `\`.

## Commit parameter

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="ts"></td>
                <td class="d"><code class="c"><a href="#author">author</a></code></td>
                <td class="te"></td>
              </tr>
              <tr class="r">
                <td class="ks"></td>
                <td class="d"><code class="c"><a href="#before">before</a></code></td>
                <td class="ke"></td>
              </tr>
              <tr class="r">
                <td class="ks"></td>
                <td class="d"><code class="c"><a href="#after">after</a></code></td>
                <td class="ke"></td>
              </tr>
              <tr class="r">
                <td class="ls"></td>
                <td class="d"><code class="c"><a href="#message">message</a></code></td>
                <td class="le"></td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="ts"></td>
                <td class="d"></td>
                <td class="te"></td>
              </tr>
              <tr class="r">
                <td class="ls"></td>
                <td class="d"><code class="c">...</code></td>
                <td class="le"></td>
              </tr>
            </tbody>
          </table>
        </td>
      </tr>
    </tbody>
  </table>
</div>

Set parameters that apply only to commit and diff searches.

### Author

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="d"></td>
                <td class="d"><code class="c">author:</code></td>
                <td class="d"></td>
                <td class="d"><code class="c"><a href="#regular expression">regular expression</a></code></td>
                <td class="d"></td>
              </tr>
            </tbody>
          </table>
        </td>
      </tr>
    </tbody>
  </table>
</div>

Include commits or diffs that are authored by the user.

### Before

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="ts"></td>
                <td class="d"><code class="c">before:</code></td>
                <td class="te">
                </td>
              </tr>
              <tr class="r">
                <td class="ls"></td>
                <td class="d">
                  <table class="r">
                    <tbody>
                      <tr class="r">
                        <td class="d"><code class="c">until:</code></td>
                        <td class="d"></td>
                      </tr>
                    </tbody>
                  </table>
                </td>
                <td class="le">
                </td>
              </tr>
            </tbody>
          </table>
        </td>
        <td class="d"><code class="c"><a href="#quoted-string">quoted string</a></code></td>
        <td class="d"></td>
      </tr>
    </tbody>
  </table>
</div>

Include results which have a commit date before the specified time frame.

Example time values are `last thursday` or `november 1 2019`.

### After

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="d"><table class="r">
                    <tbody>
                      <tr class="r">
                        <td class="ts"></td>
                        <td class="d"><code class="c">after:</code></td>
                        <td class="te">
                        </td>
                      </tr>
                      <tr class="r">
                        <td class="ls"></td>
                        <td class="d">
                          <table class="r">
                            <tbody>
                              <tr class="r">
                                <td class="d"><code class="c">since:</code></td>
                              </tr>
                            </tbody>
                          </table>
                        </td>
                        <td class="le">
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </td>
                <td class="d"><code class="c"><a href="#quoted-string">quoted string</a></code></td>
                <td class="d"></td>
              </tr>
            </tbody>
          </table>
        </td>
      </tr>
    </tbody>
  </table>
</div>

Include results which have a commit date before the specified time frame.

Example time values are `last thursday` or `november 1 2019`.

### Message

<div name="r">
  <table class="r">
    <tbody>
      <tr class="r">
        <td class="d">
          <table class="r">
            <tbody>
              <tr class="r">
                <td class="d"><table class="r">
                    <tbody>
                      <tr class="r">
                        <td class="ts"></td>
                        <td class="d"><code class="c">message:</code></td>
                        <td class="te">
                        </td>
                      </tr>
                      <tr class="r">
                        <td class="ks"></td>
                        <td class="d"><code class="c">msg:</code></td>
                        <td class="ke"></td>
                      </tr>
                      <tr class="r">
                        <td class="ls"></td>
                        <td class="d"><code class="c">m:</code></td>
                        <td class="le">
                      </tr>
                    </tbody>
                  </table>
                  </td>
                  <td class="d"><code class="c"><a href="#quoted-string">quoted string</a></code></td>
                  <td class="d"></td>
              </tr>
            </tbody>
          </table>
          </td>
      </tr>
    </tbody>
  </table>
</div>

Include results which have commit messages containing the string.
