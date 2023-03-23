//
//  ContentView.swift
//  Sourcegraph App
//
//  Created by Peter Guy on 2/10/23.
//

import SwiftUI

private var token: Any?

struct ContentView: View {
    @AppStorage("windowWidth")
    var windowWidth: Double = 400.0
    @AppStorage("windowHeight")
    var windowHeight: Double = 200.0
    
    var windowSize: CGSize {
        CGSize(width: windowWidth, height: windowHeight)
    }
    
    private var STATUS_STOPPED: Int = 0
    private var STATUS_STARTING: Int = 1
    private var STATUS_RUNNING: Int = 2
    
    @State private var previousStatus: Int = 0
    
    init() {
        
    }
    
    // read the log file to detect startup
    // replaced by an http call
    // might still be needed to detect FATAL errors for display
//    func observeApp() {
//
//        let stdoutFH = stdout.fileHandleForReading
//        stdoutFH.waitForDataInBackgroundAndNotify()
//
//        // https://www.tekramer.com/observing-real-time-ouput-from-shell-commands-in-a-swift-script
//        let notificationCenter = NotificationCenter.default
//        let dataNotificationName = NSNotification.Name.NSFileHandleDataAvailable
//        var _ = notificationCenter.addObserver(forName: dataNotificationName, object: stdoutFH, queue: nil) {  notification in
//            let data = stdoutFH.availableData
//            guard data.count > 0 else {
//                return
//            }
//            let string = String(data: data, encoding: .utf8) ?? ""
//            let lines = string.split(whereSeparator: \.isNewline)
//            if lines.count > 0 && startupProgress < 100 {
//                DispatchQueue.main.async {
//                    startupProgress += Double(lines.count * 2)
//                }
//            }
//            if let _ = string.range(of: "Sourcegraph is now available") {
//
//                animateRunning()
//
//                DispatchQueue.main.async {
//                    startupProgress = 100
//                }
//                return
//            }
//            stdoutFH.waitForDataInBackgroundAndNotify()
//        }
//    }
    
    private var animationDuration: Double = 1.5
    
    @State private var watchdogRunning: Bool = false
    
    @State private var stoppedOpacity = 0.0
    @State private var startingOpacity = 0.0
    @State private var runningOpacity = 0.0
    @State private var testOpacity: Double = 0.0
    
    @Environment(\.colorScheme) var colorScheme
    
