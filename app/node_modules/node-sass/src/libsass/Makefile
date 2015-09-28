CC       ?= gcc
CXX      ?= g++
RM       ?= rm -f
CP       ?= cp -a
MKDIR    ?= mkdir
WINDRES  ?= windres
CFLAGS   ?= -Wall
CXXFLAGS ?= -Wall
LDFLAGS  ?= -Wall
ifneq "$(COVERAGE)" "yes"
  CFLAGS   += -O2
  CXXFLAGS += -O2
  LDFLAGS  += -O2
endif
LDFLAGS  += -Wl,-undefined,error
CAT      ?= $(if $(filter $(OS),Windows_NT),type,cat)

ifneq (,$(findstring /cygdrive/,$(PATH)))
	UNAME := Cygwin
else
	ifneq (,$(findstring WINDOWS,$(PATH)))
		UNAME := Windows
	else
		ifneq (,$(findstring mingw32,$(MAKE)))
			UNAME := MinGW
		else
			ifneq (,$(findstring MINGW32,$(shell uname -s)))
				UNAME = MinGW
			else
				UNAME := $(shell uname -s)
			endif
		endif
	endif
endif

ifeq "$(LIBSASS_VERSION)" ""
	ifneq "$(wildcard ./.git/ )" ""
		LIBSASS_VERSION ?= $(shell git describe --abbrev=4 --dirty --always --tags)
	endif
endif

ifeq "$(LIBSASS_VERSION)" ""
	ifneq ("$(wildcard VERSION)","")
		LIBSASS_VERSION ?= $(shell $(CAT) VERSION)
	endif
endif

ifneq "$(LIBSASS_VERSION)" ""
	CFLAGS   += -DLIBSASS_VERSION="\"$(LIBSASS_VERSION)\""
	CXXFLAGS += -DLIBSASS_VERSION="\"$(LIBSASS_VERSION)\""
endif

# enable mandatory flag
ifeq (MinGW,$(UNAME))
	CXXFLAGS += -std=gnu++0x
	LDFLAGS  += -std=gnu++0x
else
	CXXFLAGS += -std=c++0x
	LDFLAGS  += -std=c++0x
endif

ifneq "$(SASS_LIBSASS_PATH)" ""
	CFLAGS   += -I $(SASS_LIBSASS_PATH)
	CXXFLAGS += -I $(SASS_LIBSASS_PATH)
endif

ifneq "$(EXTRA_CFLAGS)" ""
	CFLAGS   += $(EXTRA_CFLAGS)
endif
ifneq "$(EXTRA_CXXFLAGS)" ""
	CXXFLAGS += $(EXTRA_CXXFLAGS)
endif
ifneq "$(EXTRA_LDFLAGS)" ""
	LDFLAGS  += $(EXTRA_LDFLAGS)
endif

LDLIBS = -lstdc++ -lm
ifeq ($(UNAME),Darwin)
	CFLAGS += -stdlib=libc++
	CXXFLAGS += -stdlib=libc++
	LDFLAGS += -stdlib=libc++
endif

ifneq (MinGW,$(UNAME))
	LDFLAGS += -ldl
	LDLIBS += -ldl
endif

ifneq ($(BUILD),shared)
	BUILD = static
endif

ifeq (,$(PREFIX))
	ifeq (,$(TRAVIS_BUILD_DIR))
		PREFIX = /usr/local
	else
		PREFIX = $(TRAVIS_BUILD_DIR)
	endif
endif

SASS_SASSC_PATH ?= sassc
SASS_SPEC_PATH ?= sass-spec
SASS_SPEC_SPEC_DIR ?= spec
SASSC_BIN = $(SASS_SASSC_PATH)/bin/sassc
RUBY_BIN = ruby

ifeq (MinGW,$(UNAME))
	SASSC_BIN = $(SASS_SASSC_PATH)/bin/sassc.exe
endif
ifeq (Windows,$(UNAME))
	SASSC_BIN = $(SASS_SASSC_PATH)/bin/sassc.exe
endif

SOURCES = \
	ast.cpp \
	base64vlq.cpp \
	bind.cpp \
	constants.cpp \
	context.cpp \
	contextualize.cpp \
	contextualize_eval.cpp \
	cssize.cpp \
	listize.cpp \
	error_handling.cpp \
	eval.cpp \
	expand.cpp \
	extend.cpp \
	file.cpp \
	functions.cpp \
	inspect.cpp \
	lexer.cpp \
	node.cpp \
	json.cpp \
	emitter.cpp \
	output.cpp \
	parser.cpp \
	plugins.cpp \
	position.cpp \
	prelexer.cpp \
	remove_placeholders.cpp \
	sass.cpp \
	sass_util.cpp \
	sass_values.cpp \
	sass_context.cpp \
	sass_functions.cpp \
	sass_interface.cpp \
	sass2scss.cpp \
	source_map.cpp \
	to_c.cpp \
	to_string.cpp \
	units.cpp \
	utf8_string.cpp \
	util.cpp

