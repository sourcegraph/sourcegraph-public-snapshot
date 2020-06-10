import com.intellij.openapi.components.PersistentStateComponent;
import com.intellij.openapi.components.ServiceManager;
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

    @Nullable
    @Override
    public Config getState() {
        return this;
    }

    @Override
    public void loadState(@NotNull Config config) {
        this.url = config.url;
    }

    @Nullable static Config getInstance(Project project) {
        return ServiceManager.getService(project, Config.class);
    }
}
