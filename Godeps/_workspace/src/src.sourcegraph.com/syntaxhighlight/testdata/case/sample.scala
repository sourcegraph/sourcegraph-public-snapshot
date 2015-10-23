/* sbt -- Simple Build Tool
 * Copyright 2011 Mark Harrah
 */
package sbt

import scala.concurrent.duration.{ FiniteDuration, Duration }
import sbt.internal.util.Attributed
import sbt.internal.util.Attributed.data
import Scope.{ fillTaskAxis, GlobalScope, ThisScope }
import sbt.Compiler.InputsWithPrevious
import sbt.internal.librarymanagement.mavenint.{ PomExtraDependencyAttributes, SbtPomExtraProperties }
import xsbt.api.Discovery
import xsbti.compile.CompileOrder
import Project.{ inConfig, inScope, inTask, richInitialize, richInitializeTask, richTaskSessionVar }
import Def.{ Initialize, ScopedKey, Setting, SettingsDefinition }
import sbt.internal.librarymanagement.{ CustomPomParser, DependencyFilter }
import sbt.librarymanagement.Artifact.{ DocClassifier, SourceClassifier }
import sbt.librarymanagement.{ Configuration, Configurations, ConflictManager, CrossVersion, MavenRepository, Resolver, ScalaArtifacts, UpdateOptions }
import sbt.librarymanagement.Configurations.{ Compile, CompilerPlugin, IntegrationTest, names, Provided, Runtime, Test }
import sbt.librarymanagement.CrossVersion.{ binarySbtVersion, binaryScalaVersion, partialVersion }
import sbt.internal.util.complete._
import std.TaskExtra._
import sbt.internal.inc.{ Analysis, ClassfileManager, ClasspathOptions, CompilerCache, FileValueCache, IncOptions, Locate, LoggerReporter, MixedAnalyzingCompiler, ScalaInstance }
import testing.{ Framework, Runner, AnnotatedFingerprint, SubclassFingerprint }

import sbt.librarymanagement._
import sbt.internal.librarymanagement._
import sbt.internal.librarymanagement.syntax._
import sbt.internal.util._
import sbt.util.Level

import sys.error
import scala.xml.NodeSeq
import org.apache.ivy.core.module.{ descriptor, id }
import descriptor.ModuleDescriptor, id.ModuleRevisionId
import java.io.{ File, PrintWriter }
import java.net.{ URI, URL, MalformedURLException }
import java.util.concurrent.{ TimeUnit, Callable }
import sbinary.DefaultProtocol.StringFormat
import sbt.internal.util.Cache.seqFormat
import sbt.util.Logger
import CommandStrings.ExportStream

import sbt.internal.util.Types._

import sbt.internal.io.WatchState
import sbt.io.{ AllPassFilter, FileFilter, GlobFilter, HiddenFileFilter, IO, NameFilter, NothingFilter, Path, PathFinder }

import Path._
import Keys._

object Defaults extends BuildCommon {
  final val CacheDirectoryName = "cache"

  def configSrcSub(key: SettingKey[File]): Initialize[File] = (key in ThisScope.copy(config = Global), configuration) { (src, conf) => src / nameForSrc(conf.name) }
  def nameForSrc(config: String) = if (config == Configurations.Compile.name) "main" else config
  def prefix(config: String) = if (config == Configurations.Compile.name) "" else config + "-"

  def lock(app: xsbti.AppConfiguration): xsbti.GlobalLock = app.provider.scalaProvider.launcher.globalLock

  def extractAnalysis[T](a: Attributed[T]): (T, Analysis) =
    (a.data, a.metadata get Keys.analysis getOrElse Analysis.Empty)

  def analysisMap[T](cp: Seq[Attributed[T]]): T => Option[Analysis] =
    {
      val m = (for (a <- cp; an <- a.metadata get Keys.analysis) yield (a.data, an)).toMap
      m.get _
    }
  private[sbt] def globalDefaults(ss: Seq[Setting[_]]): Seq[Setting[_]] = Def.defaultSettings(inScope(GlobalScope)(ss))

  def buildCore: Seq[Setting[_]] = thisBuildCore ++ globalCore
  def thisBuildCore: Seq[Setting[_]] = inScope(GlobalScope.copy(project = Select(ThisBuild)))(Seq(
    managedDirectory := baseDirectory.value / "lib_managed"
  ))
  @deprecated("Use AutoPlugins and globalSbtCore instead.", "0.13.2")
  lazy val globalCore: Seq[Setting[_]] = globalDefaults(defaultTestTasks(test) ++ defaultTestTasks(testOnly) ++ defaultTestTasks(testQuick) ++ Seq(
    excludeFilter :== HiddenFileFilter
  ) ++ globalIvyCore ++ globalJvmCore) ++ globalSbtCore

  private[sbt] lazy val globalJvmCore: Seq[Setting[_]] =
    Seq(
      compilerCache := state.value get Keys.stateCompilerCache getOrElse CompilerCache.fresh,
      sourcesInBase :== true,
      autoAPIMappings := false,
      apiMappings := Map.empty,
      autoScalaLibrary :== true,
      managedScalaInstance :== true,
      definesClass :== FileValueCache(Locate.definesClass _).get,
      traceLevel in run :== 0,
      traceLevel in runMain :== 0,
      traceLevel in console :== Int.MaxValue,
      traceLevel in consoleProject :== Int.MaxValue,
      autoCompilerPlugins :== true,
      scalaHome :== None,
      apiURL := None,
      javaHome :== None,
      testForkedParallel :== false,
      javaOptions :== Nil,
      sbtPlugin :== false,
      crossPaths :== true,
      sourcePositionMappers :== Nil,
      artifactClassifier in packageSrc :== Some(SourceClassifier),
      artifactClassifier in packageDoc :== Some(DocClassifier),
      includeFilter :== NothingFilter,
      includeFilter in unmanagedSources :== "*.java" | "*.scala",
      includeFilter in unmanagedJars :== "*.jar" | "*.so" | "*.dll" | "*.jnilib" | "*.zip",
      includeFilter in unmanagedResources :== AllPassFilter
    )

  private[sbt] lazy val globalIvyCore: Seq[Setting[_]] =
    Seq(
      internalConfigurationMap :== Configurations.internalMap _,
      credentials :== Nil,
      exportJars :== false,
      retrieveManaged :== false,
      retrieveManagedSync :== false,
      configurationsToRetrieve :== None,
      scalaOrganization :== ScalaArtifacts.Organization,
      sbtResolver := { if (sbtVersion.value endsWith "-SNAPSHOT") Classpaths.typesafeSnapshots else Classpaths.typesafeReleases },
      crossVersion :== CrossVersion.Disabled,
      buildDependencies <<= Classpaths.constructBuildDependencies,
      version :== "0.1-SNAPSHOT",
      classpathTypes :== Set("jar", "bundle") ++ CustomPomParser.JarPackagings,
      artifactClassifier :== None,
      checksums := Classpaths.bootChecksums(appConfiguration.value),
      conflictManager := ConflictManager.default,
      pomExtra :== NodeSeq.Empty,
      pomPostProcess :== idFun,
      pomAllRepositories :== false,
      pomIncludeRepository :== Classpaths.defaultRepositoryFilter,
      updateOptions := UpdateOptions(),
      forceUpdatePeriod :== None
    )

  /** Core non-plugin settings for sbt builds.  These *must* be on every build or the sbt engine will fail to run at all. */
  private[sbt] lazy val globalSbtCore: Seq[Setting[_]] = globalDefaults(Seq(
    outputStrategy :== None, // TODO - This might belong elsewhere.
    buildStructure := Project.structure(state.value),
    settingsData := buildStructure.value.data,
    trapExit :== true,
    connectInput :== false,
    cancelable :== false,
    taskCancelStrategy := { state: State =>
      if (cancelable.value) TaskCancellationStrategy.Signal
      else TaskCancellationStrategy.Null
    },
    envVars :== Map.empty,
    sbtVersion := appConfiguration.value.provider.id.version,
    sbtBinaryVersion := binarySbtVersion(sbtVersion.value),
    watchingMessage := Watched.defaultWatchingMessage,
    triggeredMessage := Watched.defaultTriggeredMessage,
    onLoad := idFun[State],
    onUnload := idFun[State],
    onUnload := { s => try onUnload.value(s) finally IO.delete(taskTemporaryDirectory.value) },
    extraLoggers :== { _ => Nil },
    watchSources :== Nil,
    skip :== false,
    taskTemporaryDirectory := { val dir = IO.createTemporaryDirectory; dir.deleteOnExit(); dir },
    onComplete := { val dir = taskTemporaryDirectory.value; () => { IO.delete(dir); IO.createDirectory(dir) } },
    Previous.cache <<= Previous.cacheSetting,
    Previous.references :== new Previous.References,
    concurrentRestrictions <<= defaultRestrictions,
    parallelExecution :== true,
    pollInterval :== 500,
    logBuffered :== false,
    commands :== Nil,
    showSuccess :== true,
    showTiming :== true,
    timingFormat :== Aggregation.defaultFormat,
    aggregate :== true,
    maxErrors :== 100,
    fork :== false,
    initialize :== {},
    forcegc :== sys.props.get("sbt.task.forcegc").map(java.lang.Boolean.parseBoolean).getOrElse(GCUtil.defaultForceGarbageCollection),
    minForcegcInterval :== GCUtil.defaultMinForcegcInterval
  ))
  def defaultTestTasks(key: Scoped): Seq[Setting[_]] = inTask(key)(Seq(
    tags := Seq(Tags.Test -> 1),
    logBuffered := true
  ))
  // TODO: This should be on the new default settings for a project.
  def projectCore: Seq[Setting[_]] = Seq(
    name := thisProject.value.id,
    logManager := LogManager.defaults(extraLoggers.value, StandardMain.console),
    onLoadMessage <<= onLoadMessage or (name, thisProjectRef)("Set current project to " + _ + " (in build " + _.build + ")")
  )
  def paths = Seq(
    baseDirectory := thisProject.value.base,
    target := baseDirectory.value / "target",
    historyPath <<= historyPath or target(t => Some(t / ".history")),
    sourceDirectory := baseDirectory.value / "src",
    sourceManaged := crossTarget.value / "src_managed",
    resourceManaged := crossTarget.value / "resource_managed",
    cacheDirectory := crossTarget.value / CacheDirectoryName / thisProject.value.id / "global"
  )

  lazy val configPaths = sourceConfigPaths ++ resourceConfigPaths ++ outputConfigPaths
  lazy val sourceConfigPaths = Seq(
    sourceDirectory <<= configSrcSub(sourceDirectory),
    sourceManaged <<= configSrcSub(sourceManaged),
    scalaSource := sourceDirectory.value / "scala",
    javaSource := sourceDirectory.value / "java",
    unmanagedSourceDirectories := makeCrossSources(scalaSource.value, javaSource.value, scalaBinaryVersion.value, crossPaths.value),
    unmanagedSources <<= collectFiles(unmanagedSourceDirectories, includeFilter in unmanagedSources, excludeFilter in unmanagedSources),
    watchSources in ConfigGlobal <++= unmanagedSources,
    managedSourceDirectories := Seq(sourceManaged.value),
    managedSources <<= generate(sourceGenerators),
    sourceGenerators :== Nil,
    sourceDirectories <<= Classpaths.concatSettings(unmanagedSourceDirectories, managedSourceDirectories),
    sources <<= Classpaths.concat(unmanagedSources, managedSources)
  )
  lazy val resourceConfigPaths = Seq(
    resourceDirectory := sourceDirectory.value / "resources",
    resourceManaged <<= configSrcSub(resourceManaged),
    unmanagedResourceDirectories := Seq(resourceDirectory.value),
    managedResourceDirectories := Seq(resourceManaged.value),
    resourceDirectories <<= Classpaths.concatSettings(unmanagedResourceDirectories, managedResourceDirectories),
    unmanagedResources <<= collectFiles(unmanagedResourceDirectories, includeFilter in unmanagedResources, excludeFilter in unmanagedResources),
    watchSources in ConfigGlobal ++= unmanagedResources.value,
    resourceGenerators :== Nil,
    resourceGenerators <+= (discoveredSbtPlugins, resourceManaged) map PluginDiscovery.writeDescriptors,
    managedResources <<= generate(resourceGenerators),
    resources <<= Classpaths.concat(managedResources, unmanagedResources)
  )
  lazy val outputConfigPaths = Seq(
    cacheDirectory := crossTarget.value / CacheDirectoryName / thisProject.value.id / configuration.value.name,
    classDirectory := crossTarget.value / (prefix(configuration.value.name) + "classes"),
    target in doc := crossTarget.value / (prefix(configuration.value.name) + "api")
  )
  def addBaseSources = Seq(
    unmanagedSources := {
      val srcs = unmanagedSources.value
      val f = (includeFilter in unmanagedSources).value
      val excl = (excludeFilter in unmanagedSources).value
      if (sourcesInBase.value) (srcs +++ baseDirectory.value * (f -- excl)).get else srcs
    }
  )

  def compileBase = inTask(console)(compilersSetting :: Nil) ++ compileBaseGlobal ++ Seq(
    incOptions := incOptions.value.withNewClassfileManager(
      ClassfileManager.transactional(crossTarget.value / "classes.bak", sbt.util.Logger.Null)),
    scalaInstance <<= scalaInstanceTask,
    crossVersion := (if (crossPaths.value) CrossVersion.binary else CrossVersion.Disabled),
    crossTarget := makeCrossTarget(target.value, scalaBinaryVersion.value, sbtBinaryVersion.value, sbtPlugin.value, crossPaths.value),
    clean := {
      val _ = clean.value
      IvyActions.cleanCachedResolutionCache(ivyModule.value, streams.value.log)
    },
    scalaCompilerBridgeSource := Compiler.defaultCompilerBridgeSource(scalaVersion.value)
  )
  // must be a val: duplication detected by object identity
  private[this] lazy val compileBaseGlobal: Seq[Setting[_]] = globalDefaults(Seq(
    incOptions := IncOptions.Default,
    classpathOptions :== ClasspathOptions.boot,
    classpathOptions in console :== ClasspathOptions.repl,
    compileOrder :== CompileOrder.Mixed,
    javacOptions :== Nil,
    scalacOptions :== Nil,
    scalaVersion := appConfiguration.value.provider.scalaProvider.version,
    derive(crossScalaVersions := Seq(scalaVersion.value)),
    derive(compilersSetting),
    derive(scalaBinaryVersion := binaryScalaVersion(scalaVersion.value))
  ))

  def makeCrossSources(scalaSrcDir: File, javaSrcDir: File, sv: String, cross: Boolean): Seq[File] = {
    if (cross)
      Seq(scalaSrcDir.getParentFile / s"${scalaSrcDir.name}-$sv", scalaSrcDir, javaSrcDir)
    else
      Seq(scalaSrcDir, javaSrcDir)
  }

