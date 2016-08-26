import re
import os
import sys
import json
import posixpath
from urlparse import urljoin
from docutils import nodes
from docutils.io import StringOutput
from docutils.nodes import document, section

from sphinx import addnodes
from sphinx.domains import Domain
from sphinx.util.osutil import relative_uri
from sphinx.builders.html import StandaloneHTMLBuilder, DirectoryHTMLBuilder


_edition_re = re.compile(r'^(\s*)..\s+sentry:edition::\s*(.*?)$')
_docedition_re = re.compile(r'^..\s+sentry:docedition::\s*(.*?)$')


EXTERNAL_DOCS_URL = 'https://docs.getsentry.com/hosted/'


def make_link_builder(app, base_page):
    def link_builder(edition, to_current=False):
        here = app.builder.get_target_uri(base_page)
        if to_current:
            uri = relative_uri(here, '../' + edition + '/' +
                               here.lstrip('/')) or './'
        else:
            root = app.builder.get_target_uri(app.env.config.master_doc) or './'
            uri = relative_uri(here, root) or ''
            if app.builder.name in ('sentryhtml', 'html'):
                uri = (posixpath.dirname(uri or '.') or '.').rstrip('/') + \
                    '/../' + edition + '/index.html'
            else:
                uri = uri.rstrip('/') + '/../' + edition + '/'
        return uri
    return link_builder


def html_page_context(app, pagename, templatename, context, doctree):
    rendered_toc = get_rendered_toctree(app.builder, pagename)
    context['full_toc'] = rendered_toc

    context['link_to_edition'] = make_link_builder(app, pagename)

    def render_sitemap():
        return get_rendered_toctree(app.builder, 'sitemap', collapse=False)
    context['render_sitemap'] = render_sitemap

    context['sentry_doc_variant'] = app.env.config.sentry_doc_variant


def get_rendered_toctree(builder, docname, prune=False, collapse=True):
    fulltoc = build_full_toctree(builder, docname, prune=prune,
                                 collapse=collapse)
    rendered_toc = builder.render_partial(fulltoc)['fragment']
    return rendered_toc


def build_full_toctree(builder, docname, prune=False, collapse=True):
    env = builder.env
    doctree = env.get_doctree(env.config.master_doc)
    toctrees = []
    for toctreenode in doctree.traverse(addnodes.toctree):
        toctrees.append(env.resolve_toctree(docname, builder, toctreenode,
                                            collapse=collapse,
                                            titles_only=True,
                                            includehidden=True,
                                            prune=prune))
    if not toctrees:
        return None
    result = toctrees[0]
    for toctree in toctrees[1:]:
        if toctree:
            result.extend(toctree.children)
    env.resolve_references(result, docname, builder)
    return result


def parse_rst(state, content_offset, doc):
    node = nodes.section()
    # hack around title style bookkeeping
    surrounding_title_styles = state.memo.title_styles
    surrounding_section_level = state.memo.section_level
    state.memo.title_styles = []
    state.memo.section_level = 0
    state.nested_parse(doc, content_offset, node, match_titles=1)
    state.memo.title_styles = surrounding_title_styles
    state.memo.section_level = surrounding_section_level
    return node.children


class SentryDomain(Domain):
    name = 'sentry'
    label = 'Sentry'
    directives = {
    }


def preprocess_source(app, docname, source):
    source_lines = source[0].splitlines()

    def _find_block(indent, lineno):
        block_indent = len(indent.expandtabs())
        rv = []
        actual_indent = None

        while lineno < end:
            line = source_lines[lineno]
            if not line.strip():
                rv.append(u'')
            else:
                expanded_line = line.expandtabs()
                indent = len(expanded_line) - len(expanded_line.lstrip())
                if indent > block_indent:
                    if actual_indent is None or indent < actual_indent:
                        actual_indent = indent
                    rv.append(line)
                else:
                    break
            lineno += 1

        if rv:
            rv.append(u'')
            if actual_indent:
                rv = [x[actual_indent:] for x in rv]
        return rv, lineno

    result = []

    lineno = 0
    end = len(source_lines)
    while lineno < end:
        line = source_lines[lineno]
        match = _edition_re.match(line)
        if match is None:
            # Skip sentry:docedition.  We don't want those.
            match = _docedition_re.match(line)
            if match is None:
                result.append(line)
            lineno += 1
            continue
        lineno += 1
        indent, tags = match.groups()
        tags = set(x.strip() for x in tags.split(',') if x.strip())
        should_include = app.env.config.sentry_doc_variant in tags
        block_lines, lineno = _find_block(indent, lineno)
        if should_include:
            result.extend(block_lines)

    source[:] = [u'\n'.join(result)]


def builder_inited(app):
    app.env.sentry_referenced_docs = {}


def remove_irrelevant_trees(app, doctree):
    docname = app.env.temp_data['docname']
    rd = app.env.sentry_referenced_docs
    for toctreenode in doctree.traverse(addnodes.toctree):
        for e in toctreenode['entries']:
            rd.setdefault(str(e[1]), set()).add(docname)


def is_referenced(docname, references):
    if docname == 'index':
        return True
    seen = set([docname])
    to_process = set(references.get(docname) or ())
    while to_process:
        if 'index' in to_process:
            return True
        next = to_process.pop()
        seen.add(next)
        for backlink in references.get(next) or ():
            if backlink in seen:
                continue
            else:
                to_process.add(backlink)
    return False


