public class RepoInfo {
    public String fileRel;
    public String remoteURL;
    public String branch;

    public RepoInfo(String sFileRel, String sRemoteURL, String sBranch) {
        fileRel = sFileRel;
        remoteURL = sRemoteURL;
        branch = sBranch;
    }
}
