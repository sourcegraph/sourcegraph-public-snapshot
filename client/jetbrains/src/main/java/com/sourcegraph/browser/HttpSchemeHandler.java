package com.sourcegraph.browser;

import com.google.common.collect.ImmutableMap;
import org.cef.callback.CefCallback;
import org.cef.handler.CefResourceHandlerAdapter;
import org.cef.misc.IntRef;
import org.cef.misc.StringRef;
import org.cef.network.CefRequest;
import org.cef.network.CefResponse;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.InputStream;
import java.util.Map;
import java.util.Optional;

public class HttpSchemeHandler extends CefResourceHandlerAdapter {
    private byte[] data;
    private String mimeType;
    private int responseHeader = 400;
    private int offset = 0;

    public synchronized boolean processRequest(CefRequest request, CefCallback callback) {
        String extension = getExtension(request.getURL());
        mimeType = getMimeType(extension);
        String url = request.getURL();
        String path = url.replace("http://sourcegraph", "");

        if (mimeType != null) {
            data = loadResource(path);
            responseHeader = data != null ? 200 : 404;
            if (data == null) {
                data = getDefaultContent(extension, path).getBytes();
            }
            callback.Continue();
            return true;
        } else {
            return false;
        }
    }

    public void getResponseHeaders(
        CefResponse response, IntRef responseLength, StringRef redirectUrl) {
        response.setMimeType(mimeType);
        response.setStatus(responseHeader);
        responseLength.set(data.length);
    }

    public synchronized boolean readResponse(
        byte[] dataOut, int bytesToRead, IntRef bytesRead, CefCallback callback) {
        boolean hasData = false;

        if (offset < data.length) {
            int transferSize = Math.min(bytesToRead, (data.length - offset));
            System.arraycopy(data, offset, dataOut, 0, transferSize);
            offset += transferSize;
            bytesRead.set(transferSize);
            hasData = true;
        } else {
            offset = 0;
            bytesRead.set(0);
        }

        return hasData;
    }

    private byte[] loadResource(String resourceName) {
        try (
            InputStream inStream = getClass().getResourceAsStream(resourceName)
        ) {
            if (inStream != null) {
                ByteArrayOutputStream outFile = new ByteArrayOutputStream();
                int readByte;
                while ((readByte = inStream.read()) >= 0) outFile.write(readByte);
                return outFile.toByteArray();
            }
        } catch (IOException e) {
            return null;
        }
        return null;
    }

    public String getExtension(String filename) {
        return Optional.ofNullable(filename)
            .filter(f -> f.contains("."))
            .map(f -> f.substring(filename.lastIndexOf(".") + 1)).orElse(null);
    }

    public String getDefaultContent(String extension, String path) {
        final Map<String, String> extensionToDefaultContent = ImmutableMap.of(
            "html", "<html><head><title>Error 404</title></head>" +
                "<body>" +
                "<h1>Error 404</h1>" +
                "File " + path + "  does not exist." +
                "</body></html>",
            "js", "",
            "css", ""
        );
        return extensionToDefaultContent.get(extension);
    }

    public String getMimeType(String extension) {
        final Map<String, String> extensionToMimeType = ImmutableMap.of(
            "html", "text/html",
            "js", "text/javascript",
            "css", "text/css"
        );
        return extensionToMimeType.get(extension);
    }
}
