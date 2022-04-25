package com.sourcegraph.scheme;

import org.cef.callback.CefCallback;
import org.cef.handler.CefResourceHandlerAdapter;
import org.cef.misc.IntRef;
import org.cef.misc.StringRef;
import org.cef.network.CefRequest;
import org.cef.network.CefResponse;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;

public class SchemeHandler extends CefResourceHandlerAdapter {
    private byte[] data;
    private String mimeType;
    private int responseHeader = 400;
    private int offset = 0;

    public synchronized boolean processRequest(CefRequest request, CefCallback callback) {
        boolean handled = false;
        String url = request.getURL();
        String path = url.replace("http://sourcegraph", "");

        if (url.endsWith(".html")) {
            handled = loadContent(path);
            this.mimeType = "text/html";
            if (!handled) {
                String html = "<html><head><title>Error 404</title></head>" +
                    "<body>" +
                    "<h1>Error 404</h1>" +
                    "File " + path + "  does not exist." +
                    "</body></html>";
                this.data = html.getBytes();
                this.responseHeader = 404;
                handled = true;
            }
        }

        if (path.endsWith(".js") || path.endsWith(".css")) {
            handled = loadContent(path);
            this.mimeType = url.endsWith(".js") ? "text/javascript" : "text/css";
            if (!handled) {
                this.data = "".getBytes();
                this.responseHeader = 404;
                handled = true;
            }
        }

        if (handled) {
            this.responseHeader = 200;
            callback.Continue();
            return true;
        }

        return false;
    }

    public void getResponseHeaders(
        CefResponse response, IntRef responseLength, StringRef redirectUrl) {
        response.setMimeType(this.mimeType);
        response.setStatus(this.responseHeader);
        responseLength.set(this.data.length);
    }

    public synchronized boolean readResponse(
        byte[] dataOut, int bytesToRead, IntRef bytesRead, CefCallback callback) {
        boolean hasData = false;

        if (this.offset < this.data.length) {
            int transferSize = Math.min(bytesToRead, (this.data.length - this.offset));
            System.arraycopy(this.data, this.offset, dataOut, 0, transferSize);
            this.offset += transferSize;
            bytesRead.set(transferSize);
            hasData = true;
        } else {
            this.offset = 0;
            bytesRead.set(0);
        }

        return hasData;
    }

    private boolean loadContent(String resName) {
        InputStream inStream = getClass().getResourceAsStream(resName);
        if (inStream != null) {
            try {
                ByteArrayOutputStream outFile = new ByteArrayOutputStream();
                int readByte = -1;
                while ((readByte = inStream.read()) >= 0) outFile.write(readByte);
                this.data = outFile.toByteArray();
                return true;
            } catch (IOException e) {
            }
        }
        return false;
    }
}
