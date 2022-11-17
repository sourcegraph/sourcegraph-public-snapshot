package com.sourcegraph.vcs;

import com.intellij.openapi.project.Project;
import com.intellij.openapi.util.Pair;
import com.intellij.openapi.vcs.VcsException;
import com.intellij.openapi.vfs.VirtualFile;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;
import org.jetbrains.idea.perforce.application.PerforceVcs;
import org.jetbrains.idea.perforce.perforce.connections.P4Connection;

import java.util.Collection;

public class PerforceUtil {

    @NotNull
    // Returned format: perforce@perforce.company.com:depot-name.perforce
    public static String getRemoteRepoUrl(@NotNull Project project, @NotNull VirtualFile file) throws Exception {
        PerforceVcs vcs = PerforceVcs.getInstance(project);
        Collection<Pair<P4Connection, Collection<VirtualFile>>> rootsByConnections = vcs.getRootsByConnections();
        P4Connection connection = rootsByConnections.stream().filter(
            (Pair<P4Connection, Collection<VirtualFile>> pair) -> pair.getSecond().stream().anyMatch(
                (VirtualFile root) -> file.getPath().startsWith(root.getPath()))).map(x -> Pair.getFirst(x)).findFirst().orElse(null);

        if (connection == null) {
            throw new Exception("No Perforce connection found.");
        }

        String serverUrl = connection.getConnectionKey().getServer();
        // Remove port if present
        String serverName = serverUrl.split(":")[0];

        String depotName = getDepotName(project, connection, file);

        if (depotName == null) {
            throw new Exception("No depot name found.");
        }

        return "perforce@" + serverName + ":" + depotName + ".perforce";
    }

    @Nullable
    private static String getDepotName(@NotNull Project project, @NotNull P4Connection connection, @NotNull VirtualFile file) throws VcsException {
        PerforceVcs vcs = PerforceVcs.getInstance(project);
        Collection<Pair<P4Connection, Collection<VirtualFile>>> rootsByConnections = vcs.getRootsByConnections();
        Pair<P4Connection, Collection<VirtualFile>> pair = rootsByConnections.stream().filter(
            (Pair<P4Connection, Collection<VirtualFile>> x) -> x.getFirst() == connection).findFirst().orElse(null);
        if (pair == null) {
            return null;
        }
        VirtualFile root = pair.getSecond().stream().filter((VirtualFile x) -> file.getPath().startsWith(x.getPath())).findFirst().orElse(null);
        if (root == null) {
            return null;
        }

        String relativePath = file.getPath().substring(root.getPath().length());
        if (relativePath.startsWith("/")) {
            relativePath = relativePath.substring(1);
        }
        return relativePath.trim().split("/")[0];
    }
}