CSOURCES = cencode.c

RESOURCES =

LIBRARIES = lib/libsass.so

ifeq (MinGW,$(UNAME))
	ifeq (shared,$(BUILD))
		CFLAGS    += -D ADD_EXPORTS
		CXXFLAGS  += -D ADD_EXPORTS
		LIBRARIES += lib/libsass.dll
		RESOURCES += res/resource.rc
	endif
else
	CFLAGS   += -fPIC
	CXXFLAGS += -fPIC
	LDFLAGS  += -fPIC
endif

OBJECTS = $(SOURCES:.cpp=.o)
COBJECTS = $(CSOURCES:.c=.o)
RCOBJECTS = $(RESOURCES:.rc=.o)

DEBUG_LVL ?= NONE

all: $(BUILD)

debug: $(BUILD)

debug-static: LDFLAGS := -g $(filter-out -O2,$(LDFLAGS))
debug-static: CFLAGS := -g -DDEBUG -DDEBUG_LVL="$(DEBUG_LVL)" $(filter-out -O2,$(CFLAGS))
debug-static: CXXFLAGS := -g -DDEBUG -DDEBUG_LVL="$(DEBUG_LVL)" $(filter-out -O2,$(CXXFLAGS))
debug-static: static

debug-shared: LDFLAGS := -g $(filter-out -O2,$(LDFLAGS))
debug-shared: CFLAGS := -g -DDEBUG -DDEBUG_LVL="$(DEBUG_LVL)" $(filter-out -O2,$(CFLAGS))
debug-shared: CXXFLAGS := -g -DDEBUG -DDEBUG_LVL="$(DEBUG_LVL)" $(filter-out -O2,$(CXXFLAGS))
debug-shared: shared

static: lib/libsass.a
shared: $(LIBRARIES)

lib:
	$(MKDIR) lib

lib/libsass.a: lib $(COBJECTS) $(OBJECTS)
	$(AR) rcvs $@ $(COBJECTS) $(OBJECTS)

lib/libsass.so: lib $(COBJECTS) $(OBJECTS)
	$(CXX) -shared $(LDFLAGS) -o $@ $(COBJECTS) $(OBJECTS) $(LDLIBS)

lib/libsass.dll: lib $(COBJECTS) $(OBJECTS) $(RCOBJECTS)
	$(CXX) -shared $(LDFLAGS) -o $@ $(COBJECTS) $(OBJECTS) $(RCOBJECTS) $(LDLIBS) -s -Wl,--subsystem,windows,--out-implib,lib/libsass.a

%.o: %.c
	$(CC) $(CFLAGS) -c -o $@ $<

%.o: %.rc
	$(WINDRES) -i $< -o $@

%.o: %.cpp
	$(CXX) $(CXXFLAGS) -c -o $@ $<

%: %.o static
	$(CXX) $(CXXFLAGS) -o $@ $+ $(LDFLAGS) $(LDLIBS)

install: install-$(BUILD)

install-static: lib/libsass.a
	install -pm0755 $< $(DESTDIR)$(PREFIX)/$<

install-shared: lib/libsass.so
	install -pm0755 $< $(DESTDIR)$(PREFIX)/$<

$(SASSC_BIN): $(BUILD)
	cd $(SASS_SASSC_PATH) && $(MAKE)

sassc: $(SASSC_BIN)
	$(SASSC_BIN) -v

test: $(SASSC_BIN)
	$(RUBY_BIN) $(SASS_SPEC_PATH)/sass-spec.rb -c $(SASSC_BIN) -s $(LOG_FLAGS) $(SASS_SPEC_PATH)/$(SASS_SPEC_SPEC_DIR)

test_build: $(SASSC_BIN)
	$(RUBY_BIN) $(SASS_SPEC_PATH)/sass-spec.rb -c $(SASSC_BIN) -s --ignore-todo $(LOG_FLAGS) $(SASS_SPEC_PATH)/$(SASS_SPEC_SPEC_DIR)

test_issues: $(SASSC_BIN)
	$(RUBY_BIN) $(SASS_SPEC_PATH)/sass-spec.rb -c $(SASSC_BIN) $(LOG_FLAGS) $(SASS_SPEC_PATH)/spec/issues

clean:
	$(RM) $(RCOBJECTS) $(COBJECTS) $(OBJECTS) $(LIBRARIES) lib/*.a lib/*.so lib/*.dll lib/*.la


.PHONY: all debug debug-static debug-shared static shared install install-static install-shared sassc clean
