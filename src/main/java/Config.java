import com.intellij.openapi.components.PersistentStateComponent;
import com.intellij.openapi.components.State;
import com.intellij.openapi.components.Storage;
import com.intellij.openapi.project.Project;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

@State(
  name = "Config",
  storages = {@Storage("sourcegraph.xml")})
class Config implements PersistentStateComponent<Config> {

    public String url;

    public String getUrl() {
        return url;
    }

    public String defaultBranch;

    public String getDefaultBranch() {
        return defaultBranch;
    }

    public String remoteUrlReplacements;

    public String getRemoteUrlReplacements() {
        return remoteUrlReplacements;
    }

    @Nullable
    @Override
    public Config getState() {
        return this;
    }

    @Override
    public void loadState(@NotNull Config config) {
        this.url = config.url;
        this.defaultBranch = config.defaultBranch;
        this.remoteUrlReplacements = config.remoteUrlReplacements;
    }

    static Config getInstance(Project project) {
        return project.getService(Config.class);
    }
}
