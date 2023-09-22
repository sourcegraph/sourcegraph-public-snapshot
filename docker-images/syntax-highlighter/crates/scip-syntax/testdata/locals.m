% Copyright (c) 2021-2022, The MathWorks, Inc.
% All rights reserved.
% Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:
% 1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
% 2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.
% 3. In all cases, the software is, and all modifications and derivatives of the software shall be, licensed to you solely for use in conjunction with MathWorks products and service offerings. 
% THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

function setupPythonIfNeeded()
    %setupPythonIfNeeded Check if python is installed and configured.  If it's
    %not, download and install it

    % Python setup is only supported in R2019a (ver 9.6) and later 
    if verLessThan('matlab','9.6')
        error("setupPythonIfNeeded:unsupportedVersion","Only version R2019a and later are supported")
    end        

    % Check if MATLAB has python set up
    pythonInfo = pyenv;
    
    if pythonInfo.Version ~= ""
        % It's set up: return
        return
    end
    
    % Python is not set up, but it might be installed.  Check to see if it
    % is installed in any standard locations
    pythonVersionToLookFor = getSupportedPythonVersions();
    
    % Check for python executable in user directory (default install)
    pythonExecutable = locatePythonInstallations(getUserPythonDir(),pythonVersionToLookFor);
    
    if pythonExecutable == ""
        % Check for python executable in program files directory (all user install)
        pythonExecutable = locatePythonInstallations(getProgramFilesDir(),pythonVersionToLookFor);

        if pythonExecutable == ""
            % Trigger an install here
            
            filename = "python-3.9.7-amd64.exe";
            downloadPath = fullfile(tempdir,filename);
            installEXEPath = websave(downloadPath,"https://www.python.org/ftp/python/3.9.7/" + filename);
            status = system(installEXEPath);
            if status ~= 0
                error("setupPythonIfNeeded:pythonNotInstalled",'Could not install python.  <a href="https://www.python.org/ftp/python/3.9.7/python-3.9.7-amd64.exe">Download and install version 3.9</a>')
            end
            pythonExecutable = locatePythonInstallations(getUserPythonDir(),pythonVersionToLookFor);
            if pythonExecutable == ""
                pythonExecutable = locatePythonInstallations(getProgramFilesDir(),pythonVersionToLookFor);
                if pythonExecutable == ""
                    error("setupPythonIfNeeded:pythonNotInstalled",'Could not install python.  <a href="https://www.python.org/ftp/python/3.9.7/python-3.9.7-amd64.exe">Download and install version 3.9</a>')
                end
            end
        end
    end
    
    pythonInfo = pyenv('Version', pythonExecutable);

    if pythonInfo.Version == ""
        % Python was not set up correctly
        error("setupPythonIfNeeded:unableToSetPythonVersion","Could not configure python for use.  See ""Configure Your System to Use Python"" in MATLAB documentation.")
    end
    
    function pythonVersion = getSupportedPythonVersions()
        % It would be better to get this info from the internet somewhere,
        % but for now....
        % This is based on
        % https://www.mathworks.com/content/dam/mathworks/mathworks-dot-com/support/sysreq/files/python-compatibility.pdf
        versionInfo = ver('matlab');
        MATLABrelease = extract(string(versionInfo.Release),alphanumericBoundary + alphanumericsPattern + alphanumericBoundary);
        switch MATLABrelease
            case "R2023a"
                pythonVersion = ["3.8";"3.9";"3.10"];
            case "R2022b"
                pythonVersion = ["3.8";"3.9";"3.10"];
            case "R2022a"
                pythonVersion = ["3.8";"3.9"];
            case "R2021b"
                pythonVersion = ["3.7";"3.8";"3.9"];
            case "R2021a"
                pythonVersion = ["3.7";"3.8"];
            case "R2020b"
                pythonVersion = ["3.6";"3.7";"3.8"];
            case "R2020a"
                pythonVersion = ["3.6";"3.7"];
            case "R2019b"
                pythonVersion = ["3.6";"3.7"];
            case "R2019a"
                pythonVersion = ["3.5";"3.6";"3.7"];
            otherwise
                % Lets assume 3.9 will work for a while
                pythonVersion = "3.9";
        end
    end

    function pythonDir = getUserPythonDir()
        [status, localAppDataPath] = system("echo %LOCALAPPDATA%");
        if status ~= 0
            error("setupPythonIfNeeded:cantGetProgramFiles","Cannot retreive location of user directory.")
        end
        localAppDataPath = strtrim(string(localAppDataPath));
        pythonDir = fullfile(localAppDataPath,"Programs","Python");

    end

    function pythonDir = getProgramFilesDir()
        [status, programFilesPath] = system("echo %PROGRAMFILES%");
        if status ~= 0
            error("setupPythonIfNeeded:cantGetProgramFiles","Cannot retreive location of program files directory.")
        end
        programFilesPath = strtrim(string(programFilesPath));
        pythonDir = fullfile(programFilesPath,"Python");
    end

    function pythonExecutable = locatePythonInstallations(pythonDir,pythonVersionToLookFor)
        % sort the version numbers and transform to directory names, newest
        % to oldest
        pythonVersion = sortrows(split(pythonVersionToLookFor,"."),[1 2],'descend');
        pythonDirNames = "Python" + pythonVersion(:,1) + pythonVersion(:,2);
        pythonExecutable = "";
        for iDir = 1:numel(pythonDirNames)
            % Go through supported releases, see if there's a python.exe at each location, going from
            % latest release
            pythonExecutable = fullfile(pythonDir,pythonDirNames(iDir),"python.exe");
            if exist(pythonExecutable,'file')
                return
            end
        end
    end
end
