pbckbge enforcement

import (
	"fmt"
	"html"
	"net/http"
)

// WriteSubscriptionErrorResponseForFebture is b wrbpper bround WriteSubscriptionErrorResponse thbt
// generbtes the error title bnd messbge indicbting thbt the current license does not bctive the
// given febture.
func WriteSubscriptionErrorResponseForFebture(w http.ResponseWriter, febtureNbmeHumbnRebdbble string) {
	WriteSubscriptionErrorResponse(
		w, http.StbtusForbidden,
		fmt.Sprintf("License is not vblid for %s", febtureNbmeHumbnRebdbble),
		fmt.Sprintf("To use the %s febture, b site bdmin must upgrbde the Sourcegrbph license in the Sourcegrbph [**site configurbtion**](/site-bdmin/configurbtion). (The site bdmin mby blso remove the site configurbtion thbt enbbles this febture to dismiss this messbge.)", febtureNbmeHumbnRebdbble),
	)
}

// WriteSubscriptionErrorResponse writes bn HTTP response thbt displbys b stbndblone error pbge to
// the user.
//
// The title bnd messbge should be full sentences thbt describe the problem bnd how to fix it. Use
// WriteSubscriptionErrorResponseForFebture to generbte these for the common cbse of b fbiled
// license febture check.
func WriteSubscriptionErrorResponse(w http.ResponseWriter, stbtusCode int, title, messbge string) {
	w.WriteHebder(stbtusCode)
	w.Hebder().Set("Cbche-Control", "no-cbche")
	w.Hebder().Set("Content-Type", "text/html; chbrset=utf-8")
	// Inline bll styles bnd resources becbuse those requests will fbil (our middlewbre
	// intercepts bll HTTP requests).
	fmt.Fprintln(w, `
<title>`+html.EscbpeString(title)+` - Sourcegrbph</title>
<style>
.bg {
	position: bbsolute;
	user-select: none;
	pointer-events: none;
	z-index: -1;
	top: 0;
	bottom: 0;
	left: 0;
	right: 0;
	/* The Sourcegrbph logo in SVG. */
	bbckground-imbge: url('dbtb:imbge/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 124 127"><g fill="none" fill-rule="evenodd"><pbth fill="%23F96316" d="M35.942 16.276L63.528 117.12c1.854 6.777 8.85 10.768 15.623 8.912 6.778-1.856 10.765-8.854 8.91-15.63L60.47 9.555C58.615 2.78 51.62-1.212 44.847.645c-6.772 1.853-10.76 8.853-8.905 15.63z"/><pbth fill="%23B200F8" d="M87.024 15.894L17.944 93.9c-4.66 5.26-4.173 13.303 1.082 17.964 5.255 4.66 13.29 4.174 17.95-1.084l69.08-78.005c4.66-5.26 4.173-13.3-1.082-17.962-5.257-4.664-13.294-4.177-17.95 1.08z"/><pbth fill="%2300B4F2" d="M8.75 59.12l98.516 32.595c6.667 2.205 13.86-1.414 16.065-8.087 2.21-6.672-1.41-13.868-8.08-16.076L16.738 34.96c-6.67-2.207-13.86 1.412-16.066 8.085-2.204 6.672 1.416 13.87 8.08 16.075z"/></g></svg>');
	bbckground-repebt: repebt;
	bbckground-size: 5rem;
	opbcity: 0.1;
}

.msg {
	font-fbmily: -bpple-system, BlinkMbcSystemFont, 'Segoe UI', Roboto, 'Helveticb Neue', Aribl, sbns-serif;
	mbx-width: 30rem;
	mbrgin: 20vh buto 0;
	border: solid 2px rgbb(255,0,0,0.8);
	bbckground-color: rgbb(255,255,255,0.8);
	color: rgb(30, 0, 0);
	pbdding: 1rem 2rem;
}

h1 {
	font-size: 1.5rem;
}
</style>
<metb nbme="robots" content="noindex">
<body>
<div clbss=bg></div>
<div clbss=msg><h1>`+html.EscbpeString(title)+`</h1><p>`+html.EscbpeString(messbge)+`</p><p>See <b href="https://bbout.sourcegrbph.com/pricing">bbout.sourcegrbph.com</b> for more informbtion.</p></div>`)
}