  def makeCrossTarget(t: File, sv: String, sbtv: String, plugin: Boolean, cross: Boolean): File =
    {
      val scalaBase = if (cross) t / ("scala-" + sv) else t
      if (plugin) scalaBase / ("sbt-" + sbtv) else scalaBase
    }

  def compilersSetting = compilers := Compiler.compilers(scalaInstance.value, classpathOptions.value, javaHome.value, ivyConfiguration.value, scalaCompilerBridgeSource.value)(appConfiguration.value, streams.value.log)

  lazy val configTasks = docTaskSettings(doc) ++ inTask(compile)(compileInputsSettings) ++ configGlobal ++ compileAnalysisSettings ++ Seq(
    compile <<= compileTask,
    manipulateBytecode := compileIncremental.value,
    compileIncremental <<= compileIncrementalTask tag (Tags.Compile, Tags.CPU),
    printWarnings <<= printWarningsTask,
    compileAnalysisFilename := {
      // Here, if the user wants cross-scala-versioning, we also append it
      // to the analysis cache, so we keep the scala versions separated.
      val extra =
        if (crossPaths.value) s"_${scalaBinaryVersion.value}"
        else ""
      s"inc_compile${extra}"
    },
    compileIncSetup <<= compileIncSetupTask,
    console <<= consoleTask,
    consoleQuick <<= consoleQuickTask,
    discoveredMainClasses <<= compile map discoverMainClasses storeAs discoveredMainClasses xtriggeredBy compile,
    definedSbtPlugins <<= discoverPlugins,
    discoveredSbtPlugins <<= discoverSbtPluginNames,
    inTask(run)(runnerTask :: Nil).head,
    selectMainClass := mainClass.value orElse askForMainClass(discoveredMainClasses.value),
    mainClass in run := (selectMainClass in run).value,
    mainClass := pickMainClassOrWarn(discoveredMainClasses.value, streams.value.log),
    run <<= runTask(fullClasspath, mainClass in run, runner in run),
    runMain <<= runMainTask(fullClasspath, runner in run),
    copyResources <<= copyResourcesTask
  )
  private[this] lazy val configGlobal = globalDefaults(Seq(
    initialCommands :== "",
    cleanupCommands :== ""
  ))

  lazy val projectTasks: Seq[Setting[_]] = Seq(
    cleanFiles := Seq(managedDirectory.value, target.value),
    cleanKeepFiles := historyPath.value.toList,
    clean := doClean(cleanFiles.value, cleanKeepFiles.value),
    consoleProject <<= consoleProjectTask,
    watchTransitiveSources <<= watchTransitiveSourcesTask,
    watch <<= watchSetting
  )

  def generate(generators: SettingKey[Seq[Task[Seq[File]]]]): Initialize[Task[Seq[File]]] = generators { _.join.map(_.flatten) }

  @deprecated("Use the new <key>.all(<ScopeFilter>) API", "0.13.0")
  def inAllConfigurations[T](key: TaskKey[T]): Initialize[Task[Seq[T]]] = (state, thisProjectRef) flatMap { (state, ref) =>
    val structure = Project structure state
    val configurations = Project.getProject(ref, structure).toList.flatMap(_.configurations)
    configurations.flatMap { conf =>
      key in (ref, conf) get structure.data
    } join
  }
  def watchTransitiveSourcesTask: Initialize[Task[Seq[File]]] = {
    import ScopeFilter.Make.{ inDependencies => inDeps, _ }
    val selectDeps = ScopeFilter(inAggregates(ThisProject) || inDeps(ThisProject))
    val allWatched = (watchSources ?? Nil).all(selectDeps)
    Def.task { allWatched.value.flatten }
  }

  def transitiveUpdateTask: Initialize[Task[Seq[UpdateReport]]] = {
    import ScopeFilter.Make.{ inDependencies => inDeps, _ }
    val selectDeps = ScopeFilter(inDeps(ThisProject, includeRoot = false))
    val allUpdates = update.?.all(selectDeps)
    // If I am a "build" (a project inside project/) then I have a globalPluginUpdate.
    Def.task { allUpdates.value.flatten ++ globalPluginUpdate.?.value }
  }

  def watchSetting: Initialize[Watched] = (pollInterval, thisProjectRef, watchingMessage, triggeredMessage) { (interval, base, msg, trigMsg) =>
    new Watched {
      val scoped = watchTransitiveSources in base
      val key = ScopedKey(scoped.scope, scoped.key)
      override def pollInterval = interval
      override def watchingMessage(s: WatchState) = msg(s)
      override def triggeredMessage(s: WatchState) = trigMsg(s)
      override def watchPaths(s: State) = EvaluateTask.evaluateTask(Project structure s, key, s, base) match {
        case Some(Value(ps)) => ps
        case Some(Inc(i))    => throw i
        case None            => sys.error("key not found: " + Def.displayFull(key))
      }
    }
  }

  @deprecated("Use scalaInstanceTask.", "0.13.0")
  def scalaInstanceSetting = scalaInstanceTask
  def scalaInstanceTask: Initialize[Task[ScalaInstance]] = Def.taskDyn {
    // if this logic changes, ensure that `unmanagedScalaInstanceOnly` and `update` are changed
    //  appropriately to avoid cycles
    scalaHome.value match {
      case Some(h) => scalaInstanceFromHome(h)
      case None =>
        val scalaProvider = appConfiguration.value.provider.scalaProvider
        val version = scalaVersion.value
        if (version == scalaProvider.version) // use the same class loader as the Scala classes used by sbt
          Def.task(ScalaInstance(version, scalaProvider))
        else
          scalaInstanceFromUpdate
    }
  }
  // Returns the ScalaInstance only if it was not constructed via `update`
  //  This is necessary to prevent cycles between `update` and `scalaInstance`
  private[sbt] def unmanagedScalaInstanceOnly: Initialize[Task[Option[ScalaInstance]]] = Def.taskDyn {
    if (scalaHome.value.isDefined) Def.task(Some(scalaInstance.value)) else Def.task(None)
  }

  private[this] def noToolConfiguration(autoInstance: Boolean): String =
    {
      val pre = "Missing Scala tool configuration from the 'update' report.  "
      val post =
        if (autoInstance)
          "'scala-tool' is normally added automatically, so this may indicate a bug in sbt or you may be removing it from ivyConfigurations, for example."
        else
          "Explicitly define scalaInstance or scalaHome or include Scala dependencies in the 'scala-tool' configuration."
      pre + post
    }

  def scalaInstanceFromUpdate: Initialize[Task[ScalaInstance]] = Def.task {
    val toolReport = update.value.configuration(Configurations.ScalaTool.name) getOrElse
      sys.error(noToolConfiguration(managedScalaInstance.value))
    def files(id: String) =
      for {
        m <- toolReport.modules if m.module.name == id;
        (art, file) <- m.artifacts if art.`type` == Artifact.DefaultType
      } yield file
    def file(id: String) = files(id).headOption getOrElse sys.error(s"Missing ${id}.jar")
    val allFiles = toolReport.modules.flatMap(_.artifacts.map(_._2))
    val libraryJar = file(ScalaArtifacts.LibraryID)
    val compilerJar = file(ScalaArtifacts.CompilerID)
    val otherJars = allFiles.filterNot(x => x == libraryJar || x == compilerJar)
    new ScalaInstance(scalaVersion.value, makeClassLoader(state.value)(libraryJar :: compilerJar :: otherJars.toList), libraryJar, compilerJar, otherJars.toArray, None)
  }
  def scalaInstanceFromHome(dir: File): Initialize[Task[ScalaInstance]] = Def.task {
    ScalaInstance(dir)(makeClassLoader(state.value))
  }
  private[this] def makeClassLoader(state: State) = state.classLoaderCache.apply _

  private[this] def testDefaults = Defaults.globalDefaults(Seq(
    testFrameworks :== {
      import sbt.TestFrameworks._
      Seq(ScalaCheck, Specs2, Specs, ScalaTest, JUnit)
    },
    testListeners :== Nil,
    testOptions :== Nil,
    testResultLogger :== TestResultLogger.Default,
    testFilter in testOnly :== (selectedFilter _)
  ))
  lazy val testTasks: Seq[Setting[_]] = testTaskOptions(test) ++ testTaskOptions(testOnly) ++ testTaskOptions(testQuick) ++ testDefaults ++ Seq(
    testLoader := TestFramework.createTestLoader(data(fullClasspath.value), scalaInstance.value, IO.createUniqueDirectory(taskTemporaryDirectory.value)),
    loadedTestFrameworks := testFrameworks.value.flatMap(f => f.create(testLoader.value, streams.value.log).map(x => (f, x)).toIterable).toMap,
    definedTests <<= detectTests,
    definedTestNames <<= definedTests map (_.map(_.name).distinct) storeAs definedTestNames triggeredBy compile,
    testFilter in testQuick <<= testQuickFilter,
    executeTests <<= (streams in test, loadedTestFrameworks, testLoader, testGrouping in test, testExecution in test, fullClasspath in test, javaHome in test, testForkedParallel, javaOptions in test) flatMap allTestGroupsTask,
    testResultLogger in (Test, test) :== TestResultLogger.SilentWhenNoTests, // https://github.com/sbt/sbt/issues/1185
    test := {
      val trl = (testResultLogger in (Test, test)).value
      val taskName = Project.showContextKey(state.value)(resolvedScoped.value)
      trl.run(streams.value.log, executeTests.value, taskName)
    },
    testOnly <<= inputTests(testOnly),
    testQuick <<= inputTests(testQuick)
  )
  lazy val TaskGlobal: Scope = ThisScope.copy(task = Global)
  lazy val ConfigGlobal: Scope = ThisScope.copy(config = Global)
  def testTaskOptions(key: Scoped): Seq[Setting[_]] = inTask(key)(Seq(
    testListeners := {
      TestLogger.make(streams.value.log, closeableTestLogger(streamsManager.value, test in resolvedScoped.value.scope, logBuffered.value)) +:
        new TestStatusReporter(succeededFile(streams.in(test).value.cacheDirectory)) +:
        testListeners.in(TaskGlobal).value
    },
    testOptions := Tests.Listeners(testListeners.value) +: (testOptions in TaskGlobal).value,
    testExecution <<= testExecutionTask(key)
  )) ++ inScope(GlobalScope)(Seq(
    derive(testGrouping <<= singleTestGroupDefault)
  ))
  @deprecated("Doesn't provide for closing the underlying resources.", "0.13.1")
  def testLogger(manager: Streams, baseKey: Scoped)(tdef: TestDefinition): Logger =
    {
      val scope = baseKey.scope
      val extra = scope.extra match { case Select(x) => x; case _ => AttributeMap.empty }
      val key = ScopedKey(scope.copy(extra = Select(testExtra(extra, tdef))), baseKey.key)
      manager(key).log
    }
  private[this] def closeableTestLogger(manager: Streams, baseKey: Scoped, buffered: Boolean)(tdef: TestDefinition): TestLogger.PerTest =
    {
      val scope = baseKey.scope
      val extra = scope.extra match { case Select(x) => x; case _ => AttributeMap.empty }
      val key = ScopedKey(scope.copy(extra = Select(testExtra(extra, tdef))), baseKey.key)
      val s = manager(key)
      new TestLogger.PerTest(s.log, () => s.close(), buffered)
    }
  def buffered(log: Logger): Logger = new BufferedLogger(FullLogger(log))
  def testExtra(extra: AttributeMap, tdef: TestDefinition): AttributeMap =
    {
      val mod = tdef.fingerprint match { case f: SubclassFingerprint => f.isModule; case f: AnnotatedFingerprint => f.isModule; case _ => false }
      extra.put(name.key, tdef.name).put(isModule, mod)
    }
  def singleTestGroup(key: Scoped): Initialize[Task[Seq[Tests.Group]]] = inTask(key, singleTestGroupDefault)
  def singleTestGroupDefault: Initialize[Task[Seq[Tests.Group]]] = Def.task {
    val tests = definedTests.value
    val fk = fork.value
    val opts = forkOptions.value
    Seq(new Tests.Group("<default>", tests, if (fk) Tests.SubProcess(opts) else Tests.InProcess))
  }
  private[this] def forkOptions: Initialize[Task[ForkOptions]] =
    (baseDirectory, javaOptions, outputStrategy, envVars, javaHome, connectInput) map {
      (base, options, strategy, env, javaHomeDir, connectIn) =>
        // bootJars is empty by default because only jars on the user's classpath should be on the boot classpath
        ForkOptions(bootJars = Nil, javaHome = javaHomeDir, connectInput = connectIn, outputStrategy = strategy, runJVMOptions = options, workingDirectory = Some(base), envVars = env)
    }

  def testExecutionTask(task: Scoped): Initialize[Task[Tests.Execution]] =
    (testOptions in task, parallelExecution in task, tags in task) map {
      (opts, par, ts) =>
        new Tests.Execution(opts, par, ts)
    }

  def testQuickFilter: Initialize[Task[Seq[String] => Seq[String => Boolean]]] =
    (fullClasspath in test, streams in test) map {
      (cp, s) =>
        val ans = cp.flatMap(_.metadata get Keys.analysis)
        val succeeded = TestStatus.read(succeededFile(s.cacheDirectory))
        val stamps = collection.mutable.Map.empty[File, Long]
        def stamp(dep: String): Long = {
          val stamps = for (a <- ans; f <- a.relations.definesClass(dep)) yield intlStamp(f, a, Set.empty)
          if (stamps.isEmpty) Long.MinValue else stamps.max
        }
        def intlStamp(f: File, analysis: Analysis, s: Set[File]): Long = {
          if (s contains f) Long.MinValue else
            stamps.getOrElseUpdate(f, {
              import analysis.{ relations => rel, apis }
              rel.internalSrcDeps(f).map(intlStamp(_, analysis, s + f)) ++
                rel.externalDeps(f).map(stamp) +
                apis.internal(f).compilation.startTime
            }.max)
        }
        def noSuccessYet(test: String) = succeeded.get(test) match {
          case None     => true
          case Some(ts) => stamp(test) > ts
        }

        args => for (filter <- selectedFilter(args)) yield (test: String) => filter(test) && noSuccessYet(test)
    }
  def succeededFile(dir: File) = dir / "succeeded_tests"

