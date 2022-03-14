import com.intellij.openapi.actionSystem.AnActionEvent;
import org.jetbrains.annotations.NotNull;

public class Search extends SearchActionBase {
    @Override
    public void actionPerformed(@NotNull AnActionEvent e) {
        super.actionPerformedMode(e, "search");
    }
}
