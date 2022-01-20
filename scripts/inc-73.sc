import java.util.concurrent.TimeUnit
import java.util.concurrent.Executors
import java.util.concurrent.ArrayBlockingQueue
import java.io.PrintStream
import java.util.Base64
import java.nio.file.Paths
import java.nio.file.Files
import java.nio.ByteBuffer
import scala.jdk.CollectionConverters._
import scala.sys.process._

def base64(id: String): String = new String(
  Base64.getEncoder.encode(s"Repository:$id".getBytes)
)

val lines = Files
  .readAllLines(Paths.get("/Users/olafurpg/Downloads/repo_id_versions.csv"))
  .asScala
  .view.drop(1)
  .map(_.split(","))
val donePath = Paths.get("/Users/olafurpg/Downloads/done.txt")
val isDone = Files.readAllLines(donePath).asScala.toSet
val done = new PrintStream(Files.newOutputStream(donePath))
lines .foreach {
  case Array(id @ "54575954", version) if !isDone(id) =>
    val repo = base64(id)
    val rev = s"v$version"
    println(s"$repo $version")
    val result = List(
      "src",
      "api",
      "-query", "mutation QueueAutoIndexJobsForRepo($id: ID!, $rev: String) { queueAutoIndexJobsForRepo(repository: $id, rev: $rev) { id } }",
      s"id=$repo",
      s"rev=$rev"
    ).!!
    println(result)
    done.println(id)
  case _ =>
}