  def inputTests(key: InputKey[_]): Initialize[InputTask[Unit]] = inputTests0.mapReferenced(Def.mapScope(_ in key.key))
  private[this] lazy val inputTests0: Initialize[InputTask[Unit]] =
    {
      val parser = loadForParser(definedTestNames)((s, i) => testOnlyParser(s, i getOrElse Nil))
      Def.inputTaskDyn {
        val (selected, frameworkOptions) = parser.parsed
        val s = streams.value
        val filter = testFilter.value
        val config = testExecution.value

        implicit val display = Project.showContextKey(state.value)
        val modifiedOpts = Tests.Filters(filter(selected)) +: Tests.Argument(frameworkOptions: _*) +: config.options
        val newConfig = config.copy(options = modifiedOpts)
        val output = allTestGroupsTask(s, loadedTestFrameworks.value, testLoader.value, testGrouping.value, newConfig, fullClasspath.value, javaHome.value, testForkedParallel.value, javaOptions.value)
        val taskName = display(resolvedScoped.value)
        val trl = testResultLogger.value
        val processed = output.map(out => trl.run(s.log, out, taskName))
        Def.value(processed)
      }
    }

  def createTestRunners(frameworks: Map[TestFramework, Framework], loader: ClassLoader, config: Tests.Execution): Map[TestFramework, Runner] = {
    import Tests.Argument
    val opts = config.options.toList
    frameworks.map {
      case (tf, f) =>
        val args = opts.flatMap {
          case Argument(None | Some(`tf`), args) => args
          case _                                 => Nil
        }
        val mainRunner = f.runner(args.toArray, Array.empty[String], loader)
        tf -> mainRunner
    }
  }

  def allTestGroupsTask(s: TaskStreams, frameworks: Map[TestFramework, Framework], loader: ClassLoader, groups: Seq[Tests.Group], config: Tests.Execution, cp: Classpath, javaHome: Option[File]): Task[Tests.Output] = {
    allTestGroupsTask(s, frameworks, loader, groups, config, cp, javaHome, forkedParallelExecution = false, javaOptions = Nil)
  }

  def allTestGroupsTask(s: TaskStreams, frameworks: Map[TestFramework, Framework], loader: ClassLoader, groups: Seq[Tests.Group], config: Tests.Execution, cp: Classpath, javaHome: Option[File], forkedParallelExecution: Boolean): Task[Tests.Output] = {
    allTestGroupsTask(s, frameworks, loader, groups, config, cp, javaHome, forkedParallelExecution, javaOptions = Nil)
  }

  def allTestGroupsTask(s: TaskStreams, frameworks: Map[TestFramework, Framework], loader: ClassLoader, groups: Seq[Tests.Group], config: Tests.Execution, cp: Classpath, javaHome: Option[File], forkedParallelExecution: Boolean, javaOptions: Seq[String]): Task[Tests.Output] = {
    val runners = createTestRunners(frameworks, loader, config)
    val groupTasks = groups map {
      case Tests.Group(name, tests, runPolicy) =>
        runPolicy match {
          case Tests.SubProcess(opts) =>
            s.log.debug(s"javaOptions: ${opts.runJVMOptions}")
            val forkedConfig = config.copy(parallel = config.parallel && forkedParallelExecution)
            s.log.debug(s"Forking tests - parallelism = ${forkedConfig.parallel}")
            ForkTests(runners, tests.toList, forkedConfig, cp.files, opts, s.log) tag Tags.ForkedTestGroup
          case Tests.InProcess =>
            if (javaOptions.nonEmpty) {
              s.log.warn("javaOptions will be ignored, fork is set to false")
            }
            Tests(frameworks, loader, runners, tests, config, s.log)
        }
    }
    val output = Tests.foldTasks(groupTasks, config.parallel)
    output map { out =>
      val summaries =
        runners map {
          case (tf, r) =>
            Tests.Summary(frameworks(tf).name, r.done())
        }
      out.copy(summaries = summaries)
    }
  }

  def selectedFilter(args: Seq[String]): Seq[String => Boolean] =
    {
      def matches(nfs: Seq[NameFilter], s: String) = nfs.exists(_.accept(s))

      val (excludeArgs, includeArgs) = args.partition(_.startsWith("-"))

      val includeFilters = includeArgs map GlobFilter.apply
      val excludeFilters = excludeArgs.map(_.substring(1)).map(GlobFilter.apply)

      if (includeFilters.isEmpty && excludeArgs.isEmpty) {
        Seq(const(true))
      } else if (includeFilters.isEmpty) {
        Seq({ (s: String) => !matches(excludeFilters, s) })
      } else {
        includeFilters.map { f => (s: String) => (f.accept(s) && !matches(excludeFilters, s)) }
      }
    }
  def detectTests: Initialize[Task[Seq[TestDefinition]]] = (loadedTestFrameworks, compile, streams) map { (frameworkMap, analysis, s) =>
    Tests.discover(frameworkMap.values.toList, analysis, s.log)._1
  }
  def defaultRestrictions: Initialize[Seq[Tags.Rule]] = parallelExecution { par =>
    val max = EvaluateTask.SystemProcessors
    Tags.limitAll(if (par) max else 1) :: Tags.limit(Tags.ForkedTestGroup, 1) :: Nil
  }

  lazy val packageBase: Seq[Setting[_]] = Seq(
    artifact := Artifact(moduleName.value)
  ) ++ Defaults.globalDefaults(Seq(
      packageOptions :== Nil,
      artifactName :== (Artifact.artifactName _)
    ))

  lazy val packageConfig: Seq[Setting[_]] =
    inTask(packageBin)(Seq(
      packageOptions <<= (name, version, homepage, organization, organizationName, mainClass, packageOptions) map { (name, ver, h, org, orgName, main, p) => Package.addSpecManifestAttributes(name, ver, orgName) +: Package.addImplManifestAttributes(name, ver, h, org, orgName) +: main.map(Package.MainClass.apply) ++: p })) ++
      inTask(packageSrc)(Seq(
        packageOptions := Package.addSpecManifestAttributes(name.value, version.value, organizationName.value) +: packageOptions.value)) ++
      packageTaskSettings(packageBin, packageBinMappings) ++
      packageTaskSettings(packageSrc, packageSrcMappings) ++
      packageTaskSettings(packageDoc, packageDocMappings) ++
      Seq(`package` := packageBin.value)

  def packageBinMappings = products map { _ flatMap Path.allSubpaths }
  def packageDocMappings = doc map { Path.allSubpaths(_).toSeq }
  def packageSrcMappings = concatMappings(resourceMappings, sourceMappings)

  @deprecated("Use `packageBinMappings` instead", "0.12.0")
  def packageBinTask = packageBinMappings
  @deprecated("Use `packageDocMappings` instead", "0.12.0")
  def packageDocTask = packageDocMappings
  @deprecated("Use `packageSrcMappings` instead", "0.12.0")
  def packageSrcTask = packageSrcMappings

  private type Mappings = Initialize[Task[Seq[(File, String)]]]
  def concatMappings(as: Mappings, bs: Mappings) = (as zipWith bs)((a, b) => (a, b) map { case (a, b) => a ++ b })

  // drop base directories, since there are no valid mappings for these
  def sourceMappings = (unmanagedSources, unmanagedSourceDirectories, baseDirectory) map { (srcs, sdirs, base) =>
    (srcs --- sdirs --- base) pair (relativeTo(sdirs) | relativeTo(base) | flat)
  }
  def resourceMappings = relativeMappings(unmanagedResources, unmanagedResourceDirectories)
  def relativeMappings(files: ScopedTaskable[Seq[File]], dirs: ScopedTaskable[Seq[File]]): Initialize[Task[Seq[(File, String)]]] =
    (files, dirs) map { (rs, rdirs) =>
      (rs --- rdirs) pair (relativeTo(rdirs) | flat)
    }

  def collectFiles(dirs: ScopedTaskable[Seq[File]], filter: ScopedTaskable[FileFilter], excludes: ScopedTaskable[FileFilter]): Initialize[Task[Seq[File]]] =
    (dirs, filter, excludes) map { (d, f, excl) => d.descendantsExcept(f, excl).get }

  def artifactPathSetting(art: SettingKey[Artifact]) = (crossTarget, projectID, art, scalaVersion in artifactName, scalaBinaryVersion in artifactName, artifactName) {
    (t, module, a, sv, sbv, toString) =>
      t / toString(ScalaVersion(sv, sbv), module, a) asFile
  }
  def artifactSetting = ((artifact, artifactClassifier).identity zipWith configuration.?) {
    case ((a, classifier), cOpt) =>
      val cPart = cOpt flatMap {
        case Compile => None
        case Test    => Some(Artifact.TestsClassifier)
        case c       => Some(c.name)
      }
      val combined = cPart.toList ++ classifier.toList
      if (combined.isEmpty) a.copy(classifier = None, configurations = cOpt.toList) else {
        val classifierString = combined mkString "-"
        val confs = cOpt.toList flatMap { c => artifactConfigurations(a, c, classifier) }
        a.copy(classifier = Some(classifierString), `type` = Artifact.classifierType(classifierString), configurations = confs)
      }
  }
  def artifactConfigurations(base: Artifact, scope: Configuration, classifier: Option[String]): Iterable[Configuration] =
    classifier match {
      case Some(c) => Artifact.classifierConf(c) :: Nil
      case None    => scope :: Nil
    }

  @deprecated("Use `Util.pairID` instead", "0.12.0")
  def pairID = Util.pairID

  @deprecated("Use the cacheDirectory val on streams.", "0.13.0")
  def perTaskCache(key: TaskKey[_]): Setting[File] =
    cacheDirectory ~= { _ / ("for_" + key.key.label) }

  @deprecated("Use `packageTaskSettings` instead", "0.12.0")
  def packageTasks(key: TaskKey[File], mappingsTask: Initialize[Task[Seq[(File, String)]]]) = packageTaskSettings(key, mappingsTask)
  def packageTaskSettings(key: TaskKey[File], mappingsTask: Initialize[Task[Seq[(File, String)]]]) =
    inTask(key)(Seq(
      key in TaskGlobal <<= packageTask,
      packageConfiguration <<= packageConfigurationTask,
      mappings <<= mappingsTask,
      packagedArtifact := (artifact.value -> key.value),
      artifact <<= artifactSetting,
      artifactPath <<= artifactPathSetting(artifact)
    ))
  def packageTask: Initialize[Task[File]] =
    (packageConfiguration, streams) map { (config, s) =>
      Package(config, s.cacheDirectory, s.log)
      config.jar
    }
  def packageConfigurationTask: Initialize[Task[Package.Configuration]] =
    (mappings, artifactPath, packageOptions) map { (srcs, path, options) =>
      new Package.Configuration(srcs, path, options)
    }

  @deprecated("use Defaults.askForMainClass", "0.13.7")
  def selectRunMain(classes: Seq[String]): Option[String] = askForMainClass(classes)
  @deprecated("use Defaults.pickMainClass", "0.13.7")
  def selectPackageMain(classes: Seq[String]): Option[String] = pickMainClass(classes)
  def askForMainClass(classes: Seq[String]): Option[String] =
    sbt.SelectMainClass(Some(SimpleReader readLine _), classes)
  def pickMainClass(classes: Seq[String]): Option[String] =
    sbt.SelectMainClass(None, classes)
  private def pickMainClassOrWarn(classes: Seq[String], logger: Logger): Option[String] = {
    classes match {
      case multiple if multiple.size > 1 => logger.warn("Multiple main classes detected.  Run 'show discoveredMainClasses' to see the list")
      case _                             =>
    }
    pickMainClass(classes)
  }

  def doClean(clean: Seq[File], preserve: Seq[File]): Unit =
    IO.withTemporaryDirectory { temp =>
      val (dirs, files) = preserve.filter(_.exists).flatMap(_.allPaths.get).partition(_.isDirectory)
      val mappings = files.zipWithIndex map { case (f, i) => (f, new File(temp, i.toHexString)) }
      IO.move(mappings)
      IO.delete(clean)
      IO.createDirectories(dirs) // recreate empty directories
      IO.move(mappings.map(_.swap))
    }
  def runMainTask(classpath: Initialize[Task[Classpath]], scalaRun: Initialize[Task[ScalaRun]]): Initialize[InputTask[Unit]] =
    {
      import DefaultParsers._
      val parser = loadForParser(discoveredMainClasses)((s, names) => runMainParser(s, names getOrElse Nil))
      Def.inputTask {
        val (mainClass, args) = parser.parsed
        toError(scalaRun.value.run(mainClass, data(classpath.value), args, streams.value.log))
      }
    }

  def runTask(classpath: Initialize[Task[Classpath]], mainClassTask: Initialize[Task[Option[String]]], scalaRun: Initialize[Task[ScalaRun]]): Initialize[InputTask[Unit]] =
    {
      import Def.parserToInput
      val parser = Def.spaceDelimited()
      Def.inputTask {
        val mainClass = mainClassTask.value getOrElse sys.error("No main class detected.")
        toError(scalaRun.value.run(mainClass, data(classpath.value), parser.parsed, streams.value.log))
      }
    }

  def runnerTask = runner <<= runnerInit
  def runnerInit: Initialize[Task[ScalaRun]] = Def.task {
    val tmp = taskTemporaryDirectory.value
    val resolvedScope = resolvedScoped.value.scope
    val structure = buildStructure.value
    val si = scalaInstance.value
    val s = streams.value
    val options = javaOptions.value
    if (fork.value) {
      s.log.debug(s"javaOptions: $options")
      new ForkRun(forkOptions.value)
    } else {
      if (options.nonEmpty) {
        val mask = ScopeMask(project = false)
        val showJavaOptions = Scope.displayMasked((javaOptions in resolvedScope).scopedKey.scope, (javaOptions in resolvedScope).key.label, mask)
        val showFork = Scope.displayMasked((fork in resolvedScope).scopedKey.scope, (fork in resolvedScope).key.label, mask)
        s.log.warn(s"$showJavaOptions will be ignored, $showFork is set to false")
      }
      new Run(si, trapExit.value, tmp)
    }
  }

  @deprecated("Use `docTaskSettings` instead", "0.12.0")
  def docSetting(key: TaskKey[File]) = docTaskSettings(key)
  def docTaskSettings(key: TaskKey[File] = doc): Seq[Setting[_]] = inTask(key)(Seq(
    apiMappings ++= { if (autoAPIMappings.value) APIMappings.extract(dependencyClasspath.value, streams.value.log).toMap else Map.empty[File, URL] },
    fileInputOptions := Seq("-doc-root-content", "-diagrams-dot-path"),
    key in TaskGlobal := {
      val s = streams.value
      val cs = compilers.value
      val srcs = sources.value
      val out = target.value
      val sOpts = scalacOptions.value
      val xapis = apiMappings.value
      val hasScala = srcs.exists(_.name.endsWith(".scala"))
      val hasJava = srcs.exists(_.name.endsWith(".java"))
      val cp = data(dependencyClasspath.value).toList
      val label = nameForSrc(configuration.value.name)
      val fiOpts = fileInputOptions.value
      val logger: Logger = s.log
      val maxer = maxErrors.value
      val spms = sourcePositionMappers.value
      val reporter: xsbti.Reporter =
        (compilerReporter in compile).value match {
          case Some(r) => r
          case _       => new LoggerReporter(maxer, logger, Compiler.foldMappers(spms))
        }
      (hasScala, hasJava) match {
        case (true, _) =>
          val options = sOpts ++ Opts.doc.externalAPI(xapis)
          val runDoc = Doc.scaladoc(label, s.cacheDirectory / "scala", cs.scalac.onArgs(exported(s, "scaladoc")), fiOpts)
          runDoc(srcs, cp, out, options, maxErrors.value, s.log)
        case (_, true) =>
          val javadoc = sbt.inc.Doc.cachedJavadoc(label, s.cacheDirectory / "java", cs.javac)
          javadoc.run(srcs.toList, cp, out, javacOptions.value.toList, s.log, reporter)
        case _ => () // do nothing
      }
      out
    }
  ))

