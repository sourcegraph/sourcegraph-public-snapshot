/* Remote Selenium client implementation.

See http://code.google.com/p/selenium/wiki/JsonWireProtocol for wire protocol.
*/

package selenium

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
)

var Log = log.New(os.Stderr, "[selenium] ", log.Ltime|log.Lmicroseconds)
var Trace bool

/* Errors returned by Selenium server. */
var errorCodes = map[int]string{
	7:  "no such element",
	8:  "no such frame",
	9:  "unknown command",
	10: "stale element reference",
	11: "element not visible",
	12: "invalid element state",
	13: "unknown error",
	15: "element is not selectable",
	17: "javascript error",
	19: "xpath lookup error",
	21: "timeout",
	23: "no such window",
	24: "invalid cookie domain",
	25: "unable to set cookie",
	26: "unexpected alert open",
	27: "no alert open",
	28: "script timeout",
	29: "invalid element coordinates",
	32: "invalid selector",
}

const (
	SUCCESS         = 0
	defaultExecutor = "http://127.0.0.1:4444/wd/hub"
	jsonMIMEType    = "application/json"
)

type remoteWebDriver struct {
	id, executor string
	capabilities Capabilities
	// FIXME
	// profile             BrowserProfile
}

func (wd *remoteWebDriver) url(template string, args ...interface{}) string {
	path := fmt.Sprintf(template, args...)
	return wd.executor + path
}

func (wd *remoteWebDriver) send(method, url string, data []byte) (r *reply, err error) {
	var buf []byte
	if buf, err = wd.execute(method, url, data); err == nil {
		if len(buf) > 0 {
			err = json.Unmarshal(buf, &r)
		}
	}
	return
}

func (wd *remoteWebDriver) execute(method, url string, data []byte) ([]byte, error) {
	if Log != nil {
		Log.Printf("-> %s %s [%d bytes]", method, url, len(data))
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", jsonMIMEType)
	if method == "POST" {
		req.Header.Add("Content-Type", jsonMIMEType)
	}

	if Trace {
		if dump, err := httputil.DumpRequest(req, true); err == nil && Log != nil {
			Log.Printf("-> TRACE\n%s", dump)
		}
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if Trace {
		if dump, err := httputil.DumpResponse(res, true); err == nil && Log != nil {
			Log.Printf("<- TRACE\n%s", dump)
		}
	}

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if Log != nil {
		Log.Printf("<- %s (%s) [%d bytes]", res.Status, res.Header["Content-Type"], len(buf))
	}

	if res.StatusCode >= 400 {
		reply := new(reply)
		err := json.Unmarshal(buf, reply)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Bad server reply status: %s", res.Status))
		}
		message, ok := errorCodes[reply.Status]
		if !ok {
			message = fmt.Sprintf("unknown error - %d", reply.Status)
		}

		return nil, errors.New(message)
	}

	/* Some bug(?) in Selenium gets us nil values in output, json.Unmarshal is
	* not happy about that.
	 */
	if strings.HasPrefix(res.Header.Get("Content-Type"), jsonMIMEType) {
		reply := new(reply)
		err := json.Unmarshal(buf, reply)
		if err != nil {
			return nil, err
		}

		if reply.Status != SUCCESS {
			message, ok := errorCodes[reply.Status]
			if !ok {
				message = fmt.Sprintf("unknown error - %d", reply.Status)
			}

			return nil, errors.New(message)
		}
		return buf, err
	}

	// Nothing was returned, this is OK for some commands
	return buf, nil
}

var httpClient = http.Client{
	// WebDriver requires that all requests have an 'Accept: application/json' header. We must add
	// it here because by default net/http will not include that header when following redirects.
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		req.Header.Add("Accept", jsonMIMEType)
		if Trace {
			if dump, err := httputil.DumpRequest(req, true); err == nil && Log != nil {
				Log.Printf("-> TRACE (redirected request)\n%s", dump)
			}
		}
		return nil
	},
}

// Server reply to WebDriver command.
type reply struct {
	SessionId string
	Status    int
	Value     json.RawMessage
}

func (r *reply) readValue(v interface{}) error {
	return json.Unmarshal(r.Value, v)
}

// An active session.
type Session struct {
	Id           string
	Capabilities Capabilities
}

