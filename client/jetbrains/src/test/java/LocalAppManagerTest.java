import static org.junit.jupiter.api.Assertions.*;

import com.sourcegraph.cody.localapp.LocalAppManager;
import org.junit.jupiter.api.Test;

public class LocalAppManagerTest {
  // these tests are just a sanity check to make sure the LocalAppManager API doesn't throw random
  // exceptions, not verifying if they actually work for now

  @Test
  public void localAppInstallCheckDoesntFail() {
    LocalAppManager.isLocalAppInstalled();
  }

  @Test
  public void localAppRunningCheckDoesntFail() {
    LocalAppManager.isLocalAppRunning();
  }

  @Test
  public void localAppInfoGettingDoesntFail() {
    LocalAppManager.getLocalAppInfo();
  }

  @Test
  public void localAppAccessTokenGettingDoesntFail() {
    LocalAppManager.getLocalAppAccessToken();
  }

  @Test
  public void localAppUrlGettingDoesntFail() {
    LocalAppManager.getLocalAppUrl();
  }
}