class SphinxBuilderMixin(object):
    build_wizard_fragment = False

    @property
    def add_permalinks(self):
        return not self.build_wizard_fragment

    def get_target_uri(self, *args, **kwargs):
        rv = super(SphinxBuilderMixin, self).get_target_uri(*args, **kwargs)
        if self.build_wizard_fragment:
            rv = urljoin(EXTERNAL_DOCS_URL, rv)
        return rv

    def get_relative_uri(self, from_, to, typ=None):
        if self.build_wizard_fragment:
            return self.get_target_uri(to, typ)
        return super(SphinxBuilderMixin, self).get_relative_uri(
            from_, to, typ)

    def write_doc(self, docname, doctree):
        if is_referenced(docname, self.app.env.sentry_referenced_docs):
            return super(SphinxBuilderMixin, self).write_doc(docname, doctree)
        else:
            print 'skipping because unreferenced'

    def __iter_wizard_files(self):
        for dirpath, dirnames, filenames in os.walk(self.srcdir):
            dirnames[:] = [x for x in dirnames if x[:1] not in '_.']
            for filename in filenames:
                if filename == 'sentry-doc-config.json':
                    full_path = os.path.join(self.srcdir, dirpath)
                    base_path = full_path[len(self.srcdir):].strip('/\\') \
                        .replace(os.path.sep, '/')
                    yield os.path.join(full_path, filename), base_path

    def __build_wizard_section(self, base_path, snippets):
        trees = {}
        rv = []

        def _build_node(node):
            original_header_level = self.docsettings.initial_header_level
            # bump initial header level to two
            self.docsettings.initial_header_level = 2
            # indicate that we're building for the wizard fragements.
            # This changes url generation and more.
            self.build_wizard_fragment = True
            # Embed pygments colors as inline styles
            original_args = self.highlighter.formatter_args
            self.highlighter.formatter_args = original_args.copy()
            self.highlighter.formatter_args['noclasses'] = True
            try:
                sub_doc = document(self.docsettings,
                                   doctree.reporter)
                sub_doc += node
                destination = StringOutput(encoding='utf-8')
                self.current_docname = docname
                self.docwriter.write(sub_doc, destination)
                self.docwriter.assemble_parts()
                rv.append(self.docwriter.parts['fragment'])
            finally:
                self.build_wizard_fragment = False
                self.highlighter.formatter_args = original_args
                self.docsettings.initial_header_level = original_header_level

        for snippet in snippets:
            if '#' not in snippet:
                snippet_path = snippet
                section_name = None
            else:
                snippet_path, section_name = snippet.split('#', 1)
            docname = posixpath.join(base_path, snippet_path)
            if docname in trees:
                doctree = trees.get(docname)
            else:
                doctree = self.env.get_and_resolve_doctree(docname, self)
                trees[docname] = doctree

            if section_name is None:
                _build_node(next(iter(doctree.traverse(section))))
            else:
                for sect in doctree.traverse(section):
                    if section_name in sect['ids']:
                        _build_node(sect)

        return u'\n\n'.join(rv)

    def __write_wizard(self, data, base_path):
        for uid, framework_data in data.get('wizards', {}).iteritems():
            body = self.__build_wizard_section(base_path,
                                               framework_data['snippets'])

            fn = os.path.join(self.outdir, '_wizards', '%s.json' % uid)
            try:
                os.makedirs(os.path.dirname(fn))
            except OSError:
                pass

            doc_link = framework_data.get('doc_link')
            if doc_link is not None:
                doc_link = urljoin(EXTERNAL_DOCS_URL,
                                   posixpath.join(base_path, doc_link))
            with open(fn, 'w') as f:
                json.dump({
                    'name': framework_data.get('name') or uid.title(),
                    'is_framework': framework_data.get('is_framework', False),
                    'doc_link': doc_link,
                    'client_lib': framework_data.get('client_lib'),
                    'body': body
                }, f)
                f.write('\n')

    def __write_wizards(self):
        for filename, base_path in self.__iter_wizard_files():
            with open(filename) as f:
                data = json.load(f)
                self.__write_wizard(data, base_path)

    def finish(self):
        super(SphinxBuilderMixin, self).finish()
        self.__write_wizards()


class SentryStandaloneHTMLBuilder(SphinxBuilderMixin, StandaloneHTMLBuilder):
    name = 'sentryhtml'


class SentryDirectoryHTMLBuilder(SphinxBuilderMixin, DirectoryHTMLBuilder):
    name = 'sentrydirhtml'


def setup(app):
    from sphinx.highlighting import lexers
    from pygments.lexers.web import PhpLexer
    lexers['php'] = PhpLexer(startinline=True)

    app.add_domain(SentryDomain)
    app.connect('builder-inited', builder_inited)
    app.connect('html-page-context', html_page_context)
    app.connect('source-read', preprocess_source)
    app.connect('doctree-read', remove_irrelevant_trees)
    app.add_builder(SentryStandaloneHTMLBuilder)
    app.add_builder(SentryDirectoryHTMLBuilder)
    app.add_config_value('sentry_doc_variant', None, 'env')


def activate():
    """Changes the config to something that the sentry doc infrastructure
    expects.
    """
    frm = sys._getframe(1)
    globs = frm.f_globals

    globs.setdefault('sentry_doc_variant',
                     os.environ.get('SENTRY_DOC_VARIANT', 'self'))
    globs['extensions'] = list(globs.get('extensions') or ()) + ['sentryext']
    globs['primary_domain'] = 'std'
    globs['exclude_patterns'] = list(globs.get('exclude_patterns')
                                     or ()) + ['_sentryext']