  def mainRunTask = run <<= runTask(fullClasspath in Runtime, mainClass in run, runner in run)
  def mainRunMainTask = runMain <<= runMainTask(fullClasspath in Runtime, runner in run)

  def discoverMainClasses(analysis: Analysis): Seq[String] =
    Discovery.applications(Tests.allDefs(analysis)).collect({ case (definition, discovered) if discovered.hasMain => definition.name }).sorted

  def consoleProjectTask = (state, streams, initialCommands in consoleProject) map { (state, s, extra) => ConsoleProject(state, extra)(s.log); println() }
  def consoleTask: Initialize[Task[Unit]] = consoleTask(fullClasspath, console)
  def consoleQuickTask = consoleTask(externalDependencyClasspath, consoleQuick)
  def consoleTask(classpath: TaskKey[Classpath], task: TaskKey[_]): Initialize[Task[Unit]] =
    (compilers in task, classpath in task, scalacOptions in task, initialCommands in task, cleanupCommands in task, taskTemporaryDirectory in task, scalaInstance in task, streams) map {
      (cs, cp, options, initCommands, cleanup, temp, si, s) =>
        val cpFiles = data(cp)
        val fullcp = (cpFiles ++ si.allJars).distinct
        val loader = sbt.internal.inc.classpath.ClasspathUtilities.makeLoader(fullcp, si, IO.createUniqueDirectory(temp))
        val compiler = cs.scalac.onArgs(exported(s, "scala"))
        (new Console(compiler))(cpFiles, options, loader, initCommands, cleanup)()(s.log).foreach(msg => sys.error(msg))
        println()
    }

  private[this] def exported(w: PrintWriter, command: String): Seq[String] => Unit = args =>
    w.println((command +: args).mkString(" "))
  private[this] def exported(s: TaskStreams, command: String): Seq[String] => Unit = args => {
    val w = s.text(ExportStream)
    try exported(w, command)
    finally w.close() // workaround for #937
  }

  @deprecated("Use inTask(compile)(compileInputsSettings)", "0.13.0")
  def compileTaskSettings: Seq[Setting[_]] = inTask(compile)(compileInputsSettings)

  def compileTask: Initialize[Task[Analysis]] = Def.task {
    val setup: Compiler.IncSetup = compileIncSetup.value
    // TODO - expose bytecode manipulation phase.
    val analysisResult: Compiler.CompileResult = manipulateBytecode.value
    if (analysisResult.hasModified) {
      val store = MixedAnalyzingCompiler.staticCachedStore(setup.cacheFile)
      store.set(analysisResult.analysis, analysisResult.setup)
    }
    analysisResult.analysis
  }
  def compileIncrementalTask = Def.task {
    // TODO - Should readAnalysis + saveAnalysis be scoped by the compile task too?
    compileIncrementalTaskImpl(streams.value, (compileInputs in compile).value, previousCompile.value, (compilerReporter in compile).value)
  }
  private[this] def compileIncrementalTaskImpl(s: TaskStreams, ci: Compiler.Inputs, previous: Compiler.PreviousAnalysis, reporter: Option[xsbti.Reporter]): Compiler.CompileResult =
    {
      lazy val x = s.text(ExportStream)
      def onArgs(cs: Compiler.Compilers) = cs.copy(scalac = cs.scalac.onArgs(exported(x, "scalac")), javac = cs.javac /*.onArgs(exported(x, "javac"))*/ )
      val i = InputsWithPrevious(ci.copy(compilers = onArgs(ci.compilers)), previous)
      try reporter match {
        case Some(reporter) => Compiler.compile(i, s.log, reporter)
        case None           => Compiler.compile(i, s.log)
      }
      finally x.close() // workaround for #937
    }
  def compileIncSetupTask = Def.task {
    Compiler.IncSetup(
      analysisMap(dependencyClasspath.value),
      definesClass.value,
      (skip in compile).value,
      // TODO - this is kind of a bad way to grab the cache directory for streams...
      streams.value.cacheDirectory / compileAnalysisFilename.value,
      compilerCache.value,
      incOptions.value)
  }
  def compileInputsSettings: Seq[Setting[_]] =
    Seq(compileInputs := {
      val cp = classDirectory.value +: data(dependencyClasspath.value)
      Compiler.inputs(cp, sources.value, classDirectory.value, scalacOptions.value, javacOptions.value,
        maxErrors.value, sourcePositionMappers.value, compileOrder.value)(compilers.value, compileIncSetup.value, streams.value.log)
    },
      compilerReporter := None)
  def compileAnalysisSettings: Seq[Setting[_]] = Seq(
    previousCompile := {
      val setup: Compiler.IncSetup = compileIncSetup.value
      val store = MixedAnalyzingCompiler.staticCachedStore(setup.cacheFile)
      store.get() match {
        case Some((an, setup)) => Compiler.PreviousAnalysis(an, Some(setup))
        case None              => Compiler.PreviousAnalysis(Analysis.empty(nameHashing = setup.incOptions.nameHashing), None)
      }
    }
  )

  def printWarningsTask: Initialize[Task[Unit]] =
    (streams, compile, maxErrors, sourcePositionMappers) map { (s, analysis, max, spms) =>
      val problems = analysis.infos.allInfos.values.flatMap(i => i.reportedProblems ++ i.unreportedProblems)
      val reporter = new LoggerReporter(max, s.log, Compiler.foldMappers(spms))
      problems foreach { p => reporter.display(p.position, p.message, p.severity) }
    }

  def sbtPluginExtra(m: ModuleID, sbtV: String, scalaV: String): ModuleID =
    m.extra(PomExtraDependencyAttributes.SbtVersionKey -> sbtV, PomExtraDependencyAttributes.ScalaVersionKey -> scalaV).copy(crossVersion = CrossVersion.Disabled)

  @deprecated("Use PluginDiscovery.writeDescriptor.", "0.13.2")
  def writePluginsDescriptor(plugins: Set[String], dir: File): Seq[File] =
    PluginDiscovery.writeDescriptor(plugins.toSeq, dir, PluginDiscovery.Paths.Plugins).toList

  def discoverSbtPluginNames: Initialize[Task[PluginDiscovery.DiscoveredNames]] = Def.task {
    if (sbtPlugin.value) PluginDiscovery.discoverSourceAll(compile.value) else PluginDiscovery.emptyDiscoveredNames
  }

  @deprecated("Use discoverSbtPluginNames.", "0.13.2")
  def discoverPlugins: Initialize[Task[Set[String]]] = (compile, sbtPlugin, streams) map { (analysis, isPlugin, s) => if (isPlugin) discoverSbtPlugins(analysis, s.log) else Set.empty }

  @deprecated("Use PluginDiscovery.sourceModuleNames[Plugin].", "0.13.2")
  def discoverSbtPlugins(analysis: Analysis, log: Logger): Set[String] =
    PluginDiscovery.sourceModuleNames(analysis, classOf[Plugin].getName).toSet

  def copyResourcesTask =
    (classDirectory, resources, resourceDirectories, streams) map { (target, resrcs, dirs, s) =>
      val cacheFile = s.cacheDirectory / "copy-resources"
      val mappings = (resrcs --- dirs) pair (rebase(dirs, target) | flat(target))
      s.log.debug("Copy resource mappings: " + mappings.mkString("\n\t", "\n\t", ""))
      Sync(cacheFile)(mappings)
      mappings
    }

  def runMainParser: (State, Seq[String]) => Parser[(String, Seq[String])] =
    {
      import DefaultParsers._
      (state, mainClasses) => Space ~> token(NotSpace examples mainClasses.toSet) ~ spaceDelimited("<arg>")
    }

  def testOnlyParser: (State, Seq[String]) => Parser[(Seq[String], Seq[String])] =
    { (state, tests) =>
      import DefaultParsers._
      val selectTests = distinctParser(tests.toSet, true)
      val options = (token(Space) ~> token("--") ~> spaceDelimited("<option>")) ?? Nil
      selectTests ~ options
    }

  private def distinctParser(exs: Set[String], raw: Boolean): Parser[Seq[String]] =
    {
      import DefaultParsers._
      val base = token(Space) ~> token(NotSpace - "--" examples exs)
      val recurse = base flatMap { ex =>
        val (matching, notMatching) = exs.partition(GlobFilter(ex).accept _)
        distinctParser(notMatching, raw) map { result => if (raw) ex +: result else matching.toSeq ++ result }
      }
      recurse ?? Nil
    }

  @deprecated("Use the new <key>.all(<ScopeFilter>) API", "0.13.0")
  def inDependencies[T](key: SettingKey[T], default: ProjectRef => T, includeRoot: Boolean = true, classpath: Boolean = true, aggregate: Boolean = false): Initialize[Seq[T]] =
    forDependencies[T, T](ref => (key in ref) ?? default(ref), includeRoot, classpath, aggregate)

  @deprecated("Use the new <key>.all(<ScopeFilter>) API", "0.13.0")
  def forDependencies[T, V](init: ProjectRef => Initialize[V], includeRoot: Boolean = true, classpath: Boolean = true, aggregate: Boolean = false): Initialize[Seq[V]] =
    Def.bind((loadedBuild, thisProjectRef).identity) {
      case (lb, base) =>
        transitiveDependencies(base, lb, includeRoot, classpath, aggregate) map init join;
    }

  def transitiveDependencies(base: ProjectRef, structure: LoadedBuild, includeRoot: Boolean, classpath: Boolean = true, aggregate: Boolean = false): Seq[ProjectRef] =
    {
      def tdeps(enabled: Boolean, f: ProjectRef => Seq[ProjectRef]): Seq[ProjectRef] =
        {
          val full = if (enabled) Dag.topologicalSort(base)(f) else Nil
          if (includeRoot) full else full dropRight 1
        }
      def fullCp = tdeps(classpath, getDependencies(structure, classpath = true, aggregate = false))
      def fullAgg = tdeps(aggregate, getDependencies(structure, classpath = false, aggregate = true))
      (classpath, aggregate) match {
        case (true, true)  => (fullCp ++ fullAgg).distinct
        case (true, false) => fullCp
        case _             => fullAgg
      }
    }
  def getDependencies(structure: LoadedBuild, classpath: Boolean = true, aggregate: Boolean = false): ProjectRef => Seq[ProjectRef] =
    ref => Project.getProject(ref, structure).toList flatMap { p =>
      (if (classpath) p.dependencies.map(_.project) else Nil) ++
        (if (aggregate) p.aggregate else Nil)
    }

  val CompletionsID = "completions"

  def noAggregation: Seq[Scoped] = Seq(run, runMain, console, consoleQuick, consoleProject)
  lazy val disableAggregation = Defaults.globalDefaults(noAggregation map disableAggregate)
  def disableAggregate(k: Scoped) = aggregate in k :== false

  lazy val runnerSettings: Seq[Setting[_]] = Seq(runnerTask)
  lazy val baseTasks: Seq[Setting[_]] = projectTasks ++ packageBase

  lazy val baseClasspaths: Seq[Setting[_]] = Classpaths.publishSettings ++ Classpaths.baseSettings
  lazy val configSettings: Seq[Setting[_]] = Classpaths.configSettings ++ configTasks ++ configPaths ++ packageConfig ++ Classpaths.compilerPluginConfig

  lazy val compileSettings: Seq[Setting[_]] = configSettings ++ (mainRunMainTask +: mainRunTask +: addBaseSources) ++ Classpaths.addUnmanagedLibrary
  lazy val testSettings: Seq[Setting[_]] = configSettings ++ testTasks

  lazy val itSettings: Seq[Setting[_]] = inConfig(IntegrationTest)(testSettings)
  lazy val defaultConfigs: Seq[Setting[_]] = inConfig(Compile)(compileSettings) ++ inConfig(Test)(testSettings) ++ inConfig(Runtime)(Classpaths.configSettings)

  // settings that are not specific to a configuration
  @deprecated("Settings now split into AutoPlugins.", "0.13.2")
  lazy val projectBaseSettings: Seq[Setting[_]] = projectCore ++ runnerSettings ++ paths ++ baseClasspaths ++ baseTasks ++ compileBase ++ disableAggregation

  // These are project level settings that MUST be on every project.
  lazy val coreDefaultSettings: Seq[Setting[_]] =
    projectCore ++ disableAggregation ++ Seq(
      // Missing but core settings
      baseDirectory := thisProject.value.base,
      target := baseDirectory.value / "target"
    )
  @deprecated("Default settings split into coreDefaultSettings and IvyModule/JvmModule plugins.", "0.13.2")
  lazy val defaultSettings: Seq[Setting[_]] = projectBaseSettings ++ defaultConfigs

}
object Classpaths {
  import Path._
  import Keys._
  import Scope.ThisScope
  import Defaults._
  import Attributed.{ blank, blankSeq }

  def concatDistinct[T](a: ScopedTaskable[Seq[T]], b: ScopedTaskable[Seq[T]]): Initialize[Task[Seq[T]]] = (a, b) map { (x, y) => (x ++ y).distinct }
  def concat[T](a: ScopedTaskable[Seq[T]], b: ScopedTaskable[Seq[T]]): Initialize[Task[Seq[T]]] = (a, b) map (_ ++ _)
  def concatSettings[T](a: SettingKey[Seq[T]], b: SettingKey[Seq[T]]): Initialize[Seq[T]] = (a, b)(_ ++ _)

  lazy val configSettings: Seq[Setting[_]] = classpaths ++ Seq(
    products <<= makeProducts,
    productDirectories := classDirectory.value :: Nil,
    classpathConfiguration := findClasspathConfig(internalConfigurationMap.value, configuration.value, classpathConfiguration.?.value, update.value)
  )
  private[this] def classpaths: Seq[Setting[_]] = Seq(
    externalDependencyClasspath <<= concat(unmanagedClasspath, managedClasspath),
    dependencyClasspath <<= concat(internalDependencyClasspath, externalDependencyClasspath),
    fullClasspath <<= concatDistinct(exportedProducts, dependencyClasspath),
    internalDependencyClasspath <<= internalDependencies,
    unmanagedClasspath <<= unmanagedDependencies,
    managedClasspath := managedJars(classpathConfiguration.value, classpathTypes.value, update.value),
    exportedProducts <<= exportProductsTask,
    unmanagedJars := findUnmanagedJars(configuration.value, unmanagedBase.value, includeFilter in unmanagedJars value, excludeFilter in unmanagedJars value)
  ).map(exportClasspath)

