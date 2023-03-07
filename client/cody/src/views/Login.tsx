import { useState } from 'react';
import './App.css';
import {
  VSCodeTextField,
  VSCodeButton,
} from '@vscode/webview-ui-toolkit/react';
import { Terms } from './About';

interface LoginProps {
  setToken: (token: string) => void;
  setEndpoint: (endpoint: string) => void;
  onTokenSubmit: () => void;
}

export const Login: React.FunctionComponent<
  React.PropsWithChildren<LoginProps>
> = ({ setToken, setEndpoint, onTokenSubmit }) => {
  const [termsAccepted, setTermsAccepted] = useState(false);

  return (
    <div className="inner-container">
      <div className="non-transcript-container">
        {termsAccepted ? (
          <div className="container-getting-started">
            <h1>Login</h1>
            <p>Enter Access Token</p>
            <VSCodeTextField
              placeholder="ex 4923weaWEH"
              className="w-100"
              type="password"
              onInput={(e: any) => setToken(e.target.value)}
            />
            <p>Set Endpoint</p>
            <VSCodeTextField
              value="https://cody.sg.org"
              placeholder="https://cody.sg.org"
              className="w-100"
              autofocus={true}
              onInput={(e: any) => setEndpoint(e.target.value)}
            />
            <ul className="tab-menu-container mt-5">
              <VSCodeButton
                className={'login-btn'}
                type="button"
                onClick={() => onTokenSubmit()}
              >
                Submit
              </VSCodeButton>
            </ul>
          </div>
        ) : (
          <div className="container-getting-started">
            <Terms setTermsAccepted={setTermsAccepted} />
          </div>
        )}
      </div>
    </div>
  );
};

export default Login;
