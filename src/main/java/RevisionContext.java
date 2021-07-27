import com.intellij.openapi.project.Project;

public class RevisionContext {
  private final Project project;
  private final String revisionNumber;

  public RevisionContext(Project project, String revisionNumber) {
    this.project = project;
    this.revisionNumber = revisionNumber;
  }

  public Project getProject() {
    return project;
  }

  public String getRevisionNumber() {
    return revisionNumber;
  }
}