  private[this] def exportClasspath(s: Setting[Task[Classpath]]): Setting[Task[Classpath]] =
    s.mapInitialize(init => Def.task { exportClasspath(streams.value, init.value) })
  private[this] def exportClasspath(s: TaskStreams, cp: Classpath): Classpath =
    {
      val w = s.text(ExportStream)
      try w.println(Path.makeString(data(cp)))
      finally w.close() // workaround for #937
      cp
    }

  def defaultPackageKeys = Seq(packageBin, packageSrc, packageDoc)
  lazy val defaultPackages: Seq[TaskKey[File]] =
    for (task <- defaultPackageKeys; conf <- Seq(Compile, Test)) yield (task in conf)
  lazy val defaultArtifactTasks: Seq[TaskKey[File]] = makePom +: defaultPackages

  def findClasspathConfig(map: Configuration => Configuration, thisConfig: Configuration, delegated: Option[Configuration], report: UpdateReport): Configuration =
    {
      val defined = report.allConfigurations.toSet
      val search = map(thisConfig) +: (delegated.toList ++ Seq(Compile, Configurations.Default))
      def notFound = sys.error("Configuration to use for managed classpath must be explicitly defined when default configurations are not present.")
      search find { defined contains _.name } getOrElse notFound
    }

  def packaged(pkgTasks: Seq[TaskKey[File]]): Initialize[Task[Map[Artifact, File]]] =
    enabledOnly(packagedArtifact.task, pkgTasks) apply (_.join.map(_.toMap))
  def artifactDefs(pkgTasks: Seq[TaskKey[File]]): Initialize[Seq[Artifact]] =
    enabledOnly(artifact, pkgTasks)

  def enabledOnly[T](key: SettingKey[T], pkgTasks: Seq[TaskKey[File]]): Initialize[Seq[T]] =
    (forallIn(key, pkgTasks) zipWith forallIn(publishArtifact, pkgTasks))(_ zip _ collect { case (a, true) => a })
  def forallIn[T](key: SettingKey[T], pkgTasks: Seq[TaskKey[_]]): Initialize[Seq[T]] =
    pkgTasks.map(pkg => key in pkg.scope in pkg).join

  private[this] def publishGlobalDefaults = Defaults.globalDefaults(Seq(
    publishMavenStyle :== true,
    publishArtifact :== true,
    publishArtifact in Test :== false
  ))

  val jvmPublishSettings: Seq[Setting[_]] = Seq(
    artifacts <<= artifactDefs(defaultArtifactTasks),
    packagedArtifacts <<= packaged(defaultArtifactTasks)
  )

  val ivyPublishSettings: Seq[Setting[_]] = publishGlobalDefaults ++ Seq(
    artifacts :== Nil,
    packagedArtifacts :== Map.empty,
    crossTarget := target.value,
    makePom := { val config = makePomConfiguration.value; IvyActions.makePom(ivyModule.value, config, streams.value.log); config.file },
    packagedArtifact in makePom := ((artifact in makePom).value -> makePom.value),
    deliver <<= deliverTask(deliverConfiguration),
    deliverLocal <<= deliverTask(deliverLocalConfiguration),
    publish <<= publishTask(publishConfiguration, deliver),
    publishLocal <<= publishTask(publishLocalConfiguration, deliverLocal),
    publishM2 <<= publishTask(publishM2Configuration, deliverLocal)
  )
  @deprecated("This has been split into jvmPublishSettings and ivyPublishSettings.", "0.13.2")
  val publishSettings: Seq[Setting[_]] = ivyPublishSettings ++ jvmPublishSettings

  private[this] def baseGlobalDefaults = Defaults.globalDefaults(Seq(
    conflictWarning :== ConflictWarning.default("global"),
    homepage :== None,
    startYear :== None,
    licenses :== Nil,
    developers :== Nil,
    scmInfo :== None,
    offline :== false,
    defaultConfiguration :== Some(Configurations.Compile),
    dependencyOverrides :== Set.empty,
    libraryDependencies :== Nil,
    excludeDependencies :== Nil,
    ivyLoggingLevel :== UpdateLogging.Default,
    ivyXML :== NodeSeq.Empty,
    ivyValidate :== false,
    moduleConfigurations :== Nil,
    publishTo :== None,
    resolvers :== Nil,
    retrievePattern :== Resolver.defaultRetrievePattern,
    transitiveClassifiers :== Seq(SourceClassifier, DocClassifier),
    sbtDependency := {
      val app = appConfiguration.value
      val id = app.provider.id
      val scalaVersion = app.provider.scalaProvider.version
      val binVersion = binaryScalaVersion(scalaVersion)
      val cross = if (id.crossVersioned) CrossVersion.binary else CrossVersion.Disabled
      val base = ModuleID(id.groupID, id.name, sbtVersion.value, crossVersion = cross)
      CrossVersion(scalaVersion, binVersion)(base).copy(crossVersion = CrossVersion.Disabled)
    }
  ))

  val ivyBaseSettings: Seq[Setting[_]] = baseGlobalDefaults ++ sbtClassifiersTasks ++ Seq(
    conflictWarning := conflictWarning.value.copy(label = Reference.display(thisProjectRef.value)),
    unmanagedBase := baseDirectory.value / "lib",
    normalizedName := Project.normalizeModuleID(name.value),
    isSnapshot <<= isSnapshot or version(_ endsWith "-SNAPSHOT"),
    description <<= description or name,
    organization <<= organization or normalizedName,
    organizationName <<= organizationName or organization,
    organizationHomepage <<= organizationHomepage or homepage,
    projectInfo <<= (name, description, homepage, startYear, licenses, organizationName, organizationHomepage, scmInfo, developers) apply ModuleInfo,
    overrideBuildResolvers <<= appConfiguration(isOverrideRepositories),
    externalResolvers <<= (externalResolvers.task.?, resolvers, appResolvers) {
      case (Some(delegated), Seq(), _) => delegated
      case (_, rs, Some(ars))          => task { ars ++ rs } // TODO - Do we need to filter out duplicates?
      case (_, rs, _)                  => task { Resolver.withDefaultResolvers(rs) }
    },
    appResolvers <<= appConfiguration apply appRepositories,
    bootResolvers <<= appConfiguration map bootRepositories,
    fullResolvers <<= (projectResolver, externalResolvers, sbtPlugin, sbtResolver, bootResolvers, overrideBuildResolvers) map { (proj, rs, isPlugin, sbtr, boot, overrideFlag) =>
      boot match {
        case Some(repos) if overrideFlag => proj +: repos
        case _ =>
          val base = if (isPlugin) sbtr +: sbtPluginReleases +: rs else rs
          proj +: base
      }
    },
    moduleName <<= normalizedName,
    ivyPaths := new IvyPaths(baseDirectory.value, bootIvyHome(appConfiguration.value)),
    dependencyCacheDirectory := {
      val st = state.value
      BuildPaths.getDependencyDirectory(st, BuildPaths.getGlobalBase(st))
    },
    otherResolvers := Resolver.publishMavenLocal :: publishTo.value.toList,
    projectResolver <<= projectResolverTask,
    projectDependencies <<= projectDependenciesTask,
    // TODO - Is this the appropriate split?  Ivy defines this simply as
    //        just project + library, while the JVM plugin will define it as
    //        having the additional sbtPlugin + autoScala magikz.
    allDependencies := {
      projectDependencies.value ++ libraryDependencies.value
    },
    ivyScala <<= ivyScala or (scalaHome, scalaVersion in update, scalaBinaryVersion in update, scalaOrganization, sbtPlugin) { (sh, fv, bv, so, plugin) =>
      Some(new IvyScala(fv, bv, Nil, filterImplicit = false, checkExplicit = true, overrideScalaVersion = plugin, scalaOrganization = so))
    },
    artifactPath in makePom <<= artifactPathSetting(artifact in makePom),
    publishArtifact in makePom := publishMavenStyle.value && publishArtifact.value,
    artifact in makePom := Artifact.pom(moduleName.value),
    projectID <<= defaultProjectID,
    projectID <<= pluginProjectID,
    projectDescriptors <<= depMap,
    updateConfiguration := new UpdateConfiguration(retrieveConfiguration.value, false, ivyLoggingLevel.value),
    updateOptions := (updateOptions in Global).value,
    retrieveConfiguration := { if (retrieveManaged.value) Some(new RetrieveConfiguration(managedDirectory.value, retrievePattern.value, retrieveManagedSync.value, configurationsToRetrieve.value)) else None },
    ivyConfiguration <<= mkIvyConfiguration,
    ivyConfigurations := {
      val confs = thisProject.value.configurations
      (confs ++ confs.map(internalConfigurationMap.value) ++ (if (autoCompilerPlugins.value) CompilerPlugin :: Nil else Nil)).distinct
    },
    ivyConfigurations ++= Configurations.auxiliary,
    ivyConfigurations ++= { if (managedScalaInstance.value && scalaHome.value.isEmpty) Configurations.ScalaTool :: Nil else Nil },
    moduleSettings <<= moduleSettings0,
    makePomConfiguration := new MakePomConfiguration(artifactPath in makePom value, projectInfo.value, None, pomExtra.value, pomPostProcess.value, pomIncludeRepository.value, pomAllRepositories.value),
    deliverLocalConfiguration := deliverConfig(crossTarget.value, status = if (isSnapshot.value) "integration" else "release", logging = ivyLoggingLevel.value),
    deliverConfiguration <<= deliverLocalConfiguration,
    publishConfiguration := publishConfig(packagedArtifacts.in(publish).value, if (publishMavenStyle.value) None else Some(deliver.value), resolverName = getPublishTo(publishTo.value).name, checksums = checksums.in(publish).value, logging = ivyLoggingLevel.value, overwrite = isSnapshot.value),
    publishLocalConfiguration := publishConfig(packagedArtifacts.in(publishLocal).value, Some(deliverLocal.value), checksums.in(publishLocal).value, logging = ivyLoggingLevel.value, overwrite = isSnapshot.value),
    publishM2Configuration := publishConfig(packagedArtifacts.in(publishM2).value, None, resolverName = Resolver.publishMavenLocal.name, checksums = checksums.in(publishM2).value, logging = ivyLoggingLevel.value, overwrite = isSnapshot.value),
    ivySbt <<= ivySbt0,
    ivyModule := { val is = ivySbt.value; new is.Module(moduleSettings.value) },
    transitiveUpdate <<= transitiveUpdateTask,
    updateCacheName := "update_cache" + (if (crossPaths.value) s"_${scalaBinaryVersion.value}" else ""),
    evictionWarningOptions in update := EvictionWarningOptions.default,
    dependencyPositions <<= dependencyPositionsTask,
    unresolvedWarningConfiguration in update := UnresolvedWarningConfiguration(dependencyPositions.value),
    update <<= updateTask tag (Tags.Update, Tags.Network),
    update := {
      val report = update.value
      val log = streams.value.log
      ConflictWarning(conflictWarning.value, report, log)
      report
    },
    evictionWarningOptions in evicted := EvictionWarningOptions.full,
    evicted := {
      import ShowLines._
      val report = (updateTask tag (Tags.Update, Tags.Network)).value
      val log = streams.value.log
      val ew = EvictionWarning(ivyModule.value, (evictionWarningOptions in evicted).value, report, log)
      ew.lines foreach { log.warn(_) }
      ew.infoAllTheThings foreach { log.info(_) }
      ew
    },
    classifiersModule in updateClassifiers := {
      import language.implicitConversions
      implicit val key = (m: ModuleID) => (m.organization, m.name, m.revision)
      val projectDeps = projectDependencies.value.iterator.map(key).toSet
      val externalModules = update.value.allModules.filterNot(m => projectDeps contains key(m))
      GetClassifiersModule(projectID.value, externalModules, ivyConfigurations.in(updateClassifiers).value, transitiveClassifiers.in(updateClassifiers).value)
    },
    updateClassifiers <<= Def.task {
      val s = streams.value
      val is = ivySbt.value
      val mod = (classifiersModule in updateClassifiers).value
      val c = updateConfiguration.value
      val app = appConfiguration.value
      val out = is.withIvy(s.log)(_.getSettings.getDefaultIvyUserDir)
      val uwConfig = (unresolvedWarningConfiguration in update).value
      val depDir = dependencyCacheDirectory.value
      withExcludes(out, mod.classifiers, lock(app)) { excludes =>
        val uwConfig = (unresolvedWarningConfiguration in update).value
        val logicalClock = LogicalClock(state.value.hashCode)
        val depDir = dependencyCacheDirectory.value
        IvyActions.updateClassifiers(is, GetClassifiersConfiguration(mod, excludes, c, ivyScala.value), uwConfig, LogicalClock(state.value.hashCode), Some(depDir), s.log)
      }
    } tag (Tags.Update, Tags.Network)
  )

  val jvmBaseSettings: Seq[Setting[_]] = Seq(
    libraryDependencies ++= autoLibraryDependency(autoScalaLibrary.value && scalaHome.value.isEmpty && managedScalaInstance.value, sbtPlugin.value, scalaOrganization.value, scalaVersion.value),
    // Override the default to handle mixing in the sbtPlugin + scala dependencies.
    allDependencies := {
      val base = projectDependencies.value ++ libraryDependencies.value
      val pluginAdjust = if (sbtPlugin.value) sbtDependency.value.copy(configurations = Some(Provided.name)) +: base else base
      if (scalaHome.value.isDefined || ivyScala.value.isEmpty || !managedScalaInstance.value)
        pluginAdjust
      else
        ScalaArtifacts.toolDependencies(scalaOrganization.value, scalaVersion.value) ++ pluginAdjust
    }
  )
  @deprecated("Split into ivyBaseSettings and jvmBaseSettings.", "0.13.2")
  val baseSettings: Seq[Setting[_]] = ivyBaseSettings ++ jvmBaseSettings

  def warnResolversConflict(ress: Seq[Resolver], log: Logger): Unit = {
    val resset = ress.toSet
    for ((name, r) <- resset groupBy (_.name) if r.size > 1) {
      log.warn("Multiple resolvers having different access mechanism configured with same name '" + name + "'. To avoid conflict, Remove duplicate project resolvers (`resolvers`) or rename publishing resolver (`publishTo`).")
    }
  }

  private[sbt] def defaultProjectID: Initialize[ModuleID] = Def.setting {
    val base = ModuleID(organization.value, moduleName.value, version.value).cross(crossVersion in projectID value).artifacts(artifacts.value: _*)
    apiURL.value match {
      case Some(u) if autoAPIMappings.value => base.extra(SbtPomExtraProperties.POM_API_KEY -> u.toExternalForm)
      case _                                => base
    }
  }

