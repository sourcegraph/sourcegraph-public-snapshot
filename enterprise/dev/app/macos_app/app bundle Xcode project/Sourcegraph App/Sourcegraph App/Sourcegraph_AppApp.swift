//
//  Sourcegraph_AppApp.swift
//  Sourcegraph App
//
//  Created by Peter Guy on 2/10/23.
//

import SwiftUI
import Cocoa

func checkAppIsRunning() -> Bool {
    guard let url = URL(string: "http://127.0.0.1:3080/sign-in?returnTo=%2Fsearch") else { fatalError("Missing URL") }
    let urlRequest = URLRequest(url: url)
    var isRunning: Bool = false
    // use a semaphore to wait for the async data task to be done
    let sem = DispatchSemaphore.init(value: 0)
    let dataTask = URLSession.shared.dataTask(with: urlRequest) { (data, response, error) in
        // signal the semaphore when the data task completes
        defer { sem.signal() }
        if error == nil {
            guard let response = response as? HTTPURLResponse else { return }
            if response.statusCode == 200 {
                isRunning = true
            }
        }
    }
    dataTask.resume()
    // wait for the data task to be done
    sem.wait()
    return isRunning
}

// if the app process is not running, but the http connection is valid
// then some other app process is running
func otherAppIsRunning() -> Bool {
    return !appTask.isRunning && checkAppIsRunning()
}

func appIsStopped() -> Bool {
    return !appTask.isRunning
}

func appIsStarting() -> Bool {
    // this could also indicate a borked app where the process is running,
    // but the app is not accepting http connections
    return appTask.isRunning && !checkAppIsRunning()
}

func appIsRunning() -> Bool {
    return appTask.isRunning && checkAppIsRunning()
}

// https://stackoverflow.com/a/65162953
//let port = UInt16(10000)
//print(isPortOpen(port:port)
func isPortOpen(port: in_port_t) -> Bool {

    let socketFileDescriptor = socket(AF_INET, SOCK_STREAM, 0)
    if socketFileDescriptor == -1 {
        return false
    }

    var addr = sockaddr_in()
    let sizeOfSockkAddr = MemoryLayout<sockaddr_in>.size
    addr.sin_len = __uint8_t(sizeOfSockkAddr)
    addr.sin_family = sa_family_t(AF_INET)
    addr.sin_port = Int(OSHostByteOrder()) == OSLittleEndian ? _OSSwapInt16(port) : port
    addr.sin_addr = in_addr(s_addr: inet_addr("0.0.0.0"))
    addr.sin_zero = (0, 0, 0, 0, 0, 0, 0, 0)
    var bind_addr = sockaddr()
    memcpy(&bind_addr, &addr, Int(sizeOfSockkAddr))

    if Darwin.bind(socketFileDescriptor, &bind_addr, socklen_t(sizeOfSockkAddr)) == -1 {
        return false
    }
    let isOpen = listen(socketFileDescriptor, SOMAXCONN ) != -1
    Darwin.close(socketFileDescriptor)
    return isOpen
}

struct EOLBuffer {
    private var buffer = Data()

    mutating func append(_ data: Data) -> String? {
        buffer.append(data)
        if let string = String(data: buffer, encoding: .utf8), string.last?.isNewline == true {
            buffer.removeAll()
            return string
        }
        return nil
    }
}

var appTask: Process = Process()
var stdout = Pipe()


var statusBarItem: NSStatusItem!

let startItem = NSMenuItem()
let openItem = NSMenuItem()
let stopItem = NSMenuItem()
let logItem = NSMenuItem()

// the menubar icon is a Image Set in the Xcode assets
// that contains two icons: one light and one dark
// it needs to be rendered as a template to auto-adjust to the menu bar background color
// that rendering can be set in the Image Set's attributes using the Xcode Attributes pane
// or it can be set programmatically when it is created
let activeMenuBarIcon = NSImage(named: "ActiveMenubarIcon")
let inactiveMenuBarIcon = NSImage(named: "InactiveMenubarIcon")
let loadingMenuBarIcon = NSImage(named: "LoadingMenubarIcon")

extension String: LocalizedError {
    public var errorDescription: String? { return self }
}