/* Create new remote client, this will also start a new session.
   capabilities - the desired capabilities, see http://goo.gl/SNlAk
   executor - the URL to the Selenim server
*/
func NewRemote(capabilities Capabilities, executor string) (WebDriver, error) {
	if executor == "" {
		executor = defaultExecutor
	}

	wd := &remoteWebDriver{executor: executor, capabilities: capabilities}
	// FIXME: Handle profile

	_, err := wd.NewSession()
	if err != nil {
		return nil, err
	}

	return wd, nil
}

func (wd *remoteWebDriver) stringCommand(urlTemplate string) (v string, err error) {
	var r *reply
	if r, err = wd.send("GET", wd.url(urlTemplate, wd.id), nil); err == nil {
		err = r.readValue(&v)
	}
	return
}

func (wd *remoteWebDriver) voidCommand(urlTemplate string, params interface{}) (err error) {
	var data []byte
	if params != nil {
		data, err = json.Marshal(params)
	}
	if err == nil {
		_, err = wd.send("POST", wd.url(urlTemplate, wd.id), data)
	}
	return

}

func (wd remoteWebDriver) stringsCommand(urlTemplate string) (v []string, err error) {
	var r *reply
	if r, err = wd.send("GET", wd.url(urlTemplate, wd.id), nil); err == nil {
		err = r.readValue(&v)
	}
	return
}

func (wd *remoteWebDriver) boolCommand(urlTemplate string) (v bool, err error) {
	var r *reply
	if r, err = wd.send("GET", wd.url(urlTemplate, wd.id), nil); err == nil {
		err = r.readValue(&v)
	}
	return
}

// WebDriver interface implementation

func (wd *remoteWebDriver) Status() (v *Status, err error) {
	var r *reply
	if r, err = wd.send("GET", wd.url("/status"), nil); err == nil {
		err = r.readValue(&v)
	}
	return
}

func (wd *remoteWebDriver) Sessions() (sessions []Session, err error) {
	var r *reply
	if r, err = wd.send("GET", wd.url("/sessions"), nil); err == nil {
		err = r.readValue(&sessions)
	}
	return
}

func (wd *remoteWebDriver) NewSession() (string, error) {
	message := map[string]interface{}{
		"desiredCapabilities": wd.capabilities,
	}

	var data []byte
	data, err := json.Marshal(message)
	if err != nil {
		return "", err
	}

	r, err := wd.send("POST", wd.url("/session"), data)
	if err != nil {
		return "", err
	}
	wd.id = r.SessionId

	return r.SessionId, nil
}

func (wd *remoteWebDriver) Capabilities() (v Capabilities, err error) {
	var r *reply
	if r, err = wd.send("GET", wd.url("/session/%s", wd.id), nil); err == nil {
		r.readValue(&v)
	}
	return
}

func (wd *remoteWebDriver) SetTimeout(timeoutType string, ms uint) error {
	params := map[string]interface{}{"type": timeoutType, "ms": ms}
	return wd.voidCommand("/session/%s/timeouts", params)
}

func (wd *remoteWebDriver) SetAsyncScriptTimeout(ms uint) error {
	params := map[string]uint{"ms": ms}
	return wd.voidCommand("/session/%s/timeouts/async_script", params)
}

func (wd *remoteWebDriver) SetImplicitWaitTimeout(ms uint) error {
	params := map[string]uint{"ms": ms}
	return wd.voidCommand("/session/%s/timeouts/implicit_wait", params)
}

func (wd *remoteWebDriver) AvailableEngines() ([]string, error) {
	return wd.stringsCommand("/session/%s/ime/available_engines")
}

func (wd *remoteWebDriver) ActiveEngine() (string, error) {
	return wd.stringCommand("/session/%s/ime/active_engine")
}

func (wd *remoteWebDriver) IsEngineActivated() (bool, error) {
	return wd.boolCommand("/session/%s/ime/activated")
}

func (wd *remoteWebDriver) DeactivateEngine() error {
	return wd.voidCommand("session/%s/ime/deactivate", nil)
}

func (wd *remoteWebDriver) ActivateEngine(engine string) (err error) {
	return wd.voidCommand("/session/%s/ime/activate", map[string]string{"engine": engine})
}

func (wd *remoteWebDriver) Quit() (err error) {
	if _, err = wd.execute("DELETE", wd.url("/session/%s", wd.id), nil); err == nil {
		wd.id = ""
	}
	return
}

func (wd *remoteWebDriver) CurrentWindowHandle() (string, error) {
	return wd.stringCommand("/session/%s/window_handle")
}