  def pluginProjectID: Initialize[ModuleID] = (sbtBinaryVersion in update, scalaBinaryVersion in update, projectID, sbtPlugin) {
    (sbtBV, scalaBV, pid, isPlugin) =>
      if (isPlugin) sbtPluginExtra(pid, sbtBV, scalaBV) else pid
  }
  def ivySbt0: Initialize[Task[IvySbt]] =
    (ivyConfiguration, credentials, streams) map { (conf, creds, s) =>
      Credentials.register(creds, s.log)
      new IvySbt(conf)
    }
  def moduleSettings0: Initialize[Task[ModuleSettings]] = Def.task {
    new InlineConfigurationWithExcludes(projectID.value, projectInfo.value, allDependencies.value, dependencyOverrides.value, excludeDependencies.value,
      ivyXML.value, ivyConfigurations.value, defaultConfiguration.value, ivyScala.value, ivyValidate.value, conflictManager.value)
  }

  private[this] def sbtClassifiersGlobalDefaults = Defaults.globalDefaults(Seq(
    transitiveClassifiers in updateSbtClassifiers ~= (_.filter(_ != DocClassifier))
  ))
  def sbtClassifiersTasks = sbtClassifiersGlobalDefaults ++ inTask(updateSbtClassifiers)(Seq(
    externalResolvers := {
      val explicit = buildStructure.value.units(thisProjectRef.value.build).unit.plugins.pluginData.resolvers
      explicit orElse bootRepositories(appConfiguration.value) getOrElse externalResolvers.value
    },
    ivyConfiguration := new InlineIvyConfiguration(ivyPaths.value, externalResolvers.value, Nil, Nil, offline.value, Option(lock(appConfiguration.value)),
      checksums.value, Some(target.value / "resolution-cache"), UpdateOptions(), streams.value.log),
    ivySbt <<= ivySbt0,
    classifiersModule <<= (projectID, sbtDependency, transitiveClassifiers, loadedBuild, thisProjectRef) map { (pid, sbtDep, classifiers, lb, ref) =>
      val pluginClasspath = lb.units(ref.build).unit.plugins.fullClasspath
      val pluginJars = pluginClasspath.filter(_.data.isFile) // exclude directories: an approximation to whether they've been published
      val pluginIDs: Seq[ModuleID] = pluginJars.flatMap(_ get moduleID.key)
      GetClassifiersModule(pid, sbtDep +: pluginIDs, Configurations.Default :: Nil, classifiers)
    },
    updateSbtClassifiers in TaskGlobal <<= Def.task {
      val s = streams.value
      val is = ivySbt.value
      val mod = classifiersModule.value
      val c = updateConfiguration.value
      val app = appConfiguration.value
      val out = is.withIvy(s.log)(_.getSettings.getDefaultIvyUserDir)
      val uwConfig = (unresolvedWarningConfiguration in update).value
      val depDir = dependencyCacheDirectory.value
      withExcludes(out, mod.classifiers, lock(app)) { excludes =>
        val noExplicitCheck = ivyScala.value.map(_.copy(checkExplicit = false))
        IvyActions.transitiveScratch(is, "sbt", GetClassifiersConfiguration(mod, excludes, c, noExplicitCheck), uwConfig, LogicalClock(state.value.hashCode), Some(depDir), s.log)
      }
    } tag (Tags.Update, Tags.Network)
  ))

  def deliverTask(config: TaskKey[DeliverConfiguration]): Initialize[Task[File]] =
    (ivyModule, config, update, streams) map { (module, config, _, s) => IvyActions.deliver(module, config, s.log) }
  def publishTask(config: TaskKey[PublishConfiguration], deliverKey: TaskKey[_]): Initialize[Task[Unit]] =
    (ivyModule, config, streams) map { (module, config, s) =>
      IvyActions.publish(module, config, s.log)
    } tag (Tags.Publish, Tags.Network)

  import Cache._
  import CacheIvy.{
    classpathFormat, /*publishIC,*/ updateIC,
    updateReportFormat,
    excludeMap,
    moduleIDSeqIC,
    modulePositionMapFormat
  }

  def withExcludes(out: File, classifiers: Seq[String], lock: xsbti.GlobalLock)(f: Map[ModuleID, Set[String]] => UpdateReport): UpdateReport =
    {
      val exclName = "exclude_classifiers"
      val file = out / exclName
      lock(out / (exclName + ".lock"), new Callable[UpdateReport] {
        def call = {
          val excludes = CacheIO.fromFile[Map[ModuleID, Set[String]]](excludeMap, Map.empty[ModuleID, Set[String]])(file)
          val report = f(excludes)
          val allExcludes = excludes ++ IvyActions.extractExcludes(report)
          CacheIO.toFile(excludeMap)(allExcludes)(file)
          IvyActions.addExcluded(report, classifiers, allExcludes)
        }
      })
    }

  def updateTask: Initialize[Task[UpdateReport]] = Def.task {
    val depsUpdated = transitiveUpdate.value.exists(!_.stats.cached)
    val isRoot = executionRoots.value contains resolvedScoped.value
    val forceUpdate = forceUpdatePeriod.value
    val s = streams.value
    val fullUpdateOutput = s.cacheDirectory / "out"
    val forceUpdateByTime = forceUpdate match {
      case None => false
      case Some(period) =>
        val elapsedDuration = new FiniteDuration(System.currentTimeMillis() - fullUpdateOutput.lastModified(), TimeUnit.MILLISECONDS)
        fullUpdateOutput.exists() && elapsedDuration > period
    }
    val scalaProvider = appConfiguration.value.provider.scalaProvider

    // Only substitute unmanaged jars for managed jars when the major.minor parts of the versions the same for:
    //   the resolved Scala version and the scalaHome version: compatible (weakly- no qualifier checked)
    //   the resolved Scala version and the declared scalaVersion: assume the user intended scalaHome to override anything with scalaVersion
    def subUnmanaged(subVersion: String, jars: Seq[File]) = (sv: String) =>
      (partialVersion(sv), partialVersion(subVersion), partialVersion(scalaVersion.value)) match {
        case (Some(res), Some(sh), _) if res == sh     => jars
        case (Some(res), _, Some(decl)) if res == decl => jars
        case _                                         => Nil
      }
    val subScalaJars: String => Seq[File] = Defaults.unmanagedScalaInstanceOnly.value match {
      case Some(si) => subUnmanaged(si.version, si.allJars)
      case None     => sv => if (scalaProvider.version == sv) scalaProvider.jars else Nil
    }
    val transform: UpdateReport => UpdateReport = r => substituteScalaFiles(scalaOrganization.value, r)(subScalaJars)
    val uwConfig = (unresolvedWarningConfiguration in update).value
    val show = Reference.display(thisProjectRef.value)
    val st = state.value
    val logicalClock = LogicalClock(st.hashCode)
    val depDir = dependencyCacheDirectory.value
    val uc0 = updateConfiguration.value
    // Normally, log would capture log messages at all levels.
    // Ivy logs are treated specially using sbt.UpdateConfiguration.logging.
    // This code bumps up the sbt.UpdateConfiguration.logging to Full when logLevel is Debug.
    import UpdateLogging.{ Full, DownloadOnly, Default }
    val uc = (logLevel in update).?.value orElse st.get(logLevel.key) match {
      case Some(Level.Debug) if uc0.logging == Default => uc0.copy(logging = Full)
      case Some(x) if uc0.logging == Default => uc0.copy(logging = DownloadOnly)
      case _ => uc0
    }
    val ewo =
      if (executionRoots.value exists { _.key == evicted.key }) EvictionWarningOptions.empty
      else (evictionWarningOptions in update).value
    cachedUpdate(s.cacheDirectory / updateCacheName.value, show, ivyModule.value, uc, transform,
      skip = (skip in update).value, force = isRoot || forceUpdateByTime, depsUpdated = depsUpdated,
      uwConfig = uwConfig, logicalClock = logicalClock, depDir = Some(depDir),
      ewo = ewo, log = s.log)
  }
  @deprecated("Use cachedUpdate with the variant that takes unresolvedHandler instead.", "0.13.6")
  def cachedUpdate(cacheFile: File, label: String, module: IvySbt#Module, config: UpdateConfiguration,
    transform: UpdateReport => UpdateReport, skip: Boolean, force: Boolean, depsUpdated: Boolean, log: Logger): UpdateReport =
    cachedUpdate(cacheFile, label, module, config, transform, skip, force, depsUpdated,
      UnresolvedWarningConfiguration(), LogicalClock.unknown, None, EvictionWarningOptions.empty, log)
  private[sbt] def cachedUpdate(cacheFile: File, label: String, module: IvySbt#Module, config: UpdateConfiguration,
    transform: UpdateReport => UpdateReport, skip: Boolean, force: Boolean, depsUpdated: Boolean,
    uwConfig: UnresolvedWarningConfiguration, logicalClock: LogicalClock, depDir: Option[File],
    ewo: EvictionWarningOptions, log: Logger): UpdateReport =
    {
      implicit val updateCache = updateIC
      type In = IvyConfiguration :+: ModuleSettings :+: UpdateConfiguration :+: HNil
      def work = (_: In) match {
        case conf :+: settings :+: config :+: HNil =>
          import ShowLines._
          log.info("Updating " + label + "...")
          val r = IvyActions.updateEither(module, config, uwConfig, logicalClock, depDir, log) match {
            case Right(ur) => ur
            case Left(uw) =>
              uw.lines foreach { log.warn(_) }
              throw uw.resolveException
          }
          log.info("Done updating.")
          val result = transform(r)
          val ew = EvictionWarning(module, ewo, result, log)
          ew.lines foreach { log.warn(_) }
          ew.infoAllTheThings foreach { log.info(_) }
          result
      }
      def uptodate(inChanged: Boolean, out: UpdateReport): Boolean =
        !force &&
          !depsUpdated &&
          !inChanged &&
          out.allFiles.forall(f => fileUptodate(f, out.stamps)) &&
          fileUptodate(out.cachedDescriptor, out.stamps)

      val outCacheFile = cacheFile / "output"
      def skipWork: In => UpdateReport =
        Tracked.lastOutput[In, UpdateReport](outCacheFile) {
          case (_, Some(out)) => out
          case _              => sys.error("Skipping update requested, but update has not previously run successfully.")
        }
      def doWork: In => UpdateReport =
        Tracked.inputChanged(cacheFile / "inputs") { (inChanged: Boolean, in: In) =>
          val outCache = Tracked.lastOutput[In, UpdateReport](outCacheFile) {
            case (_, Some(out)) if uptodate(inChanged, out) => out
            case _ => work(in)
          }
          try {
            outCache(in)
          } catch {
            case e: NullPointerException =>
              val r = work(in)
              log.warn("Update task has failed to cache the report due to null.")
              log.warn("Report the following output to sbt:")
              r.toString.lines foreach { log.warn(_) }
              log.trace(e)
              r
            case e: OutOfMemoryError =>
              val r = work(in)
              log.warn("Update task has failed to cache the report due to OutOfMemoryError.")
              log.trace(e)
              r
          }
        }
      val f = if (skip && !force) skipWork else doWork
      f(module.owner.configuration :+: module.moduleSettings :+: config :+: HNil)
    }
  private[this] def fileUptodate(file: File, stamps: Map[File, Long]): Boolean =
    stamps.get(file).forall(_ == file.lastModified)
  private[sbt] def dependencyPositionsTask: Initialize[Task[Map[ModuleID, SourcePosition]]] = Def.task {
    val projRef = thisProjectRef.value
    val st = state.value
    val s = streams.value
    val cacheFile = s.cacheDirectory / updateCacheName.value
    implicit val depSourcePosCache = moduleIDSeqIC
    implicit val outFormat = modulePositionMapFormat
    def modulePositions: Map[ModuleID, SourcePosition] =
      try {
        val extracted = (Project extract st)
        val sk = (libraryDependencies in (GlobalScope in projRef)).scopedKey
        val empty = extracted.structure.data set (sk.scope, sk.key, Nil)
        val settings = extracted.structure.settings filter { s: Setting[_] =>
          (s.key.key == libraryDependencies.key) &&
            (s.key.scope.project == Select(projRef))
        }
        Map(settings flatMap {
          case s: Setting[Seq[ModuleID]] @unchecked =>
            s.init.evaluate(empty) map { _ -> s.pos }
        }: _*)
      } catch {
        case _: Throwable => Map()
      }

    val outCacheFile = cacheFile / "output_dsp"
    val f = Tracked.inputChanged(cacheFile / "input_dsp") { (inChanged: Boolean, in: Seq[ModuleID]) =>
      val outCache = Tracked.lastOutput[Seq[ModuleID], Map[ModuleID, SourcePosition]](outCacheFile) {
        case (_, Some(out)) if !inChanged => out
        case _                            => modulePositions
      }
      outCache(in)
    }
    f(libraryDependencies.value)
  }