    var body: some View {
        // Grid is macOS 13.0+; may need to switch to Lazy[VH]Grid to support macOS < 13
//        LazyVGrid(columns: <#T##[GridItem]#>, alignment: .leading, content: <#T##() -> _#>)
        Grid(alignment: .leading) {
            ZStack {
                RoundedRectangle(cornerRadius: 4)
                    .fill(colorScheme == .dark ? Color(red: 0.2, green: 0.2, blue: 0.2) : Color(.controlBackgroundColor))
                    .frame(maxHeight: 50)
                    .overlay(
                        RoundedRectangle(cornerRadius: 4)
                            .stroke(Color(red: 0.3, green: 0.3, blue: 0.3), lineWidth: 1)
                    )
                
                HStack {
                    ProgressView().progressViewStyle(CircularProgressViewStyle())
                    Text("Starting Sourcegraph...").padding(.leading, 10).font(.title2)
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .opacity(self.startingOpacity)
                .padding()
                .onAppear {
                    if appIsStarting() {
                        self.startingOpacity = 1.0
                    }
                }
                HStack {
                    Text("Sourcegraph is running.")
                        .font(.title2)
                        .foregroundColor(Color(red: 141/255, green: 238/255, blue: 156/255))
                        .padding()
                    
                    Spacer()
                    
                    Button(action: {
                        StopApp()
                    }) {
                        Text("Stop").frame(maxWidth: 60)
                    }
                    .disabled(self.runningOpacity == 0.0)
                    .gridColumnAlignment(.trailing)
                    
                    Button(action: {
                        if let url = URL(string: "http://127.0.0.1:3080") {
                            NSWorkspace.shared.open(url)
                        }
                    }) {
                        Text("Open").frame(maxWidth: 60)
                    }
                    .disabled(self.runningOpacity == 0.0)
                    .padding()
                }
                .frame(maxWidth: .infinity, alignment: .leading)
                .opacity(self.runningOpacity)
                .onAppear {
                    if appIsRunning() {
                        self.runningOpacity = 1.0
                    }
                }
                
                HStack {
                    Text("Sourcegraph is stopped.")
                        .font(.title2)
                        .foregroundColor(Color(red: 238/255, green: 187/255, blue: 141/255))
                    
                    Spacer()
                    
                    Button(action: {
                        do {
                            try StartApp()
                        } catch {
                        }
                    }) {
                        Text("Start").frame(maxWidth: 60)
                    }
                    .disabled(self.startingOpacity == 1.0 || self.runningOpacity == 1.0)
                    .gridColumnAlignment(.trailing)
                }
                .opacity(self.stoppedOpacity)
                .onAppear {
                    if appIsStopped() {
                            self.stoppedOpacity = 1.0
                    }
                }
                .padding()
                .frame(maxWidth: .infinity, alignment: .leading)
                
            }
            .ignoresSafeArea()
            
            // the buttons used to all be here before moving them into the same location as the loading display
//            GridRow {
//                VStack(alignment: .leading) {
//                    Text("Start Sourcegraph App")
//                    Text("Start the stopped server.")
//                        .font(.subheadline)
//                            .foregroundColor(.secondary)
//                }
//                Button(action: {
//                    do {
//                        try StartApp()
//                    } catch {
//                    }
//                }) {
//                    Text("Start").frame(maxWidth: 60)
//                }
//                .disabled(self.startingOpacity == 1.0 || self.runningOpacity == 1.0)
//                .gridColumnAlignment(.trailing)
//            }
//            Divider()
            
//            GridRow {
//                VStack(alignment: .leading) {
//                    Text("Open Sourcegraph App")
//                    Text("Launch in default browser.")
//                        .font(.subheadline)
//                            .foregroundColor(.secondary)
//                }
//                Button(action: {
//                    if let url = URL(string: "http://127.0.0.1:3080") {
//                        NSWorkspace.shared.open(url)
//                    }
//                }) {
//                    Text("Open").frame(maxWidth: 60)
//                }
//                .disabled(self.runningOpacity == 0.0)
//                .gridColumnAlignment(.trailing)
//            }
//            Divider()
            
//            GridRow {
//                VStack(alignment: .leading) {
//                    Text("Stop Sourcegraph App")
//                    Text("Stop the running server.")
//                        .font(.subheadline)
//                            .foregroundColor(.secondary)
//                }
//                Button(action: {
//                    StopApp()
//                    animateStopping()
//                    self.startupProgress = 0.0
//                }) {
//                    Text("Stop").frame(maxWidth: 60)
//                }
//                .disabled(self.runningOpacity == 0.0)
//                .gridColumnAlignment(.trailing)
//            }
//            Divider()
            
//            GridRow {
//                VStack(alignment: .leading) {
//                    Text("Restart Sourcegraph App")
//                    Text("All settings will be preserved.")
//                        .font(.subheadline)
//                            .foregroundColor(.secondary)
//                }
//                Button(action: {
//                    StopApp()
//                    self.startupProgress = 0.0
//                    DispatchQueue.main.asyncAfter(deadline: .now() + 2.0) {
//                        do {
//                            try StartApp()
//                        } catch {
//                        }
//                    }
//                }) {
//                    Text("Restart").frame(maxWidth: 60)
//                }
//                .disabled(!appIsStarting())
//                .gridColumnAlignment(.trailing)
//            }
//            Divider()
            
            GridRow {
                VStack(alignment: .leading) {
                    Text("Show the log file")
                    Text("Open in Console.")
                        .font(.subheadline)
                            .foregroundColor(.secondary)
                }
                Button(action: {
                    let logFileURL = FileManager.default.homeDirectoryForCurrentUser.appendingPathComponent("Library/Application Support/sourcegraph-sp/sourcegraph.log")
                    if FileManager.default.fileExists(atPath: logFileURL.path) {
                        NSWorkspace.shared.open(logFileURL)
                    }
                }) {
                    Text("Show log").frame(maxWidth: 60)
                }
                .gridColumnAlignment(.trailing)
            }
            // we don't have an uninstaller yet
//            Divider()
//            GridRow {
//                VStack(alignment: .leading) {
//                    Text("Uninstall Sourcegraph App")
//                    Text("We're sorry to see you go. This completely uninstalls Sourcegraph App.")
//                        .font(.subheadline)
//                            .foregroundColor(.secondary)
//                }
//                .gridColumnAlignment(.leading)
//                Button(action: someVoid) {
//                    Text("Uninstall").frame(maxWidth: 60)
//                }
//                .gridColumnAlignment(.trailing)
//            }
        }
        .padding(.all, 20)
        .frame(minWidth: 460, idealWidth: 572, minHeight: 220, idealHeight: 282)
        .onAppear {
            
            // set the window size, but have to delay a bit to wait for the window to be created
            DispatchQueue.main.async {
                if let window = NSApp.windows.first {
                    window.setContentSize(windowSize)
                }
            }
            
            // this feels like a hack, but I can't figure out how/why to make the elements
            // read updated values in order to change how they display
            if !self.watchdogRunning {
                Timer.scheduledTimer(withTimeInterval: 0.3, repeats: true) { timer in
                    self.watchdogRunning = true
                    if appIsRunning() {
                        withAnimation(.easeInOut(duration: animationDuration)) {
                            self.stoppedOpacity = 0.0
                            self.startingOpacity = 0.0
                            self.runningOpacity = 1.0
                        }
                    } else if appIsStarting() {
                        withAnimation(.easeInOut(duration: animationDuration)) {
                            self.stoppedOpacity = 0.0
                            self.startingOpacity = 1.0
                            self.runningOpacity = 0.0
                        }
                    } else {
                        withAnimation(.easeInOut(duration: animationDuration)) {
                            self.stoppedOpacity = 1.0
                            self.startingOpacity = 0.0
                            self.runningOpacity = 0.0
                        }
                    }
                }
            }
        }
        .onChange(of: windowSize) { newSize in
            // this doesn't appear to fire at all when the window size changes
            windowWidth = Double(newSize.width)
            windowWidth = Double(newSize.width)
        }
        .onDisappear {
            // store the window size so it can be preserved next time the window appears
            // doesn't work - NSApp.windows.first is nil, even when run async
            DispatchQueue.main.async {
                if let window = NSApp.windows.first {
                    windowWidth = Double(window.frame.width)
                    windowHeight = Double(window.frame.height)
                }
            }
        }
    }
}

struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        ContentView()
    }
}