func (wd *remoteWebDriver) WindowHandles() ([]string, error) {
	return wd.stringsCommand("/session/%s/window_handles")
}

func (wd *remoteWebDriver) CurrentURL() (string, error) {
	return wd.stringCommand("/session/%s/url")
}

func (wd *remoteWebDriver) Get(url string) error {
	return wd.voidCommand("/session/%s/url", map[string]string{"url": url})
}

func (wd *remoteWebDriver) Forward() error {
	return wd.voidCommand("/session/%s/forward", nil)
}

func (wd *remoteWebDriver) Back() error {
	return wd.voidCommand("/session/%s/back", nil)
}

func (wd *remoteWebDriver) Refresh() error {
	return wd.voidCommand("/session/%s/refresh", nil)
}

func (wd *remoteWebDriver) Title() (string, error) {
	return wd.stringCommand("/session/%s/title")
}

func (wd *remoteWebDriver) PageSource() (string, error) {
	return wd.stringCommand("/session/%s/source")
}

type element struct {
	Element string `json:"ELEMENT"`
}

func (wd *remoteWebDriver) find(by, value, suffix, url string) (r *reply, err error) {
	params := map[string]string{"using": by, "value": value}
	var data []byte
	if data, err = json.Marshal(params); err == nil {
		if url == "" {
			url = "/session/%s/element"
		}
		urlTemplate := url + suffix
		url = wd.url(urlTemplate, wd.id)
		r, err = wd.send("POST", url, data)
	}
	return
}

func decodeElement(wd *remoteWebDriver, r *reply) WebElement {
	var elem element
	if err := r.readValue(&elem); err != nil {
		panic(err.Error() + ": " + string(r.Value))
	}
	return &remoteWE{parent: wd, id: elem.Element}
}

func (wd *remoteWebDriver) FindElement(by, value string) (WebElement, error) {
	if res, err := wd.find(by, value, "", ""); err == nil {
		return decodeElement(wd, res), nil
	} else {
		return nil, err
	}
}

func decodeElements(wd *remoteWebDriver, r *reply) (welems []WebElement) {
	var elems []element
	if err := r.readValue(&elems); err != nil {
		panic(err.Error() + ": " + string(r.Value))
	}
	for _, elem := range elems {
		welems = append(welems, &remoteWE{wd, elem.Element})
	}
	return
}

func (wd *remoteWebDriver) FindElements(by, value string) ([]WebElement, error) {
	if res, err := wd.find(by, value, "s", ""); err == nil {
		return decodeElements(wd, res), nil
	} else {
		return nil, err
	}
}

func (wd *remoteWebDriver) Q(sel string) (WebElement, error) {
	return wd.FindElement(ByCSSSelector, sel)
}

func (wd *remoteWebDriver) QAll(sel string) ([]WebElement, error) {
	return wd.FindElements(ByCSSSelector, sel)
}

func (wd *remoteWebDriver) Close() error {
	_, err := wd.execute("DELETE", wd.url("/session/%s/window", wd.id), nil)
	return err
}

func (wd *remoteWebDriver) SwitchWindow(name string) error {
	params := map[string]string{"name": name}
	return wd.voidCommand("/session/%s/window", params)
}

func (wd *remoteWebDriver) CloseWindow(name string) error {
	_, err := wd.execute("DELETE", wd.url("/session/%s/window", wd.id), nil)
	return err
}

func (wd *remoteWebDriver) WindowSize(name string) (sz *Size, err error) {
	url := wd.url("/session/%s/window/%s/size", wd.id, name)
	var r *reply
	if r, err = wd.send("GET", url, nil); err == nil {
		err = r.readValue(&sz)
	}
	return
}

func (wd *remoteWebDriver) WindowPosition(name string) (pt *Point, err error) {
	url := wd.url("/session/%s/window/%s/position", wd.id, name)
	var r *reply
	if r, err = wd.send("GET", url, nil); err == nil {
		err = r.readValue(&pt)
	}
	return
}

func (wd *remoteWebDriver) ResizeWindow(name string, to Size) error {
	url := wd.url("/session/%s/window/%s/size", wd.id, name)
	data, err := json.Marshal(to)
	if err != nil {
		return err
	}
	_, err = wd.send("POST", url, data)
	return err
}

func (wd *remoteWebDriver) SwitchFrame(frame string) error {
	params := map[string]string{"id": frame}
	return wd.voidCommand("/session/%s/frame", params)
}