  /*
	// can't cache deliver/publish easily since files involved are hidden behind patterns.  publish will be difficult to verify target-side anyway
	def cachedPublish(cacheFile: File)(g: (IvySbt#Module, PublishConfiguration) => Unit, module: IvySbt#Module, config: PublishConfiguration) => Unit =
	{ case module :+: config :+: HNil =>
	/*	implicit val publishCache = publishIC
		val f = cached(cacheFile) { (conf: IvyConfiguration, settings: ModuleSettings, config: PublishConfiguration) =>*/
		    g(module, config)
		/*}
		f(module.owner.configuration :+: module.moduleSettings :+: config :+: HNil)*/
	}*/

  def defaultRepositoryFilter = (repo: MavenRepository) => !repo.root.startsWith("file:")
  def getPublishTo(repo: Option[Resolver]): Resolver = repo getOrElse sys.error("Repository for publishing is not specified.")

  def deliverConfig(outputDirectory: File, status: String = "release", logging: UpdateLogging.Value = UpdateLogging.DownloadOnly) =
    new DeliverConfiguration(deliverPattern(outputDirectory), status, None, logging)
  @deprecated("Previous semantics allowed overwriting cached files, which was unsafe. Please specify overwrite parameter.", "0.13.2")
  def publishConfig(artifacts: Map[Artifact, File], ivyFile: Option[File], checksums: Seq[String], resolverName: String, logging: UpdateLogging.Value): PublishConfiguration =
    publishConfig(artifacts, ivyFile, checksums, resolverName, logging, overwrite = true)
  def publishConfig(artifacts: Map[Artifact, File], ivyFile: Option[File], checksums: Seq[String], resolverName: String = "local", logging: UpdateLogging.Value = UpdateLogging.DownloadOnly, overwrite: Boolean = false) =
    new PublishConfiguration(ivyFile, resolverName, artifacts, checksums, logging, overwrite)

  def deliverPattern(outputPath: File): String = (outputPath / "[artifact]-[revision](-[classifier]).[ext]").absolutePath

  def projectDependenciesTask: Initialize[Task[Seq[ModuleID]]] =
    (thisProjectRef, settingsData, buildDependencies) map { (ref, data, deps) =>
      deps.classpath(ref) flatMap { dep =>
        (projectID in dep.project) get data map {
          _.copy(configurations = dep.configuration, explicitArtifacts = Nil)
        }
      }
    }

  def depMap: Initialize[Task[Map[ModuleRevisionId, ModuleDescriptor]]] =
    (thisProjectRef, settingsData, buildDependencies, streams) flatMap { (root, data, deps, s) =>
      depMap(deps classpathTransitiveRefs root, data, s.log)
    }

  def depMap(projects: Seq[ProjectRef], data: Settings[Scope], log: Logger): Task[Map[ModuleRevisionId, ModuleDescriptor]] =
    projects.flatMap(ivyModule in _ get data).join.map { mod =>
      mod map { _.dependencyMapping(log) } toMap;
    }

  def projectResolverTask: Initialize[Task[Resolver]] =
    projectDescriptors map { m =>
      new RawRepository(new ProjectResolver(ProjectResolver.InterProject, m))
    }

  def analyzed[T](data: T, analysis: Analysis) = Attributed.blank(data).put(Keys.analysis, analysis)
  def makeProducts: Initialize[Task[Seq[File]]] = Def.task {
    val x1 = compile.value
    val x2 = copyResources.value
    classDirectory.value :: Nil
  }
  def exportProductsTask: Initialize[Task[Classpath]] = Def.task {
    val art = (artifact in packageBin).value
    val module = projectID.value
    val config = configuration.value
    for (f <- productsImplTask.value) yield APIMappings.store(analyzed(f, compile.value), apiURL.value).put(artifact.key, art).put(moduleID.key, module).put(configuration.key, config)
  }

  private[this] def productsImplTask: Initialize[Task[Seq[File]]] =
    (products.task, packageBin.task, exportJars) flatMap { (psTask, pkgTask, useJars) =>
      if (useJars) Seq(pkgTask).join else psTask
    }

  def constructBuildDependencies: Initialize[BuildDependencies] = loadedBuild(lb => BuildUtil.dependencies(lb.units))

  def internalDependencies: Initialize[Task[Classpath]] =
    (thisProjectRef, classpathConfiguration, configuration, settingsData, buildDependencies) flatMap internalDependencies0
  def unmanagedDependencies: Initialize[Task[Classpath]] =
    (thisProjectRef, configuration, settingsData, buildDependencies) flatMap unmanagedDependencies0
  def mkIvyConfiguration: Initialize[Task[IvyConfiguration]] =
    (fullResolvers, ivyPaths, otherResolvers, moduleConfigurations, offline, checksums in update, appConfiguration,
      target, updateOptions, streams) map { (rs, paths, other, moduleConfs, off, check, app, t, uo, s) =>
        warnResolversConflict(rs ++: other, s.log)
        val resCacheDir = t / "resolution-cache"

        new InlineIvyConfiguration(paths, rs, other, moduleConfs, off, Option(lock(app)), check, Some(resCacheDir), uo, s.log)
      }

  import java.util.LinkedHashSet
  import collection.JavaConversions.asScalaSet
  def interSort(projectRef: ProjectRef, conf: Configuration, data: Settings[Scope], deps: BuildDependencies): Seq[(ProjectRef, String)] =
    {
      val visited = asScalaSet(new LinkedHashSet[(ProjectRef, String)])
      def visit(p: ProjectRef, c: Configuration): Unit = {
        val applicableConfigs = allConfigs(c)
        for (ac <- applicableConfigs) // add all configurations in this project
          visited add (p -> ac.name)
        val masterConfs = names(getConfigurations(projectRef, data))

        for (ResolvedClasspathDependency(dep, confMapping) <- deps.classpath(p)) {
          val configurations = getConfigurations(dep, data)
          val mapping = mapped(confMapping, masterConfs, names(configurations), "compile", "*->compile")
          // map master configuration 'c' and all extended configurations to the appropriate dependency configuration
          for (ac <- applicableConfigs; depConfName <- mapping(ac.name)) {
            for (depConf <- confOpt(configurations, depConfName))
              if (!visited((dep, depConfName)))
                visit(dep, depConf)
          }
        }
      }
      visit(projectRef, conf)
      visited.toSeq
    }
  def unmanagedDependencies0(projectRef: ProjectRef, conf: Configuration, data: Settings[Scope], deps: BuildDependencies): Task[Classpath] =
    interDependencies(projectRef, deps, conf, conf, data, true, unmanagedLibs)
  def internalDependencies0(projectRef: ProjectRef, conf: Configuration, self: Configuration, data: Settings[Scope], deps: BuildDependencies): Task[Classpath] =
    interDependencies(projectRef, deps, conf, self, data, false, productsTask)
  def interDependencies(projectRef: ProjectRef, deps: BuildDependencies, conf: Configuration, self: Configuration, data: Settings[Scope], includeSelf: Boolean,
    f: (ProjectRef, String, Settings[Scope]) => Task[Classpath]): Task[Classpath] =
    {
      val visited = interSort(projectRef, conf, data, deps)
      val tasks = asScalaSet(new LinkedHashSet[Task[Classpath]])
      for ((dep, c) <- visited)
        if (includeSelf || (dep != projectRef) || (conf.name != c && self.name != c))
          tasks += f(dep, c, data)

      (tasks.toSeq.join).map(_.flatten.distinct)
    }

  def mapped(confString: Option[String], masterConfs: Seq[String], depConfs: Seq[String], default: String, defaultMapping: String): String => Seq[String] =
    {
      lazy val defaultMap = parseMapping(defaultMapping, masterConfs, depConfs, _ :: Nil)
      parseMapping(confString getOrElse default, masterConfs, depConfs, defaultMap)
    }
  def parseMapping(confString: String, masterConfs: Seq[String], depConfs: Seq[String], default: String => Seq[String]): String => Seq[String] =
    union(confString.split(";") map parseSingleMapping(masterConfs, depConfs, default))
  def parseSingleMapping(masterConfs: Seq[String], depConfs: Seq[String], default: String => Seq[String])(confString: String): String => Seq[String] =
    {
      val ms: Seq[(String, Seq[String])] =
        trim(confString.split("->", 2)) match {
          case x :: Nil => for (a <- parseList(x, masterConfs)) yield (a, default(a))
          case x :: y :: Nil =>
            val target = parseList(y, depConfs); for (a <- parseList(x, masterConfs)) yield (a, target)
          case _ => sys.error("Invalid configuration '" + confString + "'") // shouldn't get here
        }
      val m = ms.toMap
      s => m.getOrElse(s, Nil)
    }

  def union[A, B](maps: Seq[A => Seq[B]]): A => Seq[B] =
    a => (Seq[B]() /: maps) { _ ++ _(a) } distinct;

  def parseList(s: String, allConfs: Seq[String]): Seq[String] = (trim(s split ",") flatMap replaceWildcard(allConfs)).distinct
  def replaceWildcard(allConfs: Seq[String])(conf: String): Seq[String] =
    if (conf == "") Nil else if (conf == "*") allConfs else conf :: Nil

  private def trim(a: Array[String]): List[String] = a.toList.map(_.trim)
  def missingConfiguration(in: String, conf: String) =
    sys.error("Configuration '" + conf + "' not defined in '" + in + "'")
  def allConfigs(conf: Configuration): Seq[Configuration] =
    Dag.topologicalSort(conf)(_.extendsConfigs)

  def getConfigurations(p: ResolvedReference, data: Settings[Scope]): Seq[Configuration] =
    ivyConfigurations in p get data getOrElse Nil
  def confOpt(configurations: Seq[Configuration], conf: String): Option[Configuration] =
    configurations.find(_.name == conf)
  def productsTask(dep: ResolvedReference, conf: String, data: Settings[Scope]): Task[Classpath] =
    getClasspath(exportedProducts, dep, conf, data)
  def unmanagedLibs(dep: ResolvedReference, conf: String, data: Settings[Scope]): Task[Classpath] =
    getClasspath(unmanagedJars, dep, conf, data)
  def getClasspath(key: TaskKey[Classpath], dep: ResolvedReference, conf: String, data: Settings[Scope]): Task[Classpath] =
    (key in (dep, ConfigKey(conf))) get data getOrElse constant(Nil)
  def defaultConfigurationTask(p: ResolvedReference, data: Settings[Scope]): Configuration =
    flatten(defaultConfiguration in p get data) getOrElse Configurations.Default
  def flatten[T](o: Option[Option[T]]): Option[T] = o flatMap idFun

  lazy val typesafeReleases = Resolver.typesafeIvyRepo("releases")
  lazy val typesafeSnapshots = Resolver.typesafeIvyRepo("snapshots")

  @deprecated("Use `typesafeReleases` instead", "0.12.0")
  lazy val typesafeResolver = typesafeReleases
  @deprecated("Use `Resolver.typesafeIvyRepo` instead", "0.12.0")
  def typesafeRepo(status: String) = Resolver.typesafeIvyRepo(status)

  lazy val sbtPluginReleases = Resolver.sbtPluginRepo("releases")
  lazy val sbtPluginSnapshots = Resolver.sbtPluginRepo("snapshots")

  def modifyForPlugin(plugin: Boolean, dep: ModuleID): ModuleID =
    if (plugin) dep.copy(configurations = Some(Provided.name)) else dep

  @deprecated("Explicitly specify the organization using the other variant.", "0.13.0")
  def autoLibraryDependency(auto: Boolean, plugin: Boolean, version: String): Seq[ModuleID] =
    if (auto)
      modifyForPlugin(plugin, ScalaArtifacts.libraryDependency(version)) :: Nil
    else
      Nil
  def autoLibraryDependency(auto: Boolean, plugin: Boolean, org: String, version: String): Seq[ModuleID] =
    if (auto)
      modifyForPlugin(plugin, ModuleID(org, ScalaArtifacts.LibraryID, version)) :: Nil
    else
      Nil
  def addUnmanagedLibrary: Seq[Setting[_]] = Seq(
    unmanagedJars in Compile <++= unmanagedScalaLibrary
  )
  def unmanagedScalaLibrary: Initialize[Task[Seq[File]]] =
    Def.taskDyn {
      if (autoScalaLibrary.value && scalaHome.value.isDefined)
        Def.task { scalaInstance.value.libraryJar :: Nil }
      else
        Def.task { Nil }
    }

  import DependencyFilter._
  def managedJars(config: Configuration, jarTypes: Set[String], up: UpdateReport): Classpath =
    up.filter(configurationFilter(config.name) && artifactFilter(`type` = jarTypes)).toSeq.map {
      case (conf, module, art, file) =>
        Attributed(file)(AttributeMap.empty.put(artifact.key, art).put(moduleID.key, module).put(configuration.key, config))
    } distinct;

  def findUnmanagedJars(config: Configuration, base: File, filter: FileFilter, excl: FileFilter): Classpath =
    (base * (filter -- excl) +++ (base / config.name).descendantsExcept(filter, excl)).classpath

  @deprecated("Specify the classpath that includes internal dependencies", "0.13.0")
  def autoPlugins(report: UpdateReport): Seq[String] = autoPlugins(report, Nil)
  def autoPlugins(report: UpdateReport, internalPluginClasspath: Seq[File]): Seq[String] =
    {
      val pluginClasspath = report.matching(configurationFilter(CompilerPlugin.name)) ++ internalPluginClasspath
      val plugins = sbt.internal.inc.classpath.ClasspathUtilities.compilerPlugins(pluginClasspath)
      plugins.map("-Xplugin:" + _.getAbsolutePath).toSeq
    }

  private[this] lazy val internalCompilerPluginClasspath: Initialize[Task[Classpath]] =
    (thisProjectRef, settingsData, buildDependencies) flatMap { (ref, data, deps) =>
      internalDependencies0(ref, CompilerPlugin, CompilerPlugin, data, deps)
    }

  lazy val compilerPluginConfig = Seq(
    scalacOptions := {
      val options = scalacOptions.value
      val newPlugins = autoPlugins(update.value, internalCompilerPluginClasspath.value.files)
      val existing = options.toSet
      if (autoCompilerPlugins.value) options ++ newPlugins.filterNot(existing) else options
    }
  )

  @deprecated("Doesn't properly handle non-standard Scala organizations.", "0.13.0")
  def substituteScalaFiles(scalaInstance: ScalaInstance, report: UpdateReport): UpdateReport =
    substituteScalaFiles(scalaInstance, ScalaArtifacts.Organization, report)

  @deprecated("Directly provide the jar files per Scala version.", "0.13.0")
  def substituteScalaFiles(scalaInstance: ScalaInstance, scalaOrg: String, report: UpdateReport): UpdateReport =
    substituteScalaFiles(scalaOrg, report)(const(scalaInstance.allJars))

  def substituteScalaFiles(scalaOrg: String, report: UpdateReport)(scalaJars: String => Seq[File]): UpdateReport =
    report.substitute { (configuration, module, arts) =>
      if (module.organization == scalaOrg) {
        val jarName = module.name + ".jar"
        val replaceWith = scalaJars(module.revision).filter(_.getName == jarName).map(f => (Artifact(f.getName.stripSuffix(".jar")), f))
        if (replaceWith.isEmpty) arts else replaceWith
      } else
        arts
    }

  // try/catch for supporting earlier launchers
  def bootIvyHome(app: xsbti.AppConfiguration): Option[File] =
    try { Option(app.provider.scalaProvider.launcher.ivyHome) }
    catch { case _: NoSuchMethodError => None }

  def bootChecksums(app: xsbti.AppConfiguration): Seq[String] =
    try { app.provider.scalaProvider.launcher.checksums.toSeq }
    catch { case _: NoSuchMethodError => IvySbt.DefaultChecksums }

  def isOverrideRepositories(app: xsbti.AppConfiguration): Boolean =
    try app.provider.scalaProvider.launcher.isOverrideRepositories
    catch { case _: NoSuchMethodError => false }

  /** Loads the `appRepositories` configured for this launcher, if supported. */
  def appRepositories(app: xsbti.AppConfiguration): Option[Seq[Resolver]] =
    try { Some(app.provider.scalaProvider.launcher.appRepositories.toSeq map bootRepository) }
    catch { case _: NoSuchMethodError => None }

  def bootRepositories(app: xsbti.AppConfiguration): Option[Seq[Resolver]] =
    try { Some(app.provider.scalaProvider.launcher.ivyRepositories.toSeq map bootRepository) }
    catch { case _: NoSuchMethodError => None }

  private[this] def mavenCompatible(ivyRepo: xsbti.IvyRepository): Boolean =
    try { ivyRepo.mavenCompatible }
    catch { case _: NoSuchMethodError => false }

  private[this] def skipConsistencyCheck(ivyRepo: xsbti.IvyRepository): Boolean =
    try { ivyRepo.skipConsistencyCheck }
    catch { case _: NoSuchMethodError => false }

  private[this] def descriptorOptional(ivyRepo: xsbti.IvyRepository): Boolean =
    try { ivyRepo.descriptorOptional }
    catch { case _: NoSuchMethodError => false }

  private[this] def bootRepository(repo: xsbti.Repository): Resolver =
    {
      import xsbti.Predefined
      repo match {
        case m: xsbti.MavenRepository => MavenRepository(m.id, m.url.toString)
        case i: xsbti.IvyRepository =>
          val patterns = Patterns(i.ivyPattern :: Nil, i.artifactPattern :: Nil, mavenCompatible(i), descriptorOptional(i), skipConsistencyCheck(i))
          i.url.getProtocol match {
            case "file" =>
              // This hackery is to deal suitably with UNC paths on Windows. Once we can assume Java7, Paths should save us from this.
              val file = try { new File(i.url.toURI) } catch { case e: java.net.URISyntaxException => new File(i.url.getPath) }
              Resolver.file(i.id, file)(patterns)
            case _ => Resolver.url(i.id, i.url)(patterns)
          }
        case p: xsbti.PredefinedRepository => p.id match {
          case Predefined.Local                => Resolver.defaultLocal
          case Predefined.MavenLocal           => Resolver.mavenLocal
          case Predefined.MavenCentral         => DefaultMavenRepository
          case Predefined.ScalaToolsReleases   => Resolver.ScalaToolsReleases
          case Predefined.ScalaToolsSnapshots  => Resolver.ScalaToolsSnapshots
          case Predefined.SonatypeOSSReleases  => Resolver.sonatypeRepo("releases")
          case Predefined.SonatypeOSSSnapshots => Resolver.sonatypeRepo("snapshots")
          case unknown                         => sys.error("Unknown predefined resolver '" + unknown + "'.  This resolver may only be supported in newer sbt versions.")
        }
      }
    }
}