func StartApp() throws {
    if appTask.isRunning {
        // disable buttons, but also explicitly gaurd against multiple launches
        return
    }
    if let shellScriptURL = Bundle.main.url(forResource: "sourcegraph_launcher", withExtension: "sh") {
        stdout = Pipe()
        appTask = Process()
        appTask.launchPath = "/bin/bash"
        appTask.arguments = ["-c", "\"/" +  shellScriptURL.resolvingSymlinksInPath().relativePath + "\""]
        appTask.standardInput = nil
        appTask.standardError = nil
        appTask.standardOutput = stdout
        try appTask.run()
    } else {
        throw "could not find app"
    }
}

// StopApp doesn't do any view updating, so it can live in the app
func StopApp() {
    // look for running app pid
    // kill it
    // shut down the database
    // TODO: also run this func when the macOS app is Quit
    if appTask.isRunning {
        appTask.terminate()
    }
}

@main
struct Sourcegraph_AppApp: App {
    init() {
        // placeholder to put stuff to run when the app starts
        do {
            try StartApp()
        } catch {
            print(error.localizedDescription)
        }
    }
    
    @NSApplicationDelegateAdaptor(AppDelegate.self) var appDelegate

    var body: some Scene {
        WindowGroup {
            ContentView()
        }
        // 13.0+
//        MenuBarExtra("Sourcegraph App Menu Bar Extra", image: "wildcard-black") {
//        }
    }
    
    class clickables {
        
        @objc func onStartClick(_ sender: Any?) {
            do {
                try StartApp()
            } catch {
                
            }
        }
    }
    
    class AppDelegate: NSObject, NSApplicationDelegate, NSWindowDelegate {
        
        func applicationDidFinishLaunching(_ notification: Notification) {
            
            // Create a new status bar item with the app icon
            statusBarItem = NSStatusBar.system.statusItem(withLength: NSStatusItem.variableLength)
            statusBarItem.button?.setFrameSize(NSSize(width: 60, height: 20))
            statusBarItem.button?.image = inactiveMenuBarIcon
            statusBarItem.button?.needsDisplay = true
            statusBarItem.menu = NSMenu()
            statusBarItem.menu?.autoenablesItems = false
            
            startItem.title = "Start"
            startItem.action = #selector(self.onStartClick(_:))
            statusBarItem.menu?.addItem(startItem)
            
            openItem.title = "Open"
            openItem.action = #selector(self.onOpenClick(_:))
            statusBarItem.menu?.addItem(openItem)
            
            stopItem.title = "Stop"
            stopItem.action = #selector(self.onStopClick(_:))
            statusBarItem.menu?.addItem(stopItem)
            
            logItem.title = "Logs"
            logItem.action = #selector(self.onLogsClick(_:))
            statusBarItem.menu?.addItem(logItem)
            
            let quitItem = NSMenuItem(title: "Quit", action: #selector(NSApplication.terminate(_:)), keyEquivalent: "q")
            statusBarItem.menu?.addItem(quitItem)
        }

        func applicationWillFinishLaunching(_ notification: Notification) {
            // another place to run stuff when the app starts
        }
        func windowShouldClose(_ sender: NSWindow) -> Bool {
            // can override closing (not quitting) the app
            return true
        }
        func applicationWillTerminate(_ notification: Notification) {
            // this is the only place that I have found to run stuff when the app quits
            StopApp()
        }
        @objc func onStartClick(_ sender: Any?) {
            do {
                try StartApp()
            } catch {
                
            }
        }
        @objc func onStopClick(_ sender: Any?) {
            StopApp()
        }
        @objc func onOpenClick(_ sender: Any?) {
            if let url = URL(string: "http://127.0.0.1:3080") {
                NSWorkspace.shared.open(url)
            }
        }
        @objc func onLogsClick(_ sender: Any?) {
            let logFileURL = FileManager.default.homeDirectoryForCurrentUser.appendingPathComponent("Library/Application Support/sourcegraph-sp/sourcegraph.log")
            if FileManager.default.fileExists(atPath: logFileURL.path) {
                NSWorkspace.shared.open(logFileURL)
            }
        }
    }
}