func (wd *remoteWebDriver) SwitchFrameParent() error {
	return wd.voidCommand("/session/%s/frame/parent", nil)
}

func (wd *remoteWebDriver) ActiveElement() (WebElement, error) {
	url := wd.url("/session/%s/element/active", wd.id)
	if r, err := wd.send("GET", url, nil); err == nil {
		return decodeElement(wd, r), nil
	} else {
		return nil, err
	}
}

func (wd *remoteWebDriver) GetCookies() (c []Cookie, err error) {
	var r *reply
	if r, err = wd.send("GET", wd.url("/session/%s/cookie", wd.id), nil); err == nil {
		err = r.readValue(&c)
		if err == nil {
			parseCookieExpiry(&c, r.Value)
		}
	}
	return
}

func parseCookieExpiry(cookies *[]Cookie, raw json.RawMessage) {
	var expiries []struct {
		Expiry json.Number
	}

	err := json.Unmarshal(raw, &expiries)
	if err != nil {
		return
	}

	for i, _ := range *cookies {
		expiry, err := expiries[i].Expiry.Float64()
		if err != nil {
			continue
		}

		(*cookies)[i].Expiry = uint(expiry)
	}
}

func (wd *remoteWebDriver) AddCookie(cookie *Cookie) error {
	params := map[string]*Cookie{"cookie": cookie}
	return wd.voidCommand("/session/%s/cookie", params)
}

func (wd *remoteWebDriver) DeleteAllCookies() error {
	_, err := wd.execute("DELETE", wd.url("/session/%s/cookie", wd.id), nil)
	return err
}

func (wd *remoteWebDriver) DeleteCookie(name string) error {
	_, err := wd.execute("DELETE", wd.url("/session/%s/cookie/%s", wd.id, name), nil)
	return err
}

func (wd *remoteWebDriver) Click(button int) error {
	params := map[string]int{"button": button}
	return wd.voidCommand("/session/%s/click", params)
}

func (wd *remoteWebDriver) DoubleClick() error {
	return wd.voidCommand("/session/%s/doubleclick", nil)
}

func (wd *remoteWebDriver) ButtonDown() error {
	return wd.voidCommand("/session/%s/buttondown", nil)
}

func (wd *remoteWebDriver) ButtonUp() error {
	return wd.voidCommand("/session/%s/buttonup", nil)
}

func (wd *remoteWebDriver) SendModifier(modifier string, isDown bool) error {
	params := map[string]interface{}{
		"value":  modifier,
		"isdown": isDown,
	}

	data, err := json.Marshal(params)
	if err != nil {
		return err
	}

	return wd.voidCommand("/session/%s/modifier", data)
}

func (wd *remoteWebDriver) DismissAlert() error {
	return wd.voidCommand("/session/%s/dismiss_alert", nil)
}

func (wd *remoteWebDriver) AcceptAlert() error {
	return wd.voidCommand("/session/%s/accept_alert", nil)
}

func (wd *remoteWebDriver) AlertText() (string, error) {
	return wd.stringCommand("/session/%s/alert_text")
}

func (wd *remoteWebDriver) SetAlertText(text string) error {
	params := map[string]string{"text": text}
	return wd.voidCommand("/session/%s/alert_text", params)
}

func (wd *remoteWebDriver) execScript(script string, args []interface{}, suffix string) (res interface{}, err error) {
	if args == nil {
		args = []interface{}{}
	}
	params := map[string]interface{}{
		"script": script,
		"args":   args,
	}
	var data []byte
	if data, err = json.Marshal(params); err != nil {
		return nil, err
	}
	url := wd.url("/session/%s/execute"+suffix, wd.id)
	var r *reply
	if r, err = wd.send("POST", url, data); err == nil {
		err = r.readValue(&res)
	}
	return
}

func (wd *remoteWebDriver) ExecuteScript(script string, args []interface{}) (interface{}, error) {
	return wd.execScript(script, args, "")
}

func (wd *remoteWebDriver) ExecuteScriptAsync(script string, args []interface{}) (interface{}, error) {
	return wd.execScript(script, args, "_async")
}

func (wd *remoteWebDriver) Screenshot() ([]byte, error) {
	data, err := wd.stringCommand("/session/%s/screenshot")
	if err != nil {
		return nil, err
	}

	// Selenium returns base64 encoded image
	buf := []byte(data)
	decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewBuffer(buf))
	return ioutil.ReadAll(decoder)
}

func (wd *remoteWebDriver) T(t TestingT) WebDriverT {
	return &webDriverT{wd, t}
}