trait BuildExtra extends BuildCommon with DefExtra {
  import Defaults._

  /**
   * Defines an alias given by `name` that expands to `value`.
   * This alias is defined globally after projects are loaded.
   * The alias is undefined when projects are unloaded.
   * Names are restricted to be either alphanumeric or completely symbolic.
   * As an exception, '-' and '_' are allowed within an alphanumeric name.
   */
  def addCommandAlias(name: String, value: String): Seq[Setting[State => State]] =
    {
      val add = (s: State) => BasicCommands.addAlias(s, name, value)
      val remove = (s: State) => BasicCommands.removeAlias(s, name)
      def compose(setting: SettingKey[State => State], f: State => State) = setting in Global ~= (_ compose f)
      Seq(compose(onLoad, add), compose(onUnload, remove))
    }

  /**
   * Adds Maven resolver plugin.
   */
  def addMavenResolverPlugin: Setting[Seq[ModuleID]] =
    libraryDependencies += sbtPluginExtra(ModuleID("org.scala-sbt", "sbt-maven-resolver", sbtVersion.value), sbtBinaryVersion.value, scalaBinaryVersion.value)

  /**
   * Adds `dependency` as an sbt plugin for the specific sbt version `sbtVersion` and Scala version `scalaVersion`.
   * Typically, use the default values for these versions instead of specifying them explicitly.
   */
  def addSbtPlugin(dependency: ModuleID, sbtVersion: String, scalaVersion: String): Setting[Seq[ModuleID]] =
    libraryDependencies += sbtPluginExtra(dependency, sbtVersion, scalaVersion)

  /**
   * Adds `dependency` as an sbt plugin for the specific sbt version `sbtVersion`.
   * Typically, use the default value for this version instead of specifying it explicitly.
   */
  def addSbtPlugin(dependency: ModuleID, sbtVersion: String): Setting[Seq[ModuleID]] =
    libraryDependencies <+= (scalaBinaryVersion in update) { scalaV => sbtPluginExtra(dependency, sbtVersion, scalaV) }

  /**
   * Adds `dependency` as an sbt plugin for the sbt and Scala versions configured by
   * `sbtBinaryVersion` and `scalaBinaryVersion` scoped to `update`.
   */
  def addSbtPlugin(dependency: ModuleID): Setting[Seq[ModuleID]] =
    libraryDependencies <+= (sbtBinaryVersion in update, scalaBinaryVersion in update) { (sbtV, scalaV) => sbtPluginExtra(dependency, sbtV, scalaV) }

  /** Transforms `dependency` to be in the auto-compiler plugin configuration. */
  def compilerPlugin(dependency: ModuleID): ModuleID =
    dependency.copy(configurations = Some("plugin->default(compile)"))

  /** Adds `dependency` to `libraryDependencies` in the auto-compiler plugin configuration. */
  def addCompilerPlugin(dependency: ModuleID): Setting[Seq[ModuleID]] =
    libraryDependencies += compilerPlugin(dependency)

  /** Constructs a setting that declares a new artifact `a` that is generated by `taskDef`. */
  def addArtifact(a: Artifact, taskDef: TaskKey[File]): SettingsDefinition =
    {
      val pkgd = packagedArtifacts := packagedArtifacts.value updated (a, taskDef.value)
      Seq(artifacts += a, pkgd)
    }
  /** Constructs a setting that declares a new artifact `artifact` that is generated by `taskDef`. */
  def addArtifact(artifact: Initialize[Artifact], taskDef: Initialize[Task[File]]): SettingsDefinition =
    {
      val artLocal = SettingKey.local[Artifact]
      val taskLocal = TaskKey.local[File]
      val art = artifacts := artLocal.value +: artifacts.value
      val pkgd = packagedArtifacts := packagedArtifacts.value updated (artLocal.value, taskLocal.value)
      Seq(artLocal := artifact.value, taskLocal := taskDef.value, art, pkgd)
    }

  // because this was commonly used, this might need to be kept longer than usual
  @deprecated("In build.sbt files, this call can be removed.  In other cases, this can usually be replaced by Seq.", "0.13.0")
  def seq(settings: Setting[_]*): SettingsDefinition = new Def.SettingList(settings)

  def externalIvySettings(file: Initialize[File] = inBase("ivysettings.xml"), addMultiResolver: Boolean = true): Setting[Task[IvyConfiguration]] =
    externalIvySettingsURI(file(_.toURI), addMultiResolver)
  def externalIvySettingsURL(url: URL, addMultiResolver: Boolean = true): Setting[Task[IvyConfiguration]] =
    externalIvySettingsURI(Def.value(url.toURI), addMultiResolver)
  def externalIvySettingsURI(uri: Initialize[URI], addMultiResolver: Boolean = true): Setting[Task[IvyConfiguration]] =
    {
      val other = (baseDirectory, appConfiguration, projectResolver, updateOptions, streams).identityMap
      ivyConfiguration <<= (uri zipWith other) {
        case (u, otherTask) =>
          otherTask map {
            case (base, app, pr, uo, s) =>
              val extraResolvers = if (addMultiResolver) pr :: Nil else Nil
              new ExternalIvyConfiguration(base, u, Option(lock(app)), extraResolvers, uo, s.log)
          }
      }
    }
  private[this] def inBase(name: String): Initialize[File] = Def.setting { baseDirectory.value / name }

  def externalIvyFile(file: Initialize[File] = inBase("ivy.xml"), iScala: Initialize[Option[IvyScala]] = ivyScala): Setting[Task[ModuleSettings]] =
    moduleSettings := new IvyFileConfiguration(file.value, iScala.value, ivyValidate.value, managedScalaInstance.value)
  def externalPom(file: Initialize[File] = inBase("pom.xml"), iScala: Initialize[Option[IvyScala]] = ivyScala): Setting[Task[ModuleSettings]] =
    moduleSettings := new PomConfiguration(file.value, ivyScala.value, ivyValidate.value, managedScalaInstance.value)

  def runInputTask(config: Configuration, mainClass: String, baseArguments: String*): Initialize[InputTask[Unit]] =
    inputTask { result =>
      (fullClasspath in config, runner in (config, run), streams, result) map { (cp, r, s, args) =>
        toError(r.run(mainClass, data(cp), baseArguments ++ args, s.log))
      }
    }
  def runTask(config: Configuration, mainClass: String, arguments: String*): Initialize[Task[Unit]] =
    (fullClasspath in config, runner in (config, run), streams) map { (cp, r, s) =>
      toError(r.run(mainClass, data(cp), arguments, s.log))
    }

  def fullRunInputTask(scoped: InputKey[Unit], config: Configuration, mainClass: String, baseArguments: String*): Setting[InputTask[Unit]] =
    scoped <<= inputTask { result =>
      (initScoped(scoped.scopedKey, runnerInit) zipWith (fullClasspath in config, streams, result).identityMap) { (rTask, t) =>
        (t, rTask) map {
          case ((cp, s, args), r) =>
            toError(r.run(mainClass, data(cp), baseArguments ++ args, s.log))
        }
      }
    }
  def fullRunTask(scoped: TaskKey[Unit], config: Configuration, mainClass: String, arguments: String*): Setting[Task[Unit]] =
    scoped <<= (initScoped(scoped.scopedKey, runnerInit) zipWith (fullClasspath in config, streams).identityMap) {
      case (rTask, t) =>
        (t, rTask) map {
          case ((cp, s), r) =>
            toError(r.run(mainClass, data(cp), arguments, s.log))
        }
    }
  def initScoped[T](sk: ScopedKey[_], i: Initialize[T]): Initialize[T] = initScope(fillTaskAxis(sk.scope, sk.key), i)
  def initScope[T](s: Scope, i: Initialize[T]): Initialize[T] = i mapReferenced Project.mapScope(Scope.replaceThis(s))

  /**
   * Disables post-compilation hook for determining tests for tab-completion (such as for 'test-only').
   * This is useful for reducing test:compile time when not running test.
   */
  def noTestCompletion(config: Configuration = Test): Setting[_] = inConfig(config)(Seq(definedTests <<= detectTests)).head

  def filterKeys(ss: Seq[Setting[_]], transitive: Boolean = false)(f: ScopedKey[_] => Boolean): Seq[Setting[_]] =
    ss filter (s => f(s.key) && (!transitive || s.dependencies.forall(f)))
}
trait DefExtra {
  private[this] val ts: TaskSequential = new TaskSequential {}
  implicit def toTaskSequential(d: Def.type): TaskSequential = ts
}
trait BuildCommon {
  @deprecated("Use Def.inputTask with the `Def.spaceDelimited()` parser.", "0.13.0")
  def inputTask[T](f: TaskKey[Seq[String]] => Initialize[Task[T]]): Initialize[InputTask[T]] =
    InputTask.apply(Def.value((s: State) => Def.spaceDelimited()))(f)

  /**
   * Allows a String to be used where a `NameFilter` is expected.
   * Asterisks (`*`) in the string are interpreted as wildcards.
   * All other characters must match exactly.  See [[sbt.GlobFilter]].
   */
  implicit def globFilter(expression: String): NameFilter = GlobFilter(expression)

  implicit def richAttributed(s: Seq[Attributed[File]]): RichAttributed = new RichAttributed(s)
  implicit def richFiles(s: Seq[File]): RichFiles = new RichFiles(s)
  implicit def richPathFinder(s: PathFinder): RichPathFinder = new RichPathFinder(s)
  final class RichPathFinder private[sbt] (s: PathFinder) {
    /** Converts the `PathFinder` to a `Classpath`, which is an alias for `Seq[Attributed[File]]`. */
    def classpath: Classpath = Attributed blankSeq s.get
  }
  final class RichAttributed private[sbt] (s: Seq[Attributed[File]]) {
    /** Extracts the plain `Seq[File]` from a Classpath (which is a `Seq[Attributed[File]]`).*/
    def files: Seq[File] = Attributed.data(s)
  }
  final class RichFiles private[sbt] (s: Seq[File]) {
    /** Converts the `Seq[File]` to a Classpath, which is an alias for `Seq[Attributed[File]]`. */
    def classpath: Classpath = Attributed blankSeq s
  }
  def toError(o: Option[String]): Unit = o foreach error

  def overrideConfigs(cs: Configuration*)(configurations: Seq[Configuration]): Seq[Configuration] =
    {
      val existingName = configurations.map(_.name).toSet
      val newByName = cs.map(c => (c.name, c)).toMap
      val overridden = configurations map { conf => newByName.getOrElse(conf.name, conf) }
      val newConfigs = cs filter { c => !existingName(c.name) }
      overridden ++ newConfigs
    }

  // these are intended for use in input tasks for creating parsers
  def getFromContext[T](task: TaskKey[T], context: ScopedKey[_], s: State): Option[T] =
    SessionVar.get(SessionVar.resolveContext(task.scopedKey, context.scope, s), s)

  def loadFromContext[T](task: TaskKey[T], context: ScopedKey[_], s: State)(implicit f: sbinary.Format[T]): Option[T] =
    SessionVar.load(SessionVar.resolveContext(task.scopedKey, context.scope, s), s)

  // intended for use in constructing InputTasks
  def loadForParser[P, T](task: TaskKey[T])(f: (State, Option[T]) => Parser[P])(implicit format: sbinary.Format[T]): Initialize[State => Parser[P]] =
    loadForParserI(task)(Def value f)(format)
  def loadForParserI[P, T](task: TaskKey[T])(init: Initialize[(State, Option[T]) => Parser[P]])(implicit format: sbinary.Format[T]): Initialize[State => Parser[P]] =
    (resolvedScoped, init)((ctx, f) => (s: State) => f(s, loadFromContext(task, ctx, s)(format)))

  def getForParser[P, T](task: TaskKey[T])(init: (State, Option[T]) => Parser[P]): Initialize[State => Parser[P]] =
    getForParserI(task)(Def value init)
  def getForParserI[P, T](task: TaskKey[T])(init: Initialize[(State, Option[T]) => Parser[P]]): Initialize[State => Parser[P]] =
    (resolvedScoped, init)((ctx, f) => (s: State) => f(s, getFromContext(task, ctx, s)))

  // these are for use for constructing Tasks
  def loadPrevious[T](task: TaskKey[T])(implicit f: sbinary.Format[T]): Initialize[Task[Option[T]]] =
    (state, resolvedScoped) map { (s, ctx) => loadFromContext(task, ctx, s)(f) }

  def getPrevious[T](task: TaskKey[T]): Initialize[Task[Option[T]]] =
    (state, resolvedScoped) map { (s, ctx) => getFromContext(task, ctx, s) }

  private[sbt] def derive[T](s: Setting[T]): Setting[T] = Def.derive(s, allowDynamic = true, trigger = _ != streams.key, default = true)
}