// WebElement interface implementation

type remoteWE struct {
	parent *remoteWebDriver
	id     string
}

func (elem *remoteWE) Click() error {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/click", elem.id)
	return elem.parent.voidCommand(urlTemplate, nil)
}

func (elem *remoteWE) SendKeys(keys string) error {
	chars := make([]string, len(keys))
	for i, c := range keys {
		chars[i] = string(c)
	}
	params := map[string][]string{"value": chars}
	urltmpl := fmt.Sprintf("/session/%%s/element/%s/value", elem.id)
	return elem.parent.voidCommand(urltmpl, params)
}

func (elem *remoteWE) TagName() (string, error) {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/name", elem.id)
	return elem.parent.stringCommand(urlTemplate)
}

func (elem *remoteWE) Text() (string, error) {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/text", elem.id)
	return elem.parent.stringCommand(urlTemplate)
}

func (elem *remoteWE) Submit() error {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/submit", elem.id)
	return elem.parent.voidCommand(urlTemplate, nil)
}

func (elem *remoteWE) Clear() error {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/clear", elem.id)
	return elem.parent.voidCommand(urlTemplate, nil)
}

func (elem *remoteWE) MoveTo(xOffset, yOffset int) error {
	params := map[string]interface{}{
		"element": elem.id,
		"xoffset": xOffset,
		"yoffset": yOffset,
	}
	return elem.parent.voidCommand("/session/%s/moveto", params)
}

func (elem *remoteWE) FindElement(by, value string) (WebElement, error) {
	res, err := elem.parent.find(by, value, "", fmt.Sprintf("/session/%%s/element/%s/element", elem.id))
	if err != nil {
		return nil, err
	}
	return decodeElement(elem.parent, res), nil
}

func (elem *remoteWE) Q(sel string) (WebElement, error) {
	return elem.FindElement(ByCSSSelector, sel)
}

func (elem *remoteWE) QAll(sel string) ([]WebElement, error) {
	return elem.FindElements(ByCSSSelector, sel)
}

func (elem *remoteWE) FindElements(by, value string) ([]WebElement, error) {
	res, err := elem.parent.find(by, value, "s", fmt.Sprintf("/session/%%s/element/%s/element", elem.id))
	if err != nil {
		return nil, err
	}
	return decodeElements(elem.parent, res), nil
}

func (elem *remoteWE) boolQuery(urlTemplate string) (bool, error) {
	url := fmt.Sprintf(urlTemplate, elem.id)
	return elem.parent.boolCommand(url)
}

// Porperties
func (elem *remoteWE) IsSelected() (bool, error) {
	return elem.boolQuery("/session/%%s/element/%s/selected")
}

func (elem *remoteWE) IsEnabled() (bool, error) {
	return elem.boolQuery("/session/%%s/element/%s/enabled")
}

func (elem *remoteWE) IsDisplayed() (bool, error) {
	return elem.boolQuery("/session/%%s/element/%s/displayed")
}

func (elem *remoteWE) GetAttribute(name string) (string, error) {
	template := "/session/%%s/element/%s/attribute/%s"
	urlTemplate := fmt.Sprintf(template, elem.id, name)

	return elem.parent.stringCommand(urlTemplate)
}

func (elem *remoteWE) location(suffix string) (pt *Point, err error) {
	wd := elem.parent
	path := "/session/%s/element/%s/location" + suffix
	url := wd.url(path, wd.id, elem.id)
	var r *reply
	if r, err = wd.send("GET", url, nil); err == nil {
		err = r.readValue(&pt)
	}
	return
}

func (elem *remoteWE) Location() (*Point, error) {
	return elem.location("")
}

func (elem *remoteWE) LocationInView() (*Point, error) {
	return elem.location("_in_view")
}

func (elem *remoteWE) Size() (sz *Size, err error) {
	wd := elem.parent
	url := wd.url("/session/%s/element/%s/size", wd.id, elem.id)
	var r *reply
	if r, err = wd.send("GET", url, nil); err == nil {
		err = r.readValue(&sz)
	}
	return
}

func (elem *remoteWE) CSSProperty(name string) (string, error) {
	urlTemplate := fmt.Sprintf("/session/%%s/element/%s/css/%s", elem.id, name)
	return elem.parent.stringCommand(urlTemplate)
}

func (elem *remoteWE) T(t TestingT) WebElementT {
	return &webElementT{elem, t}
}
